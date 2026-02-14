package main

import (
	"context"
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
