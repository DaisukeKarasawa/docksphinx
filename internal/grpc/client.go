package grpc

import (
	"context"
	"fmt"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultDialTimeout = 5 * time.Second

type Client struct {
	conn   *ggrpc.ClientConn
	client pb.DocksphinxServiceClient
}

func NewClient(ctx context.Context, address string) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()

	conn, err := ggrpc.NewClient(
		address,
		ggrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}
	if err := waitUntilReady(dialCtx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("wait for grpc readiness %s: %w", address, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewDocksphinxServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) GetSnapshot(ctx context.Context) (*pb.Snapshot, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return c.client.GetSnapshot(ctx, &pb.GetSnapshotRequest{})
}

func (c *Client) Stream(ctx context.Context, includeInitialSnapshot bool) (pb.DocksphinxService_StreamClient, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return c.client.Stream(ctx, &pb.StreamRequest{
		IncludeInitialSnapshot: includeInitialSnapshot,
	})
}

func waitUntilReady(ctx context.Context, conn *ggrpc.ClientConn) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if !conn.WaitForStateChange(ctx, state) {
			if err := ctx.Err(); err != nil {
				return err
			}
			return fmt.Errorf("connection state unchanged: %s", state.String())
		}
	}
}
