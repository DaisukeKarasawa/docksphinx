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

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (c *Client) getAPIClient() (*client.Client, error) {
	if c == nil {
		return nil, fmt.Errorf("docker client is nil")
	}
	if c.apiClient == nil {
		return nil, fmt.Errorf("docker api client is nil")
	}
	return c.apiClient, nil
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
	apiClient, err := c.getAPIClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(normalizeContext(ctx), 5*time.Second)
	defer cancel()

	if _, err := apiClient.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Docker daemon: %w", err)
	}

	return nil
}

// Close closes the Docker API client connection
// Should be called when the client is no longer needed
func (c *Client) Close() error {
	if c == nil || c.apiClient == nil {
		return nil
	}
	return c.apiClient.Close()
}

// GetAPIClient returns the underlying DockerAPI client
// This is useful when you need direct access to the API client
func (c *Client) GetAPIClient() *client.Client {
	if c == nil {
		return nil
	}
	return c.apiClient
}
