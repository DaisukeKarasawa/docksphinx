package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/monitor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server is the gRPC server for DocksphinxService
type Server struct {
	pb.UnimplementedDocksphinxServiceServer
	lis         net.Listener
	grpc        *grpc.Server
	opts        *ServerOptions
	engine      *monitor.Engine
	bcast       *Broadcaster
	cancelBcast context.CancelFunc
	mu          sync.Mutex
}

// ServerOptions configures the gRPC server
type ServerOptions struct {
	Address string // e.g. "127.0.0.1:50051"
}

// NewServer creates a new gRPC server (does not start listening).
// Engine must already be started; its event channel is fanned out via Broadcaster.
func NewServer(opts *ServerOptions, engine *monitor.Engine) (*Server, error) {
	if opts == nil {
		opts = &ServerOptions{Address: "127.0.0.1:50051"}
	}
	lis, err := net.Listen("tcp", opts.Address)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", opts.Address, err)
	}
	s := grpc.NewServer()
	bcast := NewBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{lis: lis, grpc: s, opts: opts, engine: engine, bcast: bcast, cancelBcast: cancel}
	pb.RegisterDocksphinxServiceServer(s, srv)
	reflection.Register(s)
	go bcast.Run(ctx, engine.GetEventChannel())
	return srv, nil
}

// Start starts the gRPC server (blocking). Call from a goroutine.
func (s *Server) Start() error {
	return s.grpc.Serve(s.lis)
}

// Stop gracefully stops the gRPC server and the broadcaster goroutine
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelBcast != nil {
		s.cancelBcast()
		s.cancelBcast = nil
	}
	if s.grpc != nil {
		s.grpc.GracefulStop()
		s.grpc = nil
	}
}

// GetSnapshot implements DocksphinxService
func (s *Server) GetSnapshot(ctx context.Context, req *pb.GetSnapshotRequest) (*pb.Snapshot, error) {
	sm := s.engine.GetStateManager()
	if sm == nil {
		return nil, status.Error(codes.Unavailable, "state not available")
	}
	return StateToSnapshot(sm), nil
}

// Stream implements DocksphinxService
func (s *Server) Stream(req *pb.StreamRequest, stream pb.DocksphinxService_StreamServer) error {
	if req != nil && req.IncludeInitialSnapshot {
		sm := s.engine.GetStateManager()
		if sm != nil {
			_ = stream.Send(&pb.StreamUpdate{Payload: &pb.StreamUpdate_Snapshot{Snapshot: StateToSnapshot(sm)}})
		}
	}
	sub, unsub := s.bcast.Subscribe()
	defer unsub()
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev, ok := <-sub:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.StreamUpdate{Payload: &pb.StreamUpdate_Event{Event: EventToProto(ev)}}); err != nil {
				return err
			}
		}
	}
}
