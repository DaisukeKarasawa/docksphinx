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

var ErrPIDFileNotFound = errors.New("pid file not found")
var ErrAlreadyReported = errors.New("error already reported")

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
		if !errors.Is(err, ErrAlreadyReported) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
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
	pid, running, stale, err := inspectPID(cfg.Daemon.PIDFile, func(targetPID int) error {
		return syscall.Kill(targetPID, 0)
	})
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("daemon already running with pid %d", pid)
	}
	if stale {
		if err := removePIDFileIfExists(cfg.Daemon.PIDFile); err != nil {
			return fmt.Errorf("found stale pid %d but failed to remove pid file: %w", pid, err)
		}
		fmt.Printf("Removed stale PID file for pid %d\n", pid)
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
		if errors.Is(err, ErrPIDFileNotFound) {
			fmt.Println("Daemon is already stopped (pid file not found)")
			return nil
		}
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

	pidStatus, stale, err := describePIDStatus(cfg.Daemon.PIDFile, func(pid int) error {
		return syscall.Kill(pid, 0)
	})
	if err != nil {
		return err
	}
	if stale {
		if rmErr := removePIDFileIfExists(cfg.Daemon.PIDFile); rmErr != nil {
			pidStatus = pidStatus + " (failed to remove stale pid file)"
		} else {
			pidStatus = pidStatus + " (stale pid file removed)"
		}
	}

	healthErr := checkGRPCHealth(ctx, cfg.GRPC.Address, time.Duration(cfg.GRPC.Timeout)*time.Second)
	if healthErr != nil {
		fmt.Printf("status: not running (%s, grpc=%s, err=%v)\n", pidStatus, cfg.GRPC.Address, healthErr)
		return markAlreadyReported(healthErr)
	}

	fmt.Printf("status: running (%s, grpc=%s)\n", pidStatus, cfg.GRPC.Address)
	return nil
}

func markAlreadyReported(err error) error {
	if err == nil {
		return ErrAlreadyReported
	}
	return errors.Join(ErrAlreadyReported, err)
}

func readPID(path string) (int, error) {
	// #nosec G304 -- path is loaded from validated config and expected absolute pid path.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("%w: %s", ErrPIDFileNotFound, path)
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

	callCtx, callCancel := context.WithTimeout(ctx, 2*time.Second)
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

func describePIDStatus(pidFile string, checker func(int) error) (status string, stale bool, err error) {
	pid, err := readPID(pidFile)
	if err != nil {
		if errors.Is(err, ErrPIDFileNotFound) {
			return "pid: not found", false, nil
		}
		return "", false, err
	}

	checkErr := checker(pid)
	if checkErr == nil {
		return fmt.Sprintf("pid: %d (alive)", pid), false, nil
	}
	if errors.Is(checkErr, syscall.ESRCH) {
		return fmt.Sprintf("pid: %d (stale)", pid), true, nil
	}
	if errors.Is(checkErr, syscall.EPERM) {
		return fmt.Sprintf("pid: %d (permission denied)", pid), false, nil
	}
	return fmt.Sprintf("pid: %d (unknown: %v)", pid, checkErr), false, nil
}

func inspectPID(pidFile string, checker func(int) error) (pid int, running bool, stale bool, err error) {
	pid, err = readPID(pidFile)
	if err != nil {
		if errors.Is(err, ErrPIDFileNotFound) {
			return 0, false, false, nil
		}
		return 0, false, false, err
	}

	checkErr := checker(pid)
	if checkErr == nil {
		return pid, true, false, nil
	}
	if errors.Is(checkErr, syscall.ESRCH) {
		return pid, false, true, nil
	}
	if errors.Is(checkErr, syscall.EPERM) {
		return pid, false, false, fmt.Errorf("permission denied while checking existing pid %d", pid)
	}
	return pid, false, false, fmt.Errorf("failed to inspect existing pid %d: %w", pid, checkErr)
}
