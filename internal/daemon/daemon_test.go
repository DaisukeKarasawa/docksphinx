package daemon

import (
	"context"
	"testing"
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
