package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"docksphinx/internal/config"
)

type countCloser struct {
	closeCount int
}

func (c *countCloser) Close() error {
	c.closeCount++
	return nil
}

func TestDaemonRunNilReceiver(t *testing.T) {
	var d *Daemon

	err := d.Run(context.Background())
	if err == nil {
		t.Fatal("expected nil daemon run to fail")
	}
	if got := err.Error(); got != "daemon is nil" {
		t.Fatalf("expected nil daemon error, got %q", got)
	}

	d.Stop() // should not panic
	if err := d.writePID(); err != nil {
		t.Fatalf("expected nil daemon writePID to be no-op, got %v", err)
	}
	if err := d.removePID(); err != nil {
		t.Fatalf("expected nil daemon removePID to be no-op, got %v", err)
	}
}

func TestDaemonRunReturnsErrorWhenUninitialized(t *testing.T) {
	d := &Daemon{}

	err := d.Run(context.Background())
	if err == nil {
		t.Fatal("expected uninitialized daemon run to fail")
	}
	if got := err.Error(); got != "daemon is not initialized" {
		t.Fatalf("expected uninitialized daemon error, got %q", got)
	}
}

func TestDaemonCleanupIsSafeForPartiallyInitializedState(t *testing.T) {
	d := &Daemon{running: true}

	d.cleanup()
	if d.running {
		t.Fatal("expected cleanup to clear running flag")
	}

	d.Stop() // should remain no-op after cleanup
}

func TestNewLoggerUsesTrimmedCaseInsensitiveLevel(t *testing.T) {
	cfg := config.Default()
	cfg.Log.Level = "  DeBuG  "
	cfg.Log.File = ""

	logger, sink, err := newLogger(cfg)
	if err != nil {
		t.Fatalf("expected newLogger to accept trimmed/case-mixed level: %v", err)
	}
	if sink != nil {
		t.Fatalf("expected stdout logger to have nil sink, got %#v", sink)
	}
	if !logger.Handler().Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("expected logger to enable debug level for trimmed/case-mixed debug input")
	}
}

func TestNewLoggerNilConfigUsesDefaults(t *testing.T) {
	logger, sink, err := newLogger(nil)
	if err != nil {
		t.Fatalf("expected newLogger(nil) to use defaults, got error: %v", err)
	}
	if sink != nil {
		t.Fatalf("expected default logger sink to be nil (stdout), got %#v", sink)
	}
	if !logger.Handler().Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected default logger to enable info level")
	}
}

func TestNewLoggerTrimsLogFilePath(t *testing.T) {
	cfg := config.Default()
	cfg.Log.Level = "info"
	trimmedPath := filepath.Join(t.TempDir(), "daemon.log")
	cfg.Log.File = "  " + trimmedPath + "  "

	_, sink, err := newLogger(cfg)
	if err != nil {
		t.Fatalf("expected newLogger to accept trimmed log file path: %v", err)
	}
	if sink == nil {
		t.Fatal("expected file logger to return closable sink")
	}
	if err := sink.Close(); err != nil {
		t.Fatalf("expected sink close to succeed: %v", err)
	}

	if _, err := os.Stat(trimmedPath); err != nil {
		t.Fatalf("expected log file to be created at trimmed path %q: %v", trimmedPath, err)
	}
}

func TestDaemonPIDHelpersTrimConfiguredPath(t *testing.T) {
	trimmedPath := filepath.Join(t.TempDir(), "docksphinxd.pid")
	d := &Daemon{
		cfg: &config.Config{
			Daemon: config.DaemonConfig{
				PIDFile: "  " + trimmedPath + "  ",
			},
		},
	}

	if err := d.writePID(); err != nil {
		t.Fatalf("expected writePID to succeed with trim-capable path handling: %v", err)
	}
	if _, err := os.Stat(trimmedPath); err != nil {
		t.Fatalf("expected pid file at trimmed path %q: %v", trimmedPath, err)
	}

	if err := d.removePID(); err != nil {
		t.Fatalf("expected removePID to succeed with trim-capable path handling: %v", err)
	}
	if _, err := os.Stat(trimmedPath); !os.IsNotExist(err) {
		t.Fatalf("expected pid file to be removed, stat err=%v", err)
	}
}

func TestDaemonCleanupRunsEvenWhenNotRunningAndIsIdempotent(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "docksphinxd.pid")
	if err := os.WriteFile(pidPath, []byte("1234\n"), 0o600); err != nil {
		t.Fatalf("failed to prepare pid file: %v", err)
	}

	sink := &countCloser{}
	d := &Daemon{
		cfg: &config.Config{
			Daemon: config.DaemonConfig{
				PIDFile: pidPath,
			},
		},
		logSink: sink,
		running: false,
	}

	d.cleanup()
	if d.running {
		t.Fatal("expected cleanup to keep running=false")
	}
	if sink.closeCount != 1 {
		t.Fatalf("expected log sink to be closed once, got %d", sink.closeCount)
	}
	if d.logSink != nil {
		t.Fatal("expected log sink to be cleared after cleanup")
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("expected pid file to be removed by cleanup, stat err=%v", err)
	}

	d.cleanup()
	if sink.closeCount != 1 {
		t.Fatalf("expected second cleanup not to re-close already-cleared sink, got %d", sink.closeCount)
	}
}
