package docker

import (
	"context"

	"github.com/docker/docker/api/types/volume"
)

// Volume represents a Docker volume with its basic information
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
	RefCount   int64
	UsageNote  string
}

// ListVolumes lists all Docker volumes
func (c *Client) ListVolumes(ctx context.Context) ([]Volume, error) {
	volumes, err := c.apiClient.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	result := make([]Volume, 0, len(volumes.Volumes))
	for _, vol := range volumes.Volumes {
		result = append(result, Volume{
			Name:       vol.Name,
			Driver:     vol.Driver,
			Mountpoint: vol.Mountpoint,
			Labels:     vol.Labels,
			UsageNote:  "metadata-only (size unavailable via Docker API)",
		})
	}

	return result, nil
}

// GetVolume retrieves detailed information about a specific volume
func (c *Client) GetVolume(ctx context.Context, volumeName string) (*volume.Volume, error) {
	vol, err := c.apiClient.VolumeInspect(ctx, volumeName)
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &vol, nil
}
