package daemon

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"docksphinx/internal/config"
	"docksphinx/internal/docker"
	internalgrpc "docksphinx/internal/grpc"
	"docksphinx/internal/monitor"
)

// Daemon runs monitor engine and gRPC server.
type Daemon struct {
	cfg     *config.Config
	logger  *slog.Logger
	logSink io.Closer

	dockerClient *docker.Client
	engine       *monitor.Engine
	grpcServer   *internalgrpc.Server

	mu      sync.Mutex
	running bool
}

func New(cfg *config.Config) (*Daemon, error) {
	if cfg == nil {
		cfg = config.Default()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger, sink, err := newLogger(cfg)
	if err != nil {
		return nil, err
	}

	client, err := docker.NewClient()
	if err != nil {
		if sink != nil {
			_ = sink.Close()
		}
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	if err := client.Ping(context.Background()); err != nil {
		_ = client.Close()
		if sink != nil {
			_ = sink.Close()
		}
		return nil, fmt.Errorf("ping docker daemon: %w", err)
	}

	engineCfg := cfg.EngineConfig()
	engineCfg.EventHistoryMax = cfg.Event.MaxHistory
	engine, err := monitor.NewEngine(engineCfg, client)
	if err != nil {
		_ = client.Close()
		if sink != nil {
			_ = sink.Close()
		}
		return nil, fmt.Errorf("create monitor engine: %w", err)
	}
	engine.SetLogger(logger)

	server, err := internalgrpc.NewServer(&internalgrpc.ServerOptions{
		Address:          cfg.GRPC.Address,
		EnableReflection: cfg.GRPC.EnableReflection,
		RecentEventLimit: cfg.Event.MaxHistory,
	}, engine)
	if err != nil {
		engine.Stop()
		_ = client.Close()
		if sink != nil {
			_ = sink.Close()
		}
		return nil, fmt.Errorf("create grpc server: %w", err)
	}

	return &Daemon{
		cfg:          cfg,
		logger:       logger,
		logSink:      sink,
		dockerClient: client,
		engine:       engine,
		grpcServer:   server,
	}, nil
}

// Run starts monitor + grpc and blocks until context is canceled.
func (d *Daemon) Run(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("daemon is already running")
	}
	d.running = true
	d.mu.Unlock()

	d.logger.Info("starting docksphinxd", "grpc_address", d.cfg.GRPC.Address)
	if err := d.writePID(); err != nil {
		d.logger.Warn("failed to write pid file", "path", d.cfg.Daemon.PIDFile, "error", err)
	}

	if err := d.engine.Start(); err != nil {
		d.cleanup()
		return fmt.Errorf("start monitor engine: %w", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- d.grpcServer.Start()
	}()

	select {
	case <-ctx.Done():
		d.logger.Info("context canceled, stopping daemon")
		d.cleanup()
		return nil
	case err := <-serverErrCh:
		d.logger.Error("grpc server exited", "error", err)
		d.cleanup()
		return err
	}
}

func (d *Daemon) Stop() {
	d.cleanup()
}

func (d *Daemon) cleanup() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}
	d.running = false
	d.mu.Unlock()

	d.grpcServer.Stop()
	d.engine.Stop()
	if err := d.dockerClient.Close(); err != nil {
		d.logger.Warn("docker client close failed", "error", err)
	}
	if err := d.removePID(); err != nil {
		d.logger.Warn("failed to remove pid file", "path", d.cfg.Daemon.PIDFile, "error", err)
	}
	if d.logSink != nil {
		if err := d.logSink.Close(); err != nil {
			d.logger.Warn("log sink close failed", "error", err)
		}
	}
}

func newLogger(cfg *config.Config) (*slog.Logger, io.Closer, error) {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.Log.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	if cfg.Log.File == "" {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
		return logger, nil, nil
	}

	dir := filepath.Dir(cfg.Log.File)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create log directory %q: %w", dir, err)
	}
	f, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file %q: %w", cfg.Log.File, err)
	}
	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: level}))
	return logger, f, nil
}

func (d *Daemon) writePID() error {
	path := d.cfg.Daemon.PIDFile
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o600)
}

func (d *Daemon) removePID() error {
	path := d.cfg.Daemon.PIDFile
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
