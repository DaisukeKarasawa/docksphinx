package docker

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

// Network represents a Docker network with its basic information
type Network struct {
	ID             string
	Name           string
	Driver         string
	Scope          string
	Internal       bool
	Labels         map[string]string
	ContainerCount int
}

// ListNetworks lists all Docker networks
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	apiClient, err := c.getAPIClient()
	if err != nil {
		return nil, err
	}
	networks, err := apiClient.NetworkList(normalizeContext(ctx), network.ListOptions{})
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
	apiClient, err := c.getAPIClient()
	if err != nil {
		return nil, err
	}
	network, err := apiClient.NetworkInspect(normalizeContext(ctx), networkID, network.InspectOptions{
		Verbose: true, // Include detailed information
	})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &network, nil
}
