package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"syscall"
)

func TestWaitForProcessExitSuccessOnESRCH(t *testing.T) {
	pid := 1234
	calls := 0
	checker := func(_ int) error {
		calls++
		if calls < 3 {
			return nil
		}
		return syscall.ESRCH
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := waitForProcessExit(ctx, pid, 10*time.Millisecond, checker); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestWaitForProcessExitTimeout(t *testing.T) {
	pid := 2345
	checker := func(_ int) error { return nil }

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := waitForProcessExit(ctx, pid, 10*time.Millisecond, checker)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "did not stop") {
		t.Fatalf("expected timeout message, got: %v", err)
	}
}

func TestWaitForProcessExitPermissionDenied(t *testing.T) {
	pid := 3456
	checker := func(_ int) error { return syscall.EPERM }

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForProcessExit(ctx, pid, 10*time.Millisecond, checker)
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}

func TestRemovePIDFileIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docksphinxd.pid")
	if err := os.WriteFile(path, []byte("123\n"), 0o600); err != nil {
		t.Fatalf("failed to create pid file: %v", err)
	}

	if err := removePIDFileIfExists(path); err != nil {
		t.Fatalf("expected pid file removal success, got: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected pid file to be removed, stat err=%v", err)
	}
}

func TestRemovePIDFileIfExistsNoOpCases(t *testing.T) {
	if err := removePIDFileIfExists(""); err != nil {
		t.Fatalf("expected empty path no-op success, got: %v", err)
	}

	missing := filepath.Join(t.TempDir(), "missing.pid")
	if err := removePIDFileIfExists(missing); err != nil {
		t.Fatalf("expected missing file no-op success, got: %v", err)
	}
}
