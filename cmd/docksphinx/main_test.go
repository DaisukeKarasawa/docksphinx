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
