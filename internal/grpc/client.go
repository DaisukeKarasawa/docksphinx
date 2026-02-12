package grpc

import (
	"context"
	"fmt"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultDialTimeout = 5 * time.Second

// Client wraps the DocksphinxService gRPC client.
type Client struct {
	conn   *ggrpc.ClientConn
	client pb.DocksphinxServiceClient
}

// NewClient connects to the daemon at address (e.g. "127.0.0.1:50051").
func NewClient(ctx context.Context, address string) (*Client, error) {
	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()

	conn, err := ggrpc.DialContext(
		dialCtx,
		address,
		ggrpc.WithTransportCredentials(insecure.NewCredentials()),
		ggrpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewDocksphinxServiceClient(conn),
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// GetSnapshot calls the GetSnapshot RPC.
func (c *Client) GetSnapshot(ctx context.Context) (*pb.Snapshot, error) {
	return c.client.GetSnapshot(ctx, &pb.GetSnapshotRequest{})
}

// Stream calls the Stream RPC and returns a stream client.
func (c *Client) Stream(ctx context.Context, includeInitialSnapshot bool) (pb.DocksphinxService_StreamClient, error) {
	return c.client.Stream(ctx, &pb.StreamRequest{
		IncludeInitialSnapshot: includeInitialSnapshot,
	})
}
