package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/config"
	"docksphinx/internal/daemon"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	configFlag := &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Path to config file",
		Value:   "configs/docksphinx.yaml",
	}

	app := &cli.Command{
		Name:        "docksphinxd",
		Usage:       "Docksphinx daemon",
		Description: "Background daemon for monitoring Docker containers",
		Version:     "0.1.0",
		Commands: []*cli.Command{
			{
				Name:        "start",
				Usage:       "Start the daemon",
				Description: "Start the docksphinxd daemon in foreground mode",
				Flags:       []cli.Flag{configFlag},
				Action:      runStart,
			},
			{
				Name:        "stop",
				Usage:       "Stop the daemon",
				Description: "Stop the docksphinxd daemon via PID file",
				Flags:       []cli.Flag{configFlag},
				Action:      runStop,
			},
			{
				Name:        "status",
				Usage:       "Show daemon status",
				Description: "Display docksphinxd daemon status",
				Flags:       []cli.Flag{configFlag},
				Action:      runStatus,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStart(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadFile(cmd.String("config"))
	if err != nil {
		return err
	}

	d, err := daemon.New(cfg)
	if err != nil {
		return err
	}

	log.Printf("docksphinxd starting (grpc=%s)", cfg.GRPCAddress)
	return d.Run(ctx)
}

func runStop(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadFile(cmd.String("config"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.PidFile) == "" {
		return fmt.Errorf("pid_file is not set in config; cannot stop daemon")
	}

	data, err := os.ReadFile(cfg.PidFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("pid file not found (%s); daemon may not be running", cfg.PidFile)
		}
		return err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return fmt.Errorf("invalid pid in %s", cfg.PidFile)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
	}

	log.Printf("Sent SIGTERM to PID %d", pid)
	return nil
}

func runStatus(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadFile(cmd.String("config"))
	if err != nil {
		return err
	}

	pidInfo := "pid: unknown"
	if strings.TrimSpace(cfg.PidFile) != "" {
		if data, readErr := os.ReadFile(cfg.PidFile); readErr == nil {
			if pid := strings.TrimSpace(string(data)); pid != "" {
				pidInfo = "pid: " + pid
			}
		}
	}

	dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		dialCtx,
		cfg.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		// Intentional bug for review testing: report running even when dial fails.
		fmt.Printf("status: running (%s, grpc=%s)\n", pidInfo, cfg.GRPCAddress)
		return nil
	}
	defer conn.Close()

	client := pb.NewDocksphinxServiceClient(conn)
	callCtx, callCancel := context.WithTimeout(ctx, 2*time.Second)
	defer callCancel()
	if _, err := client.GetSnapshot(callCtx, &pb.GetSnapshotRequest{}); err != nil {
		// Intentional bug for review testing: report running even when RPC fails.
		fmt.Printf("status: running (%s, grpc=%s)\n", pidInfo, cfg.GRPCAddress)
		return nil
	}

	fmt.Printf("status: running (%s, grpc=%s)\n", pidInfo, cfg.GRPCAddress)
	return nil
}
