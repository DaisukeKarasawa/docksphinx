package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

// Client wraps the Docker API client
type Client struct {
	apiClient *client.Client
}

// NewClient creates a new Docker API client
// It connects to the Docker daemon using the default socket path
// On macOS with Docker Desktop, this is typically /var/run/docker.sock
func NewClient() (*Client, error) {
	apiClient, err := client.NewClientWithOpts(
		client.FromEnv,                     // Use environment variables (DOCKER_HOST, etc...)
		client.WithAPIVersionNegotiation(), // Automatically negotiate API version
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		apiClient: apiClient,
	}, nil
}

// Ping checks if the Docker daemon is accessible
// This is useful for verifying the connection before making other API calls
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := c.apiClient.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Docker daemon: %w", err)
	}

	return nil
}

// Close closes the Docker API client connection
// Should be called when the client is no longer needed
func (c *Client) Close() error {
	return c.apiClient.Close()
}

// GetAPIClient returns the underlying DockerAPI client
// This is useful when you need direct access to the API client
func (c *Client) GetAPIClient() *client.Client {
	return c.apiClient
}
