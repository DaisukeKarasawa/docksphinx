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
	bcastCancel context.CancelFunc
	mu          sync.Mutex
}

// ServerOptions configures the gRPC server
type ServerOptions struct {
	Address          string // e.g. "127.0.0.1:50051"
	EnableReflection bool
	RecentEventLimit int
}

// NewServer creates a new gRPC server (does not start listening).
// Engine must already be started; its event channel is fanned out via Broadcaster.
func NewServer(opts *ServerOptions, engine *monitor.Engine) (*Server, error) {
	if engine == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	if opts == nil {
		opts = &ServerOptions{Address: "127.0.0.1:50051", RecentEventLimit: 50}
	}
	if opts.RecentEventLimit <= 0 {
		opts.RecentEventLimit = 50
	}
	lis, err := net.Listen("tcp", opts.Address)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", opts.Address, err)
	}
	s := grpc.NewServer()
	bcast := NewBroadcaster()
	bcastCtx, bcastCancel := context.WithCancel(context.Background())
	srv := &Server{lis: lis, grpc: s, opts: opts, engine: engine, bcast: bcast, bcastCancel: bcastCancel}
	pb.RegisterDocksphinxServiceServer(s, srv)
	if opts.EnableReflection {
		reflection.Register(s)
	}
	go bcast.Run(bcastCtx, engine.GetEventChannel())
	return srv, nil
}

// Start starts the gRPC server (blocking). Call from a goroutine.
func (s *Server) Start() error {
	return s.grpc.Serve(s.lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grpc != nil {
		if s.bcastCancel != nil {
			s.bcastCancel()
		}
		s.grpc.GracefulStop()
		s.grpc = nil
	}
}

// Address returns listening address.
func (s *Server) Address() string {
	if s == nil || s.lis == nil {
		return ""
	}
	return s.lis.Addr().String()
}

// GetSnapshot implements DocksphinxService
func (s *Server) GetSnapshot(ctx context.Context, req *pb.GetSnapshotRequest) (*pb.Snapshot, error) {
	sm := s.engine.GetStateManager()
	if sm == nil {
		return nil, status.Error(codes.Unavailable, "state not available")
	}
	snapshot := StateToSnapshot(sm)
	snapshot.RecentEvents = EventsToProto(s.engine.GetRecentEvents(s.opts.RecentEventLimit))
	return snapshot, nil
}

// Stream implements DocksphinxService
func (s *Server) Stream(req *pb.StreamRequest, stream pb.DocksphinxService_StreamServer) error {
	if req != nil && req.IncludeInitialSnapshot {
		sm := s.engine.GetStateManager()
		if sm != nil {
			snapshot := StateToSnapshot(sm)
			snapshot.RecentEvents = EventsToProto(s.engine.GetRecentEvents(s.opts.RecentEventLimit))
			if err := stream.Send(&pb.StreamUpdate{Payload: &pb.StreamUpdate_Snapshot{Snapshot: snapshot}}); err != nil {
				return err
			}
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
			protoEv := EventToProto(ev)
			if protoEv == nil {
				continue
			}
			if err := stream.Send(&pb.StreamUpdate{Payload: &pb.StreamUpdate_Event{Event: protoEv}}); err != nil {
				return err
			}
		}
	}
}
