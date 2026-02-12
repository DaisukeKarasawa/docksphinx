package docker

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

// Network represents a Docker network with its basic information
type Network struct {
	ID       string
	Name     string
	Driver   string
	Scope    string
	Internal bool
	Labels   map[string]string
}

// ListNetworks lists all Docker networks
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	networks, err := c.apiClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	result := make([]Network, 0, len(networks))
	for _, net := range networks {
		result = append(result, Network{
			ID:       net.ID,
			Name:     net.Name,
			Driver:   net.Driver,
			Scope:    net.Scope,
			Internal: net.Internal,
			Labels:   net.Labels,
		})
	}

	return result, nil
}

// GetNetwork retrieves detailed information about a specific network
func (c *Client) GetNetwork(ctx context.Context, networkID string) (*network.Inspect, error) {
	network, err := c.apiClient.NetworkInspect(ctx, networkID, network.InspectOptions{
		Verbose: true, // Include detailed information
	})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &network, nil
}
