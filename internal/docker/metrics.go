package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
)

// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	ContainerID   string
	CPUPercent    float64
	MemoryUsage   int64
	MemoryLimit   int64
	MemoryPercent float64
	NetworkRx     int64
	NetworkTx     int64
	BlockRead     int64
	BlockWrite    int64
	Timestamp     time.Time
}

// GetContainerStats retrieves current statistics for a container
// This is a snapshot, not a stream
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := c.apiClient.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, HandleAPIError(err)
	}
	defer stats.Body.Close()

	var v container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// Calculate CPU percentage
	// CPU usage is calculated as: (cpuDelta / systemDelta) * number of CPUs * 100
	cpuPercent := calculateCPUPercent(&v)

	// Memory usage
	memoryUsage := int64(v.MemoryStats.Usage)
	memoryLimit := int64(v.MemoryStats.Limit)
	memoryPercent := 0.0
	if memoryLimit > 0 {
		memoryPercent = float64(memoryUsage) / float64(memoryLimit) * 100.0
	}

	// Network statistics
	var networkRx, networkTx int64
	if v.Networks != nil {
		for _, network := range v.Networks {
			networkRx += int64(network.RxBytes)
			networkTx += int64(network.TxBytes)
		}
	}

	// Block I/O statistics
	var blockRead, blockWrite int64
	if v.BlkioStats.IoServiceBytesRecursive != nil {
		for _, entry := range v.BlkioStats.IoServiceBytesRecursive {
			switch entry.Op {
			case "Read":
				blockRead += int64(entry.Value)
			case "Write":
				blockWrite += int64(entry.Value)
			}
		}
	}

	return &ContainerStats{
		ContainerID:   containerID,
		CPUPercent:    cpuPercent,
		MemoryUsage:   memoryUsage,
		MemoryLimit:   memoryLimit,
		MemoryPercent: memoryPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
		BlockRead:     blockRead,
		BlockWrite:    blockWrite,
		Timestamp:     time.Now(),
	}, nil
}

// calculateCPUPercent calculates CPU usage percentage
// This is a simplified calculation; actual implementation may need to handle
// multiple CPU cores and different calculation methods for different platforms
func calculateCPUPercent(v *container.StatsResponse) float64 {
	// Get CPU delta
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)

	// Get system delta
	systemDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)

	// Number of CPUs
	numCPU := float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	if numCPU == 0 {
		numCPU = 1.0
	}

	// Calculate percentage
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		return (cpuDelta / systemDelta) * numCPU * 100.0
	}

	return 0.0
}

// GetMemoryUsage retrieves memory usage information for a container
// Returns RSS (Resident Set Size) if available, otherwise returns total usage
func (c *Client) GetMemoryUsage(ctx context.Context, containerID string) (int64, int64, error) {
	stats, err := c.GetContainerStats(ctx, containerID)
	if err != nil {
		return 0, 0, err
	}

	// Note: InspectResponse doesn't have MemoryStats field directly
	// We'll use the stats from GetContainerStats instead

	return stats.MemoryUsage, stats.MemoryLimit, nil
}

// GetNetworkStats retrieves network statistics for a container
// Returns total received ad transmitted bytes across all network interfaces
func (c *Client) GetNetworkStats(ctx context.Context, containerID string) (int64, int64, error) {
	stats, err := c.GetContainerStats(ctx, containerID)
	if err != nil {
		return 0, 0, err
	}

	return stats.NetworkRx, stats.NetworkTx, nil
}

// VolumeMount represents a volume mount in a container
type VolumeMount struct {
	Name   		  string
	Source			string
	Destination string
	Driver      string
}

// GetVolumeUsage retrieves volume usage information for a container
// Note: This is a placeholder for MCP. Actual volume usage calculation
// requires access to the host filesystem, which is complex on macOS with Docker Desktop
// For MVP, we return the list of mounted volumes
func (c *Client) GetVolumeUsage(ctx context.Context, containerID string) ([]VolumeMount, error) {
	container, err := c.GetContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	result := make([]VolumeMount, 0, len(container.Mounts))
	for _, mount := range container.Mounts {
		if mount.Type == "volume" {
			result = append(result, VolumeMount{
				Name:  			 mount.Name,
				Source: 		 mount.Source,
				Destination: mount.Destination,
				Driver:      mount.Driver,
			})
		}
	}

	return result, nil
}
