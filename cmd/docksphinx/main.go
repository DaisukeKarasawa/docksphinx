package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
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

var insecureFlag = &cli.BoolFlag{
	Name:  "insecure",
	Usage: "allow insecure (plaintext) gRPC connection to non-loopback addresses",
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

func isLoopback(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func warnInsecure(address string, insecure bool) {
	if !insecure && !isLoopback(address) {
		fmt.Fprintf(os.Stderr, "WARNING: connecting to %s over plaintext (no TLS). Use --insecure to suppress this warning.\n", address)
	}
}

func snapshotCommand() *cli.Command {
	return &cli.Command{
		Name:  "snapshot",
		Usage: "Get current container list and metrics once",
		Flags: []cli.Flag{grpcAddressFlag, insecureFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			address := cmd.String("grpc-address")
			warnInsecure(address, cmd.Bool("insecure"))

			client, err := dgrpc.NewClient(ctx, address)
			if err != nil {
				return wrapDaemonError("connect", address, err)
			}
			defer client.Close()

			snapshot, err := client.GetSnapshot(ctx)
			if err != nil {
				return wrapDaemonError("snapshot", address, err)
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
		Flags: []cli.Flag{grpcAddressFlag, insecureFlag},
			Action: func(parentCtx context.Context, cmd *cli.Command) error {
				address := cmd.String("grpc-address")
				warnInsecure(address, cmd.Bool("insecure"))

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

			ch := make(chan *pb.StreamUpdate, 1)
			errs := make(chan error, 1)

			go func() {
					defer close(ch)
					for {
						update, err := stream.Recv()
						if err != nil {
							if !errors.Is(err, io.EOF) {
								errs <- err
							}
							return
						}
					select {
					case ch <- update:
					case <-ctx.Done():
						return
					}
				}
			}()

			for {
				select {
				case update, ok := <-ch:
					if !ok {
						select {
						case err := <-errs:
							return wrapDaemonError("stream", address, err)
						default:
							return nil
						}
					}
					if update == nil {
						continue
					}
					switch payload := update.GetPayload().(type) {
					case *pb.StreamUpdate_Snapshot:
						printSnapshot(payload.Snapshot)
					case *pb.StreamUpdate_Event:
						printEvent(payload.Event)
					}
				case err := <-errs:
					return wrapDaemonError("stream", address, err)
				case <-ctx.Done():
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
	if snapshot == nil {
		fmt.Println("Snapshot is empty")
		return
	}

	fmt.Printf("Snapshot at %s\n", time.Unix(snapshot.GetAtUnix(), 0).Format(time.RFC3339))
	fmt.Println("CONTAINER ID\tNAME\tIMAGE\tSTATE\tCPU%\tMEM%")

	containers := snapshot.GetContainers()
	metricsMap := snapshot.GetMetrics()

	for _, container := range containers {
		containerID := container.GetContainerId()
		metrics := metricsMap[containerID]

		cpu := 0.0
		mem := 0.0
		if metrics != nil {
			cpu = metrics.GetCpuPercent()
			mem = metrics.GetMemoryPercent()
		}

		name := container.GetContainerName()
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}

		fmt.Printf(
			"%s\t%s\t%s\t%s\t%.2f\t%.2f\n",
			shortContainerID(containerID),
			name,
			container.GetImageName(),
			container.GetState(),
			cpu, mem,
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

func isConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return errors.Is(opErr.Err, syscall.ECONNREFUSED)
	}
	return false
}

func wrapDaemonError(op, address string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || isConnectionRefused(err) {
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
