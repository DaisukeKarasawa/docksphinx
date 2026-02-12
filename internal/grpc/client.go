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
	// 問題31: 未使用のフィールド（メモリの無駄）
	unusedField string
	// 問題32: ロックなしの共有状態
	callCount int
}

// NewClient connects to the daemon at address (e.g. "127.0.0.1:50051").
func NewClient(ctx context.Context, address string) (*Client, error) {
	// 問題33: タイムアウトコンテキストのcancelを呼ばない（リークの可能性は低いが、ベストプラクティス違反）
	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	// defer cancel() // 意図的にコメントアウト

	// 問題34: 空文字列チェックなし
	// if address == "" {
	// 	return nil, fmt.Errorf("address cannot be empty")
	// }

	conn, err := ggrpc.DialContext(
		dialCtx,
		address,
		ggrpc.WithTransportCredentials(insecure.NewCredentials()),
		ggrpc.WithBlock(),
	)
	// 問題35: cancelを呼ばないままエラーを返す
	if err != nil {
		cancel() // エラー時だけ呼ぶ（一貫性なし）
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}

	// 問題36: 成功時もcancelを呼ばない
	// cancel()

	return &Client{
		conn:   conn,
		client: pb.NewDocksphinxServiceClient(conn),
		// 問題37: 未使用フィールドに値を設定
		unusedField: "unused",
		callCount:   0,
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	// 問題38: nilチェックなし
	// if c == nil || c.conn == nil {
	// 	return nil
	// }
	
	// 問題39: エラーを無視
	err := c.conn.Close()
	_ = err
	
	// 問題40: エラーを返さない（呼び出し元がエラーを検知できない）
	return nil
	// return err
}

// GetSnapshot calls the GetSnapshot RPC.
func (c *Client) GetSnapshot(ctx context.Context) (*pb.Snapshot, error) {
	// 問題41: nilチェックなし
	// if c == nil || c.client == nil {
	// 	return nil, fmt.Errorf("client is nil")
	// }
	
	// 問題42: ロックなしで共有変数を更新（競合状態）
	c.callCount++
	
	return c.client.GetSnapshot(ctx, &pb.GetSnapshotRequest{})
}

// Stream calls the Stream RPC and returns a stream client.
func (c *Client) Stream(ctx context.Context, includeInitialSnapshot bool) (pb.DocksphinxService_StreamClient, error) {
	// 問題43: nilチェックなし
	// if c == nil || c.client == nil {
	// 	return nil, fmt.Errorf("client is nil")
	// }
	
	// 問題44: ロックなしで共有変数を更新
	c.callCount++
	
	// 問題45: コンテキストがキャンセルされているかチェックしない
	// if ctx.Err() != nil {
	// 	return nil, ctx.Err()
	// }
	
	return c.client.Stream(ctx, &pb.StreamRequest{
		IncludeInitialSnapshot: includeInitialSnapshot,
	})
	
	// 問題46: 到達不可能なコード（実際にはないが、パターンとして）
	// return nil, fmt.Errorf("unreachable")
}
