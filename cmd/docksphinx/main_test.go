package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	dcfg "docksphinx/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFormatUptimeOrNA(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.ContainerInfo
		want string
	}{
		{
			name: "nil container",
			in:   nil,
			want: "N/A",
		},
		{
			name: "missing started and uptime",
			in: &pb.ContainerInfo{
				StartedAtUnix: 0,
				UptimeSeconds: 0,
			},
			want: "N/A",
		},
		{
			name: "has startedAt",
			in: &pb.ContainerInfo{
				StartedAtUnix: 100,
				UptimeSeconds: 0,
			},
			want: "0",
		},
		{
			name: "has uptime only",
			in: &pb.ContainerInfo{
				StartedAtUnix: 0,
				UptimeSeconds: 42,
			},
			want: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUptimeOrNA(tt.in)
			if got != tt.want {
				t.Fatalf("formatUptimeOrNA()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectRecentEvents(t *testing.T) {
	makeEvents := func(n int) []*pb.Event {
		out := make([]*pb.Event, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, &pb.Event{Id: string(rune('a' + i))})
		}
		return out
	}

	if got := selectRecentEvents(nil, 10); got != nil {
		t.Fatalf("expected nil for empty input, got %#v", got)
	}
	if got := selectRecentEvents(makeEvents(3), 0); got != nil {
		t.Fatalf("expected nil for non-positive limit, got %#v", got)
	}
	if got := selectRecentEvents(makeEvents(3), -1); got != nil {
		t.Fatalf("expected nil for negative limit, got %#v", got)
	}
	events := makeEvents(3)
	if got := selectRecentEvents(events, 10); len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
	events = makeEvents(5)
	got := selectRecentEvents(events, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}

	unsorted := []*pb.Event{
		{Id: "b", TimestampUnix: 100},
		{Id: "a", TimestampUnix: 200},
		{Id: "c", TimestampUnix: 150},
	}
	before := []string{unsorted[0].GetId(), unsorted[1].GetId(), unsorted[2].GetId()}
	got = selectRecentEvents(unsorted, 3)
	if got[0].GetId() != "a" || got[1].GetId() != "c" || got[2].GetId() != "b" {
		t.Fatalf("expected timestamp-desc order [a c b], got [%s %s %s]", got[0].GetId(), got[1].GetId(), got[2].GetId())
	}
	after := []string{unsorted[0].GetId(), unsorted[1].GetId(), unsorted[2].GetId()}
	if before[0] != after[0] || before[1] != after[1] || before[2] != after[2] {
		t.Fatalf("expected input ordering to remain unchanged, before=%v after=%v", before, after)
	}

	sameTimestamp := []*pb.Event{
		{Id: "c", TimestampUnix: 300},
		{Id: "a", TimestampUnix: 300},
		{Id: "b", TimestampUnix: 300},
	}
	got = selectRecentEvents(sameTimestamp, 3)
	if got[0].GetId() != "a" || got[1].GetId() != "b" || got[2].GetId() != "c" {
		t.Fatalf("expected id-asc tie-break [a b c], got [%s %s %s]", got[0].GetId(), got[1].GetId(), got[2].GetId())
	}

	withNil := []*pb.Event{
		nil,
		{Id: "z", TimestampUnix: 10},
		nil,
		{Id: "a", TimestampUnix: 20},
	}
	got = selectRecentEvents(withNil, 10)
	if len(got) != 2 {
		t.Fatalf("expected nil events to be filtered out, got len=%d", len(got))
	}
	if got[0].GetId() != "a" || got[1].GetId() != "z" {
		t.Fatalf("expected filtered/sorted order [a z], got [%s %s]", got[0].GetId(), got[1].GetId())
	}

	withNilAndCap := []*pb.Event{
		nil,
		{Id: "z", TimestampUnix: 10},
		{Id: "a", TimestampUnix: 20},
		{Id: "m", TimestampUnix: 15},
	}
	got = selectRecentEvents(withNilAndCap, 2)
	if len(got) != 2 {
		t.Fatalf("expected capped result length=2, got len=%d", len(got))
	}
	if got[0].GetId() != "a" || got[1].GetId() != "m" {
		t.Fatalf("expected filtered/sorted/capped order [a m], got [%s %s]", got[0].GetId(), got[1].GetId())
	}

	allNil := []*pb.Event{nil, nil}
	if got := selectRecentEvents(allNil, 5); got != nil {
		t.Fatalf("expected nil when all events are nil, got %#v", got)
	}
}

func TestPrintSnapshotToIncludesSectionsAndNA(t *testing.T) {
	snapshot := &pb.Snapshot{
		AtUnix: time.Now().Unix(),
		Containers: []*pb.ContainerInfo{
			{
				ContainerId:   "abcdef1234567890",
				ContainerName: "web",
				ImageName:     "nginx:latest",
				State:         "running",
				UptimeSeconds: 0,
			},
		},
		Metrics: map[string]*pb.ContainerMetrics{
			// intentionally empty for the container -> expect N/A rendering
		},
		RecentEvents: []*pb.Event{
			{
				Type:          "restarted",
				TimestampUnix: time.Now().Unix(),
				ContainerName: "web",
				Message:       "container restarted",
			},
		},
		Groups: []*pb.ComposeGroup{
			{
				Project:      "proj",
				Service:      "web",
				ContainerIds: []string{"abcdef1234567890"},
				NetworkNames: []string{"proj_default"},
			},
		},
		Networks: []*pb.NetworkInfo{
			{Name: "proj_default", Driver: "bridge", Scope: "local", ContainerCount: 1},
		},
		Volumes: []*pb.VolumeInfo{
			{Name: "data", Driver: "local", RefCount: 1, UsageNote: "metadata-only"},
		},
		Images: []*pb.ImageInfo{
			{Repository: "nginx", Tag: "latest", Size: 1234, CreatedUnix: time.Now().Unix()},
		},
	}

	var buf bytes.Buffer
	printSnapshotTo(snapshot, &buf)
	out := buf.String()

	required := []string{
		"CONTAINER ID",
		"RECENT EVENTS",
		"GROUPS",
		"NETWORKS",
		"VOLUMES",
		"IMAGES",
		"N/A",
	}
	for _, want := range required {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestShouldReconnectTail(t *testing.T) {
	activeCtx := context.Background()
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		ctx  context.Context
		err  error
		want bool
	}{
		{name: "nil error", ctx: activeCtx, err: nil, want: false},
		{name: "context canceled error", ctx: activeCtx, err: context.Canceled, want: false},
		{name: "ctx already canceled", ctx: canceledCtx, err: io.EOF, want: false},
		{name: "io eof should reconnect", ctx: activeCtx, err: io.EOF, want: true},
		{name: "other error should reconnect", ctx: activeCtx, err: errors.New("boom"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldReconnectTail(tt.ctx, tt.err)
			if got != tt.want {
				t.Fatalf("shouldReconnectTail()=%t, want %t", got, tt.want)
			}
		})
	}
}

func TestLogTailRetry(t *testing.T) {
	t.Run("writes expected format", func(t *testing.T) {
		var buf bytes.Buffer
		logTailRetry(&buf, "connect", errors.New("boom"), 2*time.Second)

		got := buf.String()
		want := "tail connect failed: boom (retrying in 2s)\n"
		if got != want {
			t.Fatalf("unexpected log format:\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("nil writer is no-op", func(t *testing.T) {
		logTailRetry(nil, "subscribe", errors.New("boom"), 500*time.Millisecond)
	})
}

func TestLogTailStreamReconnect(t *testing.T) {
	t.Run("writes expected format", func(t *testing.T) {
		var buf bytes.Buffer
		logTailStreamReconnect(&buf, io.EOF, time.Second)

		got := buf.String()
		want := "tail stream disconnected: EOF (retrying in 1s)\n"
		if got != want {
			t.Fatalf("unexpected stream reconnect log format:\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("nil writer is no-op", func(t *testing.T) {
		logTailStreamReconnect(nil, io.EOF, time.Second)
	})
}

func TestWrapDaemonError(t *testing.T) {
	t.Run("connection refused suggests daemon start", func(t *testing.T) {
		refused := &net.OpError{Err: syscall.ECONNREFUSED}
		err := wrapDaemonError("connect", "127.0.0.1:50051", refused)
		if err == nil {
			t.Fatal("expected wrapped error")
		}
		msg := err.Error()
		if !strings.Contains(msg, "start daemon with `docksphinxd start`") {
			t.Fatalf("expected start hint in message, got: %s", msg)
		}
	})

	t.Run("deadline exceeded suggests daemon start", func(t *testing.T) {
		err := wrapDaemonError("snapshot", "127.0.0.1:50051", context.DeadlineExceeded)
		if err == nil {
			t.Fatal("expected wrapped error")
		}
		msg := err.Error()
		if !strings.Contains(msg, "start daemon with `docksphinxd start`") {
			t.Fatalf("expected start hint in message, got: %s", msg)
		}
	})

	t.Run("grpc unavailable suggests status check", func(t *testing.T) {
		stErr := status.Error(codes.Unavailable, "transport is closing")
		err := wrapDaemonError("tail", "127.0.0.1:50051", stErr)
		if err == nil {
			t.Fatal("expected wrapped error")
		}
		msg := err.Error()
		if !strings.Contains(msg, "check daemon status") {
			t.Fatalf("expected status hint in message, got: %s", msg)
		}
	})
}

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{addr: "127.0.0.1:50051", want: true},
		{addr: "localhost:50051", want: true},
		{addr: "LOCALHOST:50051", want: true},
		{addr: "[::1]:50051", want: true},
		{addr: "0.0.0.0:50051", want: false},
		{addr: "192.168.1.10:50051", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := isLoopback(tt.addr)
			if got != tt.want {
				t.Fatalf("isLoopback(%q)=%t, want %t", tt.addr, got, tt.want)
			}
		})
	}
}

func TestWarnInsecure(t *testing.T) {
	t.Run("loopback does not warn", func(t *testing.T) {
		orig := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe setup failed: %v", err)
		}
		os.Stderr = w
		warnInsecure("127.0.0.1:50051", false)
		_ = w.Close()
		os.Stderr = orig

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if got := buf.String(); got != "" {
			t.Fatalf("expected no warning, got: %q", got)
		}
	})

	t.Run("uppercase localhost does not warn", func(t *testing.T) {
		orig := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe setup failed: %v", err)
		}
		os.Stderr = w
		warnInsecure("LOCALHOST:50051", false)
		_ = w.Close()
		os.Stderr = orig

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if got := buf.String(); got != "" {
			t.Fatalf("expected no warning for uppercase localhost, got: %q", got)
		}
	})

	t.Run("non-loopback warns unless insecure flag set", func(t *testing.T) {
		orig := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe setup failed: %v", err)
		}
		os.Stderr = w
		warnInsecure("10.0.0.5:50051", false)
		_ = w.Close()
		os.Stderr = orig

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		got := buf.String()
		if !strings.Contains(got, "WARNING: connecting to 10.0.0.5:50051 over plaintext") {
			t.Fatalf("expected warning message, got: %q", got)
		}

		r2, w2, err := os.Pipe()
		if err != nil {
			t.Fatalf("second pipe setup failed: %v", err)
		}
		os.Stderr = w2
		warnInsecure("10.0.0.5:50051", true)
		_ = w2.Close()
		os.Stderr = orig

		var buf2 bytes.Buffer
		_, _ = io.Copy(&buf2, r2)
		if got2 := buf2.String(); got2 != "" {
			t.Fatalf("expected no warning when insecure=true, got: %q", got2)
		}
	})
}

func TestIsConnectionRefused(t *testing.T) {
	t.Run("direct errno", func(t *testing.T) {
		if !isConnectionRefused(syscall.ECONNREFUSED) {
			t.Fatal("expected direct ECONNREFUSED to be detected")
		}
	})

	t.Run("nested net op error", func(t *testing.T) {
		err := &net.OpError{Err: syscall.ECONNREFUSED}
		if !isConnectionRefused(err) {
			t.Fatal("expected nested ECONNREFUSED to be detected")
		}
	})

	t.Run("non refused error", func(t *testing.T) {
		err := &net.OpError{Err: syscall.ETIMEDOUT}
		if isConnectionRefused(err) {
			t.Fatal("did not expect ETIMEDOUT to be detected as refused")
		}
	})
}

func TestResolveAddress(t *testing.T) {
	t.Run("addr override takes precedence even if config path is invalid", func(t *testing.T) {
		got, err := resolveAddress("/definitely/not/found.yaml", " 127.0.0.1:12345 ")
		if err != nil {
			t.Fatalf("expected override to bypass config loading, got error: %v", err)
		}
		if got != "127.0.0.1:12345" {
			t.Fatalf("expected trimmed override address, got %q", got)
		}
	})

	t.Run("load address from config when override is empty", func(t *testing.T) {
		cfg := dcfg.Default()
		cfg.GRPC.Address = "127.0.0.1:54321"

		path := filepath.Join(t.TempDir(), "docksphinx.yaml")
		if err := cfg.Save(path); err != nil {
			t.Fatalf("failed to save test config: %v", err)
		}

		got, err := resolveAddress(path, "")
		if err != nil {
			t.Fatalf("expected config load success, got error: %v", err)
		}
		if got != "127.0.0.1:54321" {
			t.Fatalf("expected config grpc address, got %q", got)
		}
	})
}

func TestFormatDateOrNA(t *testing.T) {
	t.Run("missing timestamp returns N/A", func(t *testing.T) {
		if got := formatDateOrNA(0); got != "N/A" {
			t.Fatalf("expected N/A for zero timestamp, got %q", got)
		}
	})

	t.Run("valid timestamp returns formatted date", func(t *testing.T) {
		unix := time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC).Unix()
		if got := formatDateOrNA(unix); got != "2026-02-14" {
			t.Fatalf("expected formatted date, got %q", got)
		}
	})
}

