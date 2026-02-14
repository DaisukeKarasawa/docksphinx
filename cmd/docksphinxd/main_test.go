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

func TestWaitForProcessExitUnexpectedCheckerError(t *testing.T) {
	pid := 4567
	checker := func(_ int) error { return errors.New("checker boom") }

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitForProcessExit(ctx, pid, 10*time.Millisecond, checker)
	if err == nil {
		t.Fatal("expected checker error to be returned")
	}
	if !strings.Contains(err.Error(), "failed to check process") {
		t.Fatalf("expected wrapped checker error, got: %v", err)
	}
}

func TestWaitForProcessExitImmediateContextCancel(t *testing.T) {
	pid := 5678
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForProcessExit(ctx, pid, 10*time.Millisecond, func(_ int) error { return nil })
	if err == nil {
		t.Fatal("expected error when context is already canceled")
	}
	if !strings.Contains(err.Error(), "did not stop") {
		t.Fatalf("expected timeout/cancel message, got: %v", err)
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

func TestDescribePIDStatus(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "docksphinxd.pid")
	if err := os.WriteFile(pidPath, []byte("123\n"), 0o600); err != nil {
		t.Fatalf("failed to create pid file: %v", err)
	}

	status, stale, err := describePIDStatus(pidPath, func(_ int) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Fatal("expected alive process to not be stale")
	}
	if !strings.Contains(status, "(alive)") {
		t.Fatalf("expected alive status, got %q", status)
	}

	status, stale, err = describePIDStatus(pidPath, func(_ int) error { return syscall.ESRCH })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stale {
		t.Fatal("expected stale=true for ESRCH")
	}
	if !strings.Contains(status, "(stale)") {
		t.Fatalf("expected stale status, got %q", status)
	}

	status, stale, err = describePIDStatus(filepath.Join(dir, "missing.pid"), func(_ int) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Fatal("expected stale=false for missing pid")
	}
	if status != "pid: not found" {
		t.Fatalf("expected not found status, got %q", status)
	}

	status, stale, err = describePIDStatus(pidPath, func(_ int) error { return syscall.EPERM })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Fatal("expected stale=false for EPERM")
	}
	if !strings.Contains(status, "(permission denied)") {
		t.Fatalf("expected permission denied status, got %q", status)
	}

	status, stale, err = describePIDStatus(pidPath, func(_ int) error { return errors.New("boom") })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Fatal("expected stale=false for unknown checker error")
	}
	if !strings.Contains(status, "(unknown)") {
		t.Fatalf("expected unknown status, got %q", status)
	}
}

func TestDescribePIDStatusInvalidPIDReturnsError(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "docksphinxd.pid")
	if err := os.WriteFile(pidPath, []byte("not-a-number\n"), 0o600); err != nil {
		t.Fatalf("failed to create pid file: %v", err)
	}

	_, _, err := describePIDStatus(pidPath, func(_ int) error { return nil })
	if err == nil {
		t.Fatal("expected invalid pid to return error")
	}
	if !strings.Contains(err.Error(), "invalid pid") {
		t.Fatalf("expected invalid pid error, got: %v", err)
	}
}

func TestInspectPID(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "docksphinxd.pid")
	if err := os.WriteFile(pidPath, []byte("123\n"), 0o600); err != nil {
		t.Fatalf("failed to create pid file: %v", err)
	}

	pid, running, stale, err := inspectPID(pidPath, func(_ int) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != 123 || !running || stale {
		t.Fatalf("unexpected inspect result: pid=%d running=%t stale=%t", pid, running, stale)
	}

	pid, running, stale, err = inspectPID(pidPath, func(_ int) error { return syscall.ESRCH })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != 123 || running || !stale {
		t.Fatalf("unexpected inspect stale result: pid=%d running=%t stale=%t", pid, running, stale)
	}

	pid, running, stale, err = inspectPID(filepath.Join(dir, "missing.pid"), func(_ int) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != 0 || running || stale {
		t.Fatalf("expected missing pid file to be treated as not running, got pid=%d running=%t stale=%t", pid, running, stale)
	}

	_, _, _, err = inspectPID(pidPath, func(_ int) error { return syscall.EPERM })
	if err == nil {
		t.Fatal("expected permission denied error")
	}
	_, _, _, err = inspectPID(pidPath, func(_ int) error { return errors.New("boom") })
	if err == nil {
		t.Fatal("expected unknown checker error")
	}
}

func TestInspectPIDInvalidPIDFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "docksphinxd.pid")
	if err := os.WriteFile(pidPath, []byte("invalid\n"), 0o600); err != nil {
		t.Fatalf("failed to create pid file: %v", err)
	}

	_, _, _, err := inspectPID(pidPath, func(_ int) error { return nil })
	if err == nil {
		t.Fatal("expected invalid pid file to return error")
	}
	if !strings.Contains(err.Error(), "invalid pid") {
		t.Fatalf("expected invalid pid error, got: %v", err)
	}
}

func TestMarkAlreadyReported(t *testing.T) {
	t.Run("wraps original error", func(t *testing.T) {
		orig := errors.New("boom")
		err := markAlreadyReported(orig)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if !errors.Is(err, ErrAlreadyReported) {
			t.Fatalf("expected ErrAlreadyReported wrapper, got: %v", err)
		}
		if !errors.Is(err, orig) {
			t.Fatalf("expected original error to be preserved, got: %v", err)
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected wrapped original message, got: %v", err)
		}
	})

	t.Run("nil input still returns sentinel", func(t *testing.T) {
		err := markAlreadyReported(nil)
		if err == nil {
			t.Fatal("expected non-nil sentinel error")
		}
		if !errors.Is(err, ErrAlreadyReported) {
			t.Fatalf("expected ErrAlreadyReported for nil input, got: %v", err)
		}
	})
}
