package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"docksphinx/internal/config"
)

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
