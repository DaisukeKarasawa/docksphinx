package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"docksphinx/internal/config"
	"docksphinx/internal/daemon"
	dgrpc "docksphinx/internal/grpc"
	"github.com/urfave/cli/v3"
)

func main() {
	configFlag := &cli.StringFlag{
		Name:  "config",
		Usage: "Path to docksphinx config YAML",
	}

	app := &cli.Command{
		Name:  "docksphinxd",
		Usage: "Docksphinx daemon",
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start daemon in foreground",
				Flags:  []cli.Flag{configFlag},
				Action: runStart,
			},
			{
				Name:   "stop",
				Usage:  "Stop daemon by PID file",
				Flags:  []cli.Flag{configFlag},
				Action: runStop,
			},
			{
				Name:   "status",
				Usage:  "Check daemon status by PID + gRPC health",
				Flags:  []cli.Flag{configFlag},
				Action: runStatus,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStart(parent context.Context, cmd *cli.Command) error {
	cfg, cfgPath, err := config.Load(cmd.String("config"))
	if err != nil {
		return err
	}
	if cfgPath != "" {
		fmt.Printf("Using config: %s\n", cfgPath)
	} else {
		fmt.Println("Using config: defaults")
	}

	d, err := daemon.New(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return d.Run(ctx)
}

func runStop(parent context.Context, cmd *cli.Command) error {
	cfg, _, err := config.Load(cmd.String("config"))
	if err != nil {
		return err
	}

	pid, err := readPID(cfg.Daemon.PIDFile)
	if err != nil {
		return err
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			if rmErr := removePIDFileIfExists(cfg.Daemon.PIDFile); rmErr != nil {
				return fmt.Errorf("process %d is already stopped, but failed to remove stale pid file: %w", pid, rmErr)
			}
			fmt.Printf("Process %d is already stopped (stale PID file removed)\n", pid)
			return nil
		}
		return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
	}
	fmt.Printf("Sent SIGTERM to PID %d\n", pid)

	waitCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := waitForProcessExit(waitCtx, pid, 100*time.Millisecond, func(targetPID int) error {
		return syscall.Kill(targetPID, 0)
	}); err != nil {
		return err
	}
	if err := removePIDFileIfExists(cfg.Daemon.PIDFile); err != nil {
		return fmt.Errorf("process %d stopped but failed to remove pid file: %w", pid, err)
	}
	fmt.Printf("Process %d stopped\n", pid)
	return nil
}

func runStatus(ctx context.Context, cmd *cli.Command) error {
	cfg, _, err := config.Load(cmd.String("config"))
	if err != nil {
		return err
	}

	pidStatus := "pid: not found"
	if pid, err := readPID(cfg.Daemon.PIDFile); err == nil {
		if err := syscall.Kill(pid, 0); err == nil {
			pidStatus = fmt.Sprintf("pid: %d (alive)", pid)
		} else {
			pidStatus = fmt.Sprintf("pid: %d (stale)", pid)
		}
	}

	healthErr := checkGRPCHealth(ctx, cfg.GRPC.Address, time.Duration(cfg.GRPC.Timeout)*time.Second)
	if healthErr != nil {
		fmt.Printf("status: not running (%s, grpc=%s, err=%v)\n", pidStatus, cfg.GRPC.Address, healthErr)
		return healthErr
	}

	fmt.Printf("status: running (%s, grpc=%s)\n", pidStatus, cfg.GRPC.Address)
	return nil
}

func readPID(path string) (int, error) {
	// #nosec G304 -- path is loaded from validated config and expected absolute pid path.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("pid file not found: %s", path)
		}
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid pid in %s", path)
	}
	return pid, nil
}

func checkGRPCHealth(parent context.Context, address string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	client, err := dgrpc.NewClient(ctx, address)
	if err != nil {
		return fmt.Errorf("dial daemon: %w", err)
	}
	defer client.Close()

	callCtx, callCancel := context.WithTimeout(parent, 2*time.Second)
	defer callCancel()
	if _, err := client.GetSnapshot(callCtx); err != nil {
		return fmt.Errorf("health rpc failed: %w", err)
	}
	return nil
}

func waitForProcessExit(
	ctx context.Context,
	pid int,
	interval time.Duration,
	checker func(int) error,
) error {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("process %d did not stop within timeout", pid)
		case <-ticker.C:
			err := checker(pid)
			if err == nil {
				continue
			}
			if errors.Is(err, syscall.ESRCH) {
				return nil
			}
			if errors.Is(err, syscall.EPERM) {
				return fmt.Errorf("permission denied while checking process %d", pid)
			}
			return fmt.Errorf("failed to check process %d: %w", pid, err)
		}
	}
}

func removePIDFileIfExists(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
