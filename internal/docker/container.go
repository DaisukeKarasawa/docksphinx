package docker

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// Container represents a Docker container with its basic information
type Container struct {
	ID      string
	Name    string
	Image   string
	Status  string
	State   string
	Created int64
}

// ListContainersOptions specifies options for listing containers
type ListContainersOptions struct {
	// All includes stopped containers (default: false)
	All bool
	// NamePattern filters containers by name (regex pattern)
	NamePattern string
	// ImagePattern filters containers by image name (regex pattern)
	ImagePattern string
}

// ListContainers lists all containers matching the given options
// This is the main function for getting container information
func (c *Client) ListContainers(ctx context.Context, opts ListContainersOptions) ([]Container, error) {
	// Build filter arguments
	filterArgs := filters.NewArgs()

	if !opts.All {
		// Only running containers
		filterArgs.Add("status", "running")
	}

	// Get container list from Docker API
	containers, err := c.apiClient.ContainerList(ctx, container.ListOptions{
		All:     opts.All,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	// Convert to our Container type
	result := make([]Container, 0, len(containers))
	for _, container := range containers {
		if len(container.Names) == 0 {
			continue
		}

		containerInfo := Container{
			ID:      container.ID,
			Name:    strings.TrimPrefix(container.Names[0], "/"), // Remove leading "/"
			Image:   container.Image,
			Status:  container.Status,
			State:   container.State,
			Created: container.Created,
		}

		// Apply name pattern filter if specified
		if opts.NamePattern != "" {
			matched, err := regexp.MatchString(opts.NamePattern, containerInfo.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid name pattern: %w", err)
			}
			if !matched {
				continue
			}
		}

		// Apply image pattern filter if specified
		if opts.ImagePattern != "" {
			matched, err := regexp.MatchString(opts.ImagePattern, containerInfo.Image)
			if err != nil {
				return nil, fmt.Errorf("invalid image pattern: %w", err)
			}
			if !matched {
				continue
			}
		}

		result = append(result, containerInfo)
	}

	return result, nil
}

// GetContainer retrieves detailed information about a specific container
// containerID can be either the full ID or a short ID prefix
func (c *Client) GetContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	container, err := c.apiClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &container, nil
}

// ContainerDetails represents detailed container information
type ContainerDetails struct {
	ID							string
	Name						string
	Image						string
	State					  string
	Status         	string 
	Created					int64
	StartedAt       string
	FinishedAt      string
	RestartCount    int
	Platform        string
	Hostname        string
	NetworkSettings *types.NetworkSettings
	Mounts          []types.MountPoint
	Config          *container.Config
}

// GetContainerDetails retrieves detailed information about a container
// This includes network settings, mounts, configuration, etc.
func (c *Client) GetContainerDetails(ctx context.Context, containerID string) (*ContainerDetails, error) {
	container, err := c.GetContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	status := calculateStatus(container.State)
	return &ContainerDetails{
		ID: 						 container.ID,
		Name: 				   strings.TrimPrefix(container.Name, "/"),
		Image:					 container.Image,
		State:				   container.State.Status,
		Status:      	 	 status,
		Created:				 parseCreatedTime(container.Created),
		StartedAt:			 container.State.StartedAt,
		FinishedAt:			 container.State.FinishedAt,
		RestartCount:		 container.RestartCount,
		Platform:				 container.Platform,
		Hostname:  		 	 container.Config.Hostname,
		NetworkSettings: container.NetworkSettings,
		Mounts:					 container.Mounts,
		Config:					 container.Config,
	}, nil
}

// calculateStatus converts a container's state to a human-readable status string
func calculateStatus(state *types.ContainerState) string {
	switch state.Status {
	case "running":
		if state.StartedAt != "" {
			started, err := time.Parse(time.RFC3339Nano, state.StartedAt)
			if err == nil {
				duration := time.Since(started)
				return fmt.Sprintf("Up %s", formatDuration(duration))
			}
		}
		return "Up"
	case "exited":
		if state.FinishedAt != "" {
			finished, err := time.Parse(time.RFC3339Nano, state.FinishedAt)
			if err == nil {
				duration := time.Since(finished)
				return fmt.Sprintf("Exited (%d) %s ago", state.ExitCode, formatDuration(duration))
			}
		}
		return fmt.Sprintf("Exited (%d)", state.ExitCode)
	case "created":
		return "Created"
	case "paused":
		return "Paused"
	case "restarting":
		return "Restarting"
	default:
		return state.Status
	}
}

// formatDuration formats a time.Duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%d days", days)
}

// parseCreatedTime parses the Created time string and returns Unix timestamp
func parseCreatedTime(createdStr string) int64 {
	// Docker returns Created as RFC3339Nano format
	// Parse it and convert to Unix timestamp
	t, err := time.Parse(time.RFC3339Nano, createdStr)
	if err != nil {
		// Fallback: try RFC3339 format
		t, err = time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return 0
		}
	}
	return t.Unix()
}
