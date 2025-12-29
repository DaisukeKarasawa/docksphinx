//go:build integration
package docker

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test ping
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping Docker daemon: %v", err)
	}
}

func TestListContainers(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	containers, err := client.ListContainers(ctx, ListContainersOptions{
		All: true,
	})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	t.Logf("Found %d containers", len(containers))
	for _, container := range containers {
		t.Logf("Container: %s (%s) - %s", container.Name, container.ID[:12], container.Status)
	}
}

func TestGetContainerStats(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get a running container
	containers, err := client.ListContainers(ctx, ListContainersOptions{
		All: false, // Only running
	})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	if len(containers) == 0 {
		t.Skip("No running containers found, skipping stats test")
	}

	// Get stats for the first container
	stats, err := client.GetContainerStats(ctx, containers[0].ID)
	if err != nil {
		t.Fatalf("Failed to get container stats: %v", err)
	}

	t.Logf("Container %s stats:", containers[0].Name)
	t.Logf("  CPU: %.2f%%", stats.CPUPercent)
	t.Logf("  Memory: %d / %d bytes (%.2f%%)",
		stats.MemoryUsage, stats.MemoryLimit, stats.MemoryPercent)
	t.Logf("  Network: RX %d bytes, TX %d bytes", stats.NetworkRx, stats.NetworkTx)
}
