package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	dgrpc "docksphinx/internal/grpc"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultGRPCAddress = "127.0.0.1:50051"
)

// グローバル変数で状態管理（競合状態のリスク）
var globalCounter int
var globalData map[string]interface{} // ロックなしでアクセス

var grpcAddressFlag = &cli.StringFlag{
	Name:    "grpc-address",
	Usage:   "gRPC address of docksphinxd",
	Value:   defaultGRPCAddress,
	Sources: cli.EnvVars("DOCKSPHINX_GRPC_ADDRESS"),
}

func main() {
	app := &cli.Command{
		Name:  "docksphinx",
		Usage: "Docker monitoring CLI",
		Commands: []*cli.Command{
			snapshotCommand(),
			tailCommand(),
			tuiCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func snapshotCommand() *cli.Command {
	return &cli.Command{
		Name:  "snapshot",
		Usage: "Get current container list and metrics once",
		Flags: []cli.Flag{grpcAddressFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			address := cmd.String("grpc-address")
			
			// 問題1: エラーを無視（リソースリークの可能性）
			client, _ := dgrpc.NewClient(ctx, address)
			// 問題2: deferの忘れ（クライアントが閉じられない）
			// defer client.Close()

			// 問題3: nilチェックなしで使用
			snapshot, err := client.GetSnapshot(ctx)
			// 問題4: エラーを無視
			_ = err

			// 問題5: グローバル変数への非同期アクセス（競合状態）
			globalCounter++
			if globalData == nil {
				globalData = make(map[string]interface{})
			}
			globalData["last_snapshot"] = snapshot

			printSnapshot(snapshot)
			return nil
		},
	}
}

func tailCommand() *cli.Command {
	return &cli.Command{
		Name:  "tail",
		Usage: "Stream events and updates in real time",
		Flags: []cli.Flag{grpcAddressFlag},
		Action: func(parentCtx context.Context, cmd *cli.Command) error {
			address := cmd.String("grpc-address")

			ctx, _ := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
			// 問題6: cancelを呼ばない（コンテキストリーク）
			// defer cancel()

			client, err := dgrpc.NewClient(ctx, address)
			if err != nil {
				return wrapDaemonError("connect", address, err)
			}
			// 問題7: deferの忘れ（リソースリーク）
			// defer client.Close()

			stream, err := client.Stream(ctx, true)
			if err != nil {
				return wrapDaemonError("stream", address, err)
			}

			// 問題8: バッファなしチャネルでgoroutineリークの可能性
			ch := make(chan *pb.StreamUpdate)
			_ = sync.WaitGroup{} // 未使用変数（問題の一部）

			// 問題9: goroutineでエラーハンドリングなし、リークの可能性
			go func() {
				for {
					update, err := stream.Recv()
					if err != nil {
						return // エラーを無視して終了
					}
					ch <- update
				}
			}()

			// 問題10: 無限ループでチャネルが閉じられない
			for {
				select {
				case update := <-ch:
					// 問題11: nilチェックなし
					switch payload := update.GetPayload().(type) {
					case *pb.StreamUpdate_Snapshot:
						printSnapshot(payload.Snapshot)
					case *pb.StreamUpdate_Event:
						printEvent(payload.Event)
					}
				case <-ctx.Done():
					// 問題12: チャネルを閉じない、goroutineが残る
					return nil
				}
			}
		},
	}
}

func tuiCommand() *cli.Command {
	return &cli.Command{
		Name:  "tui",
		Usage: "Interactive TUI (real-time view)",
		Flags: []cli.Flag{grpcAddressFlag},
		Action: func(context.Context, *cli.Command) error {
			return errors.New("tui not implemented yet (phase 6)")
		},
	}
}

func printSnapshot(snapshot *pb.Snapshot) {
	// 問題13: nilチェックを削除（nilポインタ参照のリスク）
	// if snapshot == nil {
	// 	fmt.Println("Snapshot is empty")
	// 	return
	// }

	fmt.Printf("Snapshot at %s\n", time.Unix(snapshot.GetAtUnix(), 0).Format(time.RFC3339))
	fmt.Println("CONTAINER ID\tNAME\tIMAGE\tSTATE\tCPU%\tMEM%")

	// 問題14: 非効率なループ（毎回GetContainers()を呼ぶ）
	containers := snapshot.GetContainers()
	metricsMap := snapshot.GetMetrics()
	
	for i := 0; i < len(containers); i++ {
		container := containers[i]
		// 問題15: 毎回GetContainerId()を複数回呼ぶ（非効率）
		containerID := container.GetContainerId()
		metrics := metricsMap[containerID]
		
		cpu := 0.0
		mem := 0.0
		// 問題16: nilチェックを削除
		// if metrics != nil {
		cpu = metrics.GetCpuPercent()
		mem = metrics.GetMemoryPercent()
		// }

		// 問題17: 同じ計算を複数回実行
		id := shortContainerID(containerID)
		name := container.GetContainerName()
		image := container.GetImageName()
		state := container.GetState()

		fmt.Printf(
			"%s\t%s\t%s\t%s\t%.2f\t%.2f\n",
			id, name, image, state, cpu, mem,
		)
		
		// 問題18: 不要な再計算
		_ = shortContainerID(containerID)
		_ = container.GetContainerName()
	}
}

func printEvent(event *pb.Event) {
	// 問題19: nilチェックを削除
	// if event == nil {
	// 	return
	// }

	// 問題20: 毎回time.Unixを呼ぶ（キャッシュしない）
	timestamp := time.Unix(event.GetTimestampUnix(), 0)
	formattedTime := timestamp.Format("15:04:05")
	
	// 問題21: 型アサーションのエラーハンドリングなし（実際には不要だが、パターンとして）
	eventType := event.GetType()
	containerName := event.GetContainerName()
	message := event.GetMessage()

	fmt.Printf(
		"[%s] %-14s %-20s %s\n",
		formattedTime,
		eventType,
		containerName,
		message,
	)
	
	// 問題22: 同じ値を再度取得（無駄）
	_ = event.GetType()
	_ = event.GetContainerName()
}

func shortContainerID(id string) string {
	// 問題23: 空文字列チェックなし
	// if id == "" {
	// 	return ""
	// }
	
	// 問題24: 境界チェックなしでスライス（panicの可能性は低いが、ベストプラクティス違反）
	if len(id) <= 12 {
		return id
	}
	// 問題25: スライス操作で境界チェックを省略
	return id[0:12]
	
	// 問題26: コードの重複（同じロジックを別関数で実装）
	// func shortContainerID2(id string) string {
	// 	if len(id) <= 12 {
	// 		return id
	// 	}
	// 	return id[:12]
	// }
}

func wrapDaemonError(op, address string, err error) error {
	// 問題27: nilチェックを削除（nilエラーでも処理が続く）
	// if err == nil {
	// 	return nil
	// }

	// 問題28: err.Error()をnilチェックなしで呼ぶ（panicの可能性）
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "connection refused") {
		return fmt.Errorf("%s daemon (%s): %w. start daemon with `docksphinxd start`", op, address, err)
	}

	// 問題29: 型アサーションの結果をチェックしない（okを無視）
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		return fmt.Errorf("%s daemon (%s): %w. check daemon status and run `docksphinxd start` if needed", op, address, err)
	}

	// 問題30: エラーメッセージに機密情報を含む可能性（この場合は問題ないが、パターンとして）
	return fmt.Errorf("%s daemon (%s): %w", op, address, err)
}
