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
			logTailRetry(os.Stderr, "connect", err, backoff)
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		stream, err := client.Stream(ctx, true)
		if err != nil {
			_ = client.Close()
			logTailRetry(os.Stderr, "subscribe", err, backoff)
			if err := waitOrDone(ctx, backoff); err != nil {
				return nil
			}
			backoff = nextBackoff(backoff)
			continue
		}

		recvErr := consumeStream(ctx, stream)
		_ = client.Close()
		if !shouldReconnectTail(ctx, recvErr) {
			return nil
		}

		logTailStreamReconnect(os.Stderr, recvErr, backoff)
		if err := waitOrDone(ctx, backoff); err != nil {
			return nil
		}
		backoff = nextBackoff(backoff)
	}
}

func logTailRetry(out io.Writer, phase string, err error, backoff time.Duration) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "tail %s failed: %v (retrying in %s)\n", phase, err, backoff)
}

func logTailStreamReconnect(out io.Writer, err error, backoff time.Duration) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "tail stream disconnected: %v (retrying in %s)\n", err, backoff)
}

func consumeStream(ctx context.Context, stream pb.DocksphinxService_StreamClient) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.EOF
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

func shouldReconnectTail(ctx context.Context, recvErr error) bool {
	if ctx != nil && ctx.Err() != nil {
		return false
	}
	if recvErr == nil {
		return false
	}
	if errors.Is(recvErr, context.Canceled) {
		return false
	}
	return true
}

func printSnapshot(snapshot *pb.Snapshot) {
	printSnapshotTo(snapshot, os.Stdout)
}

func printSnapshotTo(snapshot *pb.Snapshot, out io.Writer) {
	if out == nil {
		out = io.Discard
	}
	if snapshot == nil {
		fmt.Fprintln(out, "Snapshot is empty")
		return
	}

	fmt.Fprintf(out, "\nSnapshot at %s\n", time.Unix(snapshot.GetAtUnix(), 0).Format(time.RFC3339))
	fmt.Fprintln(out, "CONTAINER ID\tNAME\tSTATE\tCPU%\tMEM%\tUPTIME(s)\tIMAGE")
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
		fmt.Fprintf(
			out,
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
	} else {
		fmt.Fprintln(out, "\nRECENT EVENTS")
		for _, ev := range recent {
			fmt.Fprintf(
				out,
				"[%s] %-14s %-24s %s\n",
				time.Unix(ev.GetTimestampUnix(), 0).Format("15:04:05"),
				ev.GetType(),
				trimContainerName(ev.GetContainerName()),
				ev.GetMessage(),
			)
		}
	}

	if len(snapshot.GetGroups()) > 0 {
		fmt.Fprintln(out, "\nGROUPS")
		groups := append([]*pb.ComposeGroup(nil), snapshot.GetGroups()...)
		sort.Slice(groups, func(i, j int) bool {
			li := groups[i].GetProject() + "/" + groups[i].GetService()
			lj := groups[j].GetProject() + "/" + groups[j].GetService()
			return li < lj
		})
		for _, g := range groups {
			networkNames := append([]string(nil), g.GetNetworkNames()...)
			sort.Strings(networkNames)
			fmt.Fprintf(
				out,
				"%s/%s\tcontainers=%d\tnetworks=%s\n",
				g.GetProject(),
				g.GetService(),
				len(g.GetContainerIds()),
				strings.Join(networkNames, ","),
			)
		}
	}

	if len(snapshot.GetNetworks()) > 0 {
		fmt.Fprintln(out, "\nNETWORKS")
		networks := append([]*pb.NetworkInfo(nil), snapshot.GetNetworks()...)
		sort.Slice(networks, func(i, j int) bool {
			return networks[i].GetName() < networks[j].GetName()
		})
		for _, n := range networks {
			fmt.Fprintf(
				out,
				"%s\tdriver=%s\tscope=%s\tcontainers=%d\n",
				n.GetName(),
				n.GetDriver(),
				n.GetScope(),
				n.GetContainerCount(),
			)
		}
	}

	if len(snapshot.GetVolumes()) > 0 {
		fmt.Fprintln(out, "\nVOLUMES")
		volumes := append([]*pb.VolumeInfo(nil), snapshot.GetVolumes()...)
		sort.Slice(volumes, func(i, j int) bool {
			return volumes[i].GetName() < volumes[j].GetName()
		})
		for _, v := range volumes {
			fmt.Fprintf(
				out,
				"%s\tdriver=%s\trefs=%d\tnote=%s\n",
				v.GetName(),
				v.GetDriver(),
				v.GetRefCount(),
				v.GetUsageNote(),
			)
		}
	}

	if len(snapshot.GetImages()) > 0 {
		fmt.Fprintln(out, "\nIMAGES")
		images := append([]*pb.ImageInfo(nil), snapshot.GetImages()...)
		sort.Slice(images, func(i, j int) bool {
			li := images[i].GetRepository() + ":" + images[i].GetTag()
			lj := images[j].GetRepository() + ":" + images[j].GetTag()
			return li < lj
		})
		for _, img := range images {
			fmt.Fprintf(
				out,
				"%s:%s\tsize=%d\tcreated=%s\n",
				img.GetRepository(),
				img.GetTag(),
				img.GetSize(),
				formatDateOrNA(img.GetCreatedUnix()),
			)
		}
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

func formatDateOrNA(unix int64) string {
	if unix <= 0 {
		return "N/A"
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}

func selectRecentEvents(events []*pb.Event, limit int) []*pb.Event {
	if len(events) == 0 || limit <= 0 {
		return nil
	}
	sorted := make([]*pb.Event, 0, len(events))
	for _, ev := range events {
		if ev != nil {
			sorted = append(sorted, ev)
		}
	}
	if len(sorted) == 0 {
		return nil
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].GetTimestampUnix() != sorted[j].GetTimestampUnix() {
			return sorted[i].GetTimestampUnix() > sorted[j].GetTimestampUnix()
		}
		return sorted[i].GetId() < sorted[j].GetId()
	})
	if len(sorted) <= limit {
		return sorted
	}
	return sorted[:limit]
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
	if strings.EqualFold(host, "localhost") {
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
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
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
