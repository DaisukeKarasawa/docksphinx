package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/config"
	dgrpc "docksphinx/internal/grpc"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	configFlag := &cli.StringFlag{
		Name:  "config",
		Usage: "Path to docksphinx YAML config",
	}
	addrFlag := &cli.StringFlag{
		Name:  "addr",
		Usage: "gRPC daemon address (overrides config)",
	}
	insecureFlag := &cli.BoolFlag{
		Name:  "insecure",
		Usage: "allow insecure plaintext warning suppression for non-loopback address",
	}

	app := &cli.Command{
		Name:  "docksphinx",
		Usage: "Docksphinx CLI",
		Commands: []*cli.Command{
			{
				Name:  "snapshot",
				Usage: "Get current state once",
				Flags: []cli.Flag{configFlag, addrFlag, insecureFlag},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					address, err := resolveAddress(cmd.String("config"), cmd.String("addr"))
					if err != nil {
						return err
					}
					warnInsecure(address, cmd.Bool("insecure"))
					return runSnapshot(ctx, address)
				},
			},
			{
				Name:  "tail",
				Usage: "Stream events and snapshots continuously",
				Flags: []cli.Flag{configFlag, addrFlag, insecureFlag},
				Action: func(parent context.Context, cmd *cli.Command) error {
					address, err := resolveAddress(cmd.String("config"), cmd.String("addr"))
					if err != nil {
						return err
					}
					warnInsecure(address, cmd.Bool("insecure"))
					return runTail(parent, address)
				},
			},
			{
				Name:  "tui",
				Usage: "Interactive TUI",
				Flags: []cli.Flag{configFlag, addrFlag, insecureFlag},
				Action: func(parent context.Context, cmd *cli.Command) error {
					address, err := resolveAddress(cmd.String("config"), cmd.String("addr"))
					if err != nil {
						return err
					}
					warnInsecure(address, cmd.Bool("insecure"))
					return runTUI(parent, address)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func resolveAddress(configPath, addrOverride string) (string, error) {
	if strings.TrimSpace(addrOverride) != "" {
		return strings.TrimSpace(addrOverride), nil
	}
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return "", err
	}
	return cfg.GRPC.Address, nil
}

func runSnapshot(ctx context.Context, address string) error {
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
}

func runTail(parent context.Context, address string) error {
	ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	backoff := 500 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return nil
		}

		client, err := dgrpc.NewClient(ctx, address)
		if err != nil {
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		stream, err := client.Stream(ctx, true)
		if err != nil {
			_ = client.Close()
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		recvErr := consumeStream(ctx, stream)
		_ = client.Close()
		if recvErr == nil || errors.Is(recvErr, context.Canceled) {
			return nil
		}

		fmt.Fprintf(os.Stderr, "tail stream disconnected: %v (reconnecting)\n", recvErr)
		if err := waitOrDone(ctx, backoff); err != nil {
			return nil
		}
		backoff = nextBackoff(backoff)
	}
}

func consumeStream(ctx context.Context, stream pb.DocksphinxService_StreamClient) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
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
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func printSnapshot(snapshot *pb.Snapshot) {
	if snapshot == nil {
		fmt.Println("Snapshot is empty")
		return
	}

	fmt.Printf("\nSnapshot at %s\n", time.Unix(snapshot.GetAtUnix(), 0).Format(time.RFC3339))
	fmt.Println("CONTAINER ID\tNAME\tSTATE\tCPU%\tMEM%\tUPTIME(s)\tIMAGE")
	containers := append([]*pb.ContainerInfo(nil), snapshot.GetContainers()...)
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].GetContainerName() < containers[j].GetContainerName()
	})
	for _, c := range containers {
		m := snapshot.GetMetrics()[c.GetContainerId()]
		cpu := "N/A"
		mem := "N/A"
		if m != nil {
			cpu = fmt.Sprintf("%.2f", m.GetCpuPercent())
			mem = fmt.Sprintf("%.2f", m.GetMemoryPercent())
		}
		uptime := formatUptimeOrNA(c)
		fmt.Printf(
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			shortContainerID(c.GetContainerId()),
			trimContainerName(c.GetContainerName()),
			c.GetState(),
			cpu,
			mem,
			uptime,
			c.GetImageName(),
		)
	}

	recent := selectRecentEvents(snapshot.GetRecentEvents(), 10)
	if len(recent) == 0 {
		return
	}
	fmt.Println("\nRECENT EVENTS")
	for _, ev := range recent {
		fmt.Printf(
			"[%s] %-14s %-24s %s\n",
			time.Unix(ev.GetTimestampUnix(), 0).Format("15:04:05"),
			ev.GetType(),
			trimContainerName(ev.GetContainerName()),
			ev.GetMessage(),
		)
	}
}

func printEvent(ev *pb.Event) {
	if ev == nil {
		return
	}
	fmt.Printf(
		"[%s] %-14s %-24s %s\n",
		time.Unix(ev.GetTimestampUnix(), 0).Format("15:04:05"),
		ev.GetType(),
		trimContainerName(ev.GetContainerName()),
		ev.GetMessage(),
	)
}

func shortContainerID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func trimContainerName(name string) string {
	return strings.TrimPrefix(name, "/")
}

func formatUptimeOrNA(c *pb.ContainerInfo) string {
	if c == nil {
		return "N/A"
	}
	if c.GetStartedAtUnix() <= 0 && c.GetUptimeSeconds() <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", c.GetUptimeSeconds())
}

func selectRecentEvents(events []*pb.Event, limit int) []*pb.Event {
	if len(events) == 0 || limit <= 0 {
		return nil
	}
	if len(events) <= limit {
		return events
	}
	return events[:limit]
}

func waitOrDone(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > 5*time.Second {
		return 5 * time.Second
	}
	return next
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
	if insecure || isLoopback(address) {
		return
	}
	fmt.Fprintf(os.Stderr, "WARNING: connecting to %s over plaintext (no TLS)\n", address)
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
			return fmt.Errorf("%s daemon (%s): %w. check daemon status", op, address, err)
		}
	}
	return fmt.Errorf("%s daemon (%s): %w", op, address, err)
}
