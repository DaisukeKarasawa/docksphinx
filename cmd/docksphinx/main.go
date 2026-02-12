package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
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
			client, err := dgrpc.NewClient(ctx, address)
			if err != nil {
				return wrapDaemonError("connect", address, err)
			}
			defer client.Close()

			snapshot, err := client.GetSnapshot(ctx)
			if err != nil {
				return wrapDaemonError("get snapshot", address, err)
			}

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

			ctx, cancel := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
			defer cancel()

			client, err := dgrpc.NewClient(ctx, address)
			if err != nil {
				return wrapDaemonError("connect", address, err)
			}
			defer client.Close()

			stream, err := client.Stream(ctx, true)
			if err != nil {
				return wrapDaemonError("stream", address, err)
			}

			for {
				update, err := stream.Recv()
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
						return nil
					}
					if errors.Is(err, io.EOF) {
						return nil
					}
					return wrapDaemonError("stream recv", address, err)
				}

				switch payload := update.GetPayload().(type) {
				case *pb.StreamUpdate_Snapshot:
					printSnapshot(payload.Snapshot)
				case *pb.StreamUpdate_Event:
					printEvent(payload.Event)
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
	if snapshot == nil {
		fmt.Println("Snapshot is empty")
		return
	}

	fmt.Printf("Snapshot at %s\n", time.Unix(snapshot.GetAtUnix(), 0).Format(time.RFC3339))
	fmt.Println("CONTAINER ID\tNAME\tIMAGE\tSTATE\tCPU%\tMEM%")

	for _, container := range snapshot.GetContainers() {
		metrics := snapshot.GetMetrics()[container.GetContainerId()]
		cpu := 0.0
		mem := 0.0
		if metrics != nil {
			cpu = metrics.GetCpuPercent()
			mem = metrics.GetMemoryPercent()
		}

		fmt.Printf(
			"%s\t%s\t%s\t%s\t%.2f\t%.2f\n",
			shortContainerID(container.GetContainerId()),
			container.GetContainerName(),
			container.GetImageName(),
			container.GetState(),
			cpu,
			mem,
		)
	}
}

func printEvent(event *pb.Event) {
	if event == nil {
		return
	}

	fmt.Printf(
		"[%s] %-14s %-20s %s\n",
		time.Unix(event.GetTimestampUnix(), 0).Format("15:04:05"),
		event.GetType(),
		event.GetContainerName(),
		event.GetMessage(),
	)
}

func shortContainerID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func wrapDaemonError(op, address string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "connection refused") {
		return fmt.Errorf("%s daemon (%s): %w. start daemon with `docksphinxd start`", op, address, err)
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded:
			return fmt.Errorf("%s daemon (%s): %w. check daemon status and run `docksphinxd start` if needed", op, address, err)
		}
	}

	return fmt.Errorf("%s daemon (%s): %w", op, address, err)
}