func TestPrintSnapshotToImageCreatedMissingRendersNA(t *testing.T) {
	snapshot := &pb.Snapshot{
		AtUnix: time.Now().Unix(),
		Images: []*pb.ImageInfo{
			{
				Repository:  "busybox",
				Tag:         "latest",
				Size:        123,
				CreatedUnix: 0,
			},
		},
	}

	var buf bytes.Buffer
	printSnapshotTo(snapshot, &buf)
	out := buf.String()
	if !strings.Contains(out, "busybox:latest\tsize=123\tcreated=N/A") {
		t.Fatalf("expected created=N/A output, got:\n%s", out)
	}
}

func TestPrintSnapshotToSortsResourceSections(t *testing.T) {
	snapshot := &pb.Snapshot{
		AtUnix: time.Now().Unix(),
		Groups: []*pb.ComposeGroup{
			{Project: "zeta", Service: "api", NetworkNames: []string{"net_b", "net_a"}},
			{Project: "alpha", Service: "web", NetworkNames: []string{"net_z", "net_x"}},
		},
		Networks: []*pb.NetworkInfo{
			{Name: "znet", Driver: "bridge", Scope: "local", ContainerCount: 1},
			{Name: "anet", Driver: "bridge", Scope: "local", ContainerCount: 2},
		},
		Volumes: []*pb.VolumeInfo{
			{Name: "zvol", Driver: "local", RefCount: 1, UsageNote: "metadata-only"},
			{Name: "avol", Driver: "local", RefCount: 2, UsageNote: "metadata-only"},
		},
		Images: []*pb.ImageInfo{
			{Repository: "zrepo", Tag: "latest", Size: 2, CreatedUnix: 1},
			{Repository: "arepo", Tag: "latest", Size: 1, CreatedUnix: 1},
		},
	}

	var buf bytes.Buffer
	printSnapshotTo(snapshot, &buf)
	out := buf.String()

	mustAppearBefore := func(first, second string) {
		t.Helper()
		i := strings.Index(out, first)
		j := strings.Index(out, second)
		if i == -1 || j == -1 {
			t.Fatalf("missing expected markers %q or %q in output:\n%s", first, second, out)
		}
		if i >= j {
			t.Fatalf("expected %q to appear before %q in output:\n%s", first, second, out)
		}
	}

	mustAppearBefore("alpha/web", "zeta/api")
	mustAppearBefore("anet\tdriver=", "znet\tdriver=")
	mustAppearBefore("avol\tdriver=", "zvol\tdriver=")
	mustAppearBefore("arepo:latest\tsize=", "zrepo:latest\tsize=")
	mustAppearBefore("net_x,net_z", "net_a,net_b")
}
