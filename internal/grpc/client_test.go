package grpc

import (
	"context"
	"testing"

	pb "docksphinx/api/docksphinx/v1"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubClient struct {
	lastSnapshotCtx context.Context
	lastSnapshotReq *pb.GetSnapshotRequest
	lastStreamCtx   context.Context
	lastStreamReq   *pb.StreamRequest
	snapshotCalls   int
	streamCalls     int
}

type testCtxKey string

func (s *stubClient) GetSnapshot(ctx context.Context, in *pb.GetSnapshotRequest, opts ...ggrpc.CallOption) (*pb.Snapshot, error) {
	s.lastSnapshotCtx = ctx
	s.lastSnapshotReq = in
	s.snapshotCalls++
	return &pb.Snapshot{}, nil
}

func (s *stubClient) Stream(ctx context.Context, in *pb.StreamRequest, opts ...ggrpc.CallOption) (ggrpc.ServerStreamingClient[pb.StreamUpdate], error) {
	s.lastStreamCtx = ctx
	s.lastStreamReq = in
	s.streamCalls++
	return nil, nil
}

func TestClientGetSnapshotAndStreamForwardContextAndRequests(t *testing.T) {
	stub := &stubClient{}
	c := &Client{client: stub}
	ctx := context.WithValue(context.Background(), testCtxKey("k"), "v")

	if _, err := c.GetSnapshot(ctx); err != nil {
		t.Fatalf("GetSnapshot returned unexpected error: %v", err)
	}
	if stub.lastSnapshotCtx != ctx {
		t.Fatal("expected GetSnapshot to forward caller context")
	}
	if stub.lastSnapshotReq == nil {
		t.Fatal("expected GetSnapshot request to be non-nil")
	}

	if _, err := c.Stream(ctx, true); err != nil {
		t.Fatalf("Stream returned unexpected error: %v", err)
	}
	if stub.lastStreamCtx != ctx {
		t.Fatal("expected Stream to forward caller context")
	}
	if stub.lastStreamReq == nil || !stub.lastStreamReq.GetIncludeInitialSnapshot() {
		t.Fatalf("expected Stream request include_initial_snapshot=true, got %#v", stub.lastStreamReq)
	}
}

func TestClientMethodsReturnErrorWhenClientIsNil(t *testing.T) {
	var c *Client

	if _, err := c.GetSnapshot(context.Background()); err == nil {
		t.Fatal("expected GetSnapshot to fail for nil client")
	}
	if _, err := c.Stream(context.Background(), false); err == nil {
		t.Fatal("expected Stream to fail for nil client")
	}
	if err := c.Close(); err != nil {
		t.Fatalf("expected Close on nil client to be no-op, got %v", err)
	}
}

func TestWaitUntilReadyRejectsNilConnection(t *testing.T) {
	if err := waitUntilReady(context.Background(), nil); err == nil {
		t.Fatal("expected waitUntilReady(nil conn) to fail")
	}
}

func TestNewClientRejectsEmptyAddress(t *testing.T) {
	_, err := NewClient(context.Background(), "")
	if err == nil {
		t.Fatal("expected NewClient to fail for empty address")
	}
	if got := err.Error(); got != "address cannot be empty" {
		t.Fatalf("expected empty address error, got %q", got)
	}
}

func TestNewClientRejectsWhitespaceAddress(t *testing.T) {
	_, err := NewClient(context.Background(), "   \t  ")
	if err == nil {
		t.Fatal("expected NewClient to fail for whitespace-only address")
	}
	if got := err.Error(); got != "address cannot be empty" {
		t.Fatalf("expected empty address error for whitespace input, got %q", got)
	}
}

func TestClientMethodsReturnContextErrorBeforeRPCWhenCanceled(t *testing.T) {
	stub := &stubClient{}
	c := &Client{client: stub}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.GetSnapshot(ctx); status.Code(err) != codes.Canceled {
		t.Fatalf("expected canceled status from GetSnapshot, got %v", err)
	}
	if stub.snapshotCalls != 0 {
		t.Fatalf("expected no downstream GetSnapshot call when context canceled, got %d", stub.snapshotCalls)
	}

	if _, err := c.Stream(ctx, true); status.Code(err) != codes.Canceled {
		t.Fatalf("expected canceled status from Stream, got %v", err)
	}
	if stub.streamCalls != 0 {
		t.Fatalf("expected no downstream Stream call when context canceled, got %d", stub.streamCalls)
	}
}
