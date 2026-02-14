package grpc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/event"
	"docksphinx/internal/monitor"
	"google.golang.org/grpc/metadata"
)

func TestServerSnapshotAndStreamInitial(t *testing.T) {
	engine, err := monitor.NewEngine(monitor.EngineConfig{
		Interval:         time.Second,
		ResourceInterval: 5 * time.Second,
		Thresholds:       monitor.DefaultThresholdConfig(),
	}, nil)
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}

	engine.GetStateManager().UpdateState("c1", &monitor.ContainerState{
		ContainerID:   "c1",
		ContainerName: "web",
		ImageName:     "nginx:latest",
		State:         "running",
		Status:        "Up",
		LastSeen:      time.Now(),
	})

	server, err := NewServer(&ServerOptions{
		Address:          "127.0.0.1:0",
		EnableReflection: false,
		RecentEventLimit: 10,
	}, engine)
	if err != nil {
		t.Fatalf("new server failed: %v", err)
	}
	defer server.Stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()
	defer func() {
		select {
		case <-errCh:
		case <-time.After(100 * time.Millisecond):
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, server.Address())
	if err != nil {
		t.Fatalf("new client failed: %v", err)
	}
	defer client.Close()

	snapshot, err := client.GetSnapshot(ctx)
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}
	if len(snapshot.GetContainers()) != 1 {
		t.Fatalf("expected 1 container, got %d", len(snapshot.GetContainers()))
	}

	stream, err := client.Stream(ctx, true)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}
	update, err := stream.Recv()
	if err != nil {
		t.Fatalf("stream recv failed: %v", err)
	}
	if update.GetSnapshot() == nil {
		t.Fatalf("expected initial snapshot payload, got %#v", update.GetPayload())
	}
}

func TestServerStreamReturnsInitialSnapshotSendError(t *testing.T) {
	engine, err := monitor.NewEngine(monitor.EngineConfig{
		Interval:         time.Second,
		ResourceInterval: 5 * time.Second,
		Thresholds:       monitor.DefaultThresholdConfig(),
	}, nil)
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}

	wantErr := errors.New("send failed")
	stream := &stubStreamServer{
		ctx:       context.Background(),
		sendErrAt: 1,
		sendErr:   wantErr,
	}

	srv := &Server{
		engine: engine,
		opts: &ServerOptions{
			RecentEventLimit: 10,
		},
		bcast: NewBroadcaster(),
	}

	err = srv.Stream(&pb.StreamRequest{IncludeInitialSnapshot: true}, stream)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected initial send error %v, got %v", wantErr, err)
	}
	if got := stream.SendCount(); got != 1 {
		t.Fatalf("expected one send attempt, got %d", got)
	}
}

func TestServerStreamSkipsNilEventPayloads(t *testing.T) {
	engine, err := monitor.NewEngine(monitor.EngineConfig{
		Interval:         time.Second,
		ResourceInterval: 5 * time.Second,
		Thresholds:       monitor.DefaultThresholdConfig(),
	}, nil)
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &stubStreamServer{ctx: ctx}
	srv := &Server{
		engine: engine,
		opts: &ServerOptions{
			RecentEventLimit: 10,
		},
		bcast: NewBroadcaster(),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Stream(&pb.StreamRequest{IncludeInitialSnapshot: false}, stream)
	}()

	var sub chan *event.Event
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		srv.bcast.mu.RLock()
		for ch := range srv.bcast.subscribers {
			sub = ch
			break
		}
		srv.bcast.mu.RUnlock()
		if sub != nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if sub == nil {
		t.Fatal("subscriber was not registered in time")
	}

	sub <- nil
	sub <- &event.Event{
		ID:            "ev-1",
		Type:          "started",
		ContainerID:   "cid-1",
		ContainerName: "web",
		Message:       "ok",
		Timestamp:     time.Unix(1700000000, 0),
	}

	waitDeadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(waitDeadline) {
		if stream.SendCount() >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := stream.SendCount(); got != 1 {
		t.Fatalf("expected exactly one non-nil event send, got %d", got)
	}
	sent := stream.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected exactly one captured update, got %d", len(sent))
	}
	if got := sent[0].GetEvent().GetId(); got != "ev-1" {
		t.Fatalf("expected sent event id ev-1, got %q", got)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error on context cancel shutdown, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("stream did not stop in time after cancel")
	}
}

type stubStreamServer struct {
	ctx       context.Context
	sendErrAt int
	sendErr   error
	mu        sync.Mutex
	sendCount int
	sent      []*pb.StreamUpdate
}

func (s *stubStreamServer) Send(update *pb.StreamUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sendCount++
	if s.sendErrAt > 0 && s.sendCount == s.sendErrAt {
		return s.sendErr
	}
	s.sent = append(s.sent, update)
	return nil
}

func (s *stubStreamServer) SendCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sendCount
}

func (s *stubStreamServer) Sent() []*pb.StreamUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*pb.StreamUpdate(nil), s.sent...)
}

func (s *stubStreamServer) SetHeader(metadata.MD) error  { return nil }
func (s *stubStreamServer) SendHeader(metadata.MD) error { return nil }
func (s *stubStreamServer) SetTrailer(metadata.MD)       {}
func (s *stubStreamServer) Context() context.Context     { return s.ctx }
func (s *stubStreamServer) SendMsg(any) error            { return nil }
func (s *stubStreamServer) RecvMsg(any) error            { return nil }
