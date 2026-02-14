package grpc

import (
	"context"
	"testing"
	"time"

	"docksphinx/internal/monitor"
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
