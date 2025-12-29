package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
)

// Image represents a Docker image with its basic information
type Image struct {
	ID				  string
	Repository  string
	Tag 	 		  string
	Size  		  int64
	Created     int64
	VirtualSize int64
}

// ListImages lists all Docker images
// This is used for filtering by image name
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	images, err := c.apiClient.ImageList(ctx, image.ListOptions{
		All: true, // Include intermediate images
	})
	if err != nil {
		return nil, HandleAPIError(err)
	}

	result := make([]Image, 0, len(images))
	for _, img := range images {
		// Images can have multiple tags, we'll use the first one
		repository := "<none>"
		tag := "<none>"
		if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
			// Parse "repository:tag" format
			parts := splitImageTag(img.RepoTags[0])
			repository = parts[0]
			tag = parts[1]
		}

		result = append(result, Image{
			ID: 				 img.ID,
			Repository:  repository,
			Tag:   			 tag,
			Size:        img.Size,
			Created:     img.Created,
			VirtualSize: img.VirtualSize,
		})
	}

	return result, nil
}

// GetImage retrieves detailed information about a specific image
func (c *Client) GetImage(ctx context.Context, imageID string) (*types.ImageInspect, error) {
	image, _, err := c.apiClient.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &image, nil
}

// splitImageTag splits "repository:tag" into ["repository", "tag"]
func splitImageTag(imageTag string) []string {
	parts := strings.SplitN(imageTag, ":", 2)
	if len(parts) == 1 {
		return []string{parts[0], "latest"}
	}
	return parts
}
