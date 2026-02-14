package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
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
	events := makeEvents(3)
	if got := selectRecentEvents(events, 10); len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
	events = makeEvents(5)
	got := selectRecentEvents(events, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
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
