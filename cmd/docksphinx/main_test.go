package main

import (
	"testing"

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
