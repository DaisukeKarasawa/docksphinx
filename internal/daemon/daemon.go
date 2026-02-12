package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"docksphinx/internal/config"
	"docksphinx/internal/docker"
	"docksphinx/internal/event"
	internalgrpc "docksphinx/internal/grpc"
	"docksphinx/internal/monitor"
)

// Daemon runs the monitoring engine and gRPC server.
type Daemon struct {
	config   *config.Config
	docker   *docker.Client
	engine   *monitor.Engine
	grpc     *internalgrpc.Server
	history  *event.History
	pidPath  string
	logger   *slog.Logger

	mu      sync.Mutex
	runCtx  context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
}

// New creates a daemon from config (without starting it).
func New(cfg *config.Config) (*Daemon, error) {
	if cfg == nil {
		cfg = config.Default()
	}

	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}

	engineCfg := monitor.EngineConfig{
		Interval:             cfg.Monitor.Interval,
		ContainerNamePattern: cfg.Monitor.ContainerNamePattern,
		ImageNamePattern:     cfg.Monitor.ImageNamePattern,
		Thresholds:           cfg.Monitor.Thresholds,
	}
	engine, err := monitor.NewEngine(engineCfg, client)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("monitor engine: %w", err)
	}

	srv, err := internalgrpc.NewServer(&internalgrpc.ServerOptions{Address: cfg.GRPCAddress}, engine)
	if err != nil {
		engine.Stop()
		_ = client.Close()
		return nil, fmt.Errorf("grpc server: %w", err)
	}

	level := parseLevel(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	return &Daemon{
		config:  cfg,
		docker:  client,
		engine:  engine,
		grpc:    srv,
		history: event.NewHistory(cfg.EventHistoryMax),
		pidPath: cfg.PidFile,
		logger:  logger,
	}, nil
}

// Run starts the engine and gRPC server, and blocks until context cancellation or SIGINT/SIGTERM.
func (d *Daemon) Run(parent context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("daemon is already running")
	}
	d.runCtx, d.cancel = context.WithCancel(parent)
	d.running = true
	d.mu.Unlock()

	d.logger.Info("daemon starting", "grpc_address", d.config.GRPCAddress)

	if err := d.engine.Start(); err != nil {
		d.setStopped()
		return fmt.Errorf("start monitor engine: %w", err)
	}
	d.logger.Info("monitor engine started")

	d.wg.Add(1)
	go d.collectHistoryLoop()

	grpcErr := make(chan error, 1)
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := d.grpc.Start(); err != nil {
			grpcErr <- err
		}
	}()

	if err := d.writePID(); err != nil {
		d.logger.Warn("failed to write pid file", "error", err, "pid_file", d.pidPath)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case <-d.runCtx.Done():
		d.logger.Info("run context canceled")
	case sig := <-sigCh:
		d.logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-grpcErr:
		d.logger.Error("grpc server stopped unexpectedly", "error", err)
		_ = d.Stop()
		return fmt.Errorf("grpc server: %w", err)
	}

	return d.Stop()
}

// Stop gracefully stops grpc, monitor engine, and docker client.
func (d *Daemon) Stop() error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = false
	if d.cancel != nil {
		d.cancel()
	}
	d.mu.Unlock()

	d.logger.Info("daemon stopping")
	d.grpc.Stop()
	d.engine.Stop()
	_ = d.docker.Close()
	d.removePID()
	d.wg.Wait()
	d.logger.Info("daemon stopped")
	return nil
}

// RecentEvents returns up to n recent events in newest-first order.
func (d *Daemon) RecentEvents(n int) []*event.Event {
	return d.history.Recent(n)
}

func (d *Daemon) collectHistoryLoop() {
	defer d.wg.Done()
	for {
		select {
		case <-d.runCtx.Done():
			return
		case ev, ok := <-d.engine.GetEventChannel():
			if !ok {
				return
			}
			d.history.Add(ev)
		}
	}
}

func (d *Daemon) writePID() error {
	if strings.TrimSpace(d.pidPath) == "" {
		return nil
	}
	return os.WriteFile(d.pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
}

func (d *Daemon) removePID() {
	if strings.TrimSpace(d.pidPath) == "" {
		return
	}
	_ = os.Remove(d.pidPath)
}

func (d *Daemon) setStopped() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = false
	if d.cancel != nil {
		d.cancel()
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
