package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/image"
)

// Image represents a Docker image with its basic information
type Image struct {
	ID          string
	Repository  string
	Tag         string
	Size        int64
	Created     int64
	VirtualSize int64
}

// ListImages lists all Docker images including intermediate images.
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
			ID:          img.ID,
			Repository:  repository,
			Tag:         tag,
			Size:        img.Size,
			Created:     img.Created,
			VirtualSize: img.VirtualSize,
		})
	}

	return result, nil
}

// GetImage retrieves detailed information about a specific image
func (c *Client) GetImage(ctx context.Context, imageID string) (*image.InspectResponse, error) {
	imageInspect, err := c.apiClient.ImageInspect(ctx, imageID)
	if err != nil {
		return nil, HandleAPIError(err)
	}

	return &imageInspect, nil
}

// splitImageTag splits "repository:tag" into ["repository", "tag"]
// Handles registry ports correctly (e.g., "localhost:5000/myimage:v1")
// The tag separator is the last ':' after the last '/' (if any)
func splitImageTag(imageTag string) []string {
	// Find the last '/' to separate registry from repository
	lastSlash := strings.LastIndex(imageTag, "/")

	// If there's a '/', look for ':' after it (this is the tag separator)
	// Otherwise, look for the last ':' in the entire string
	var tagIndex int
	if lastSlash >= 0 {
		// Look for ':' after the last '/'
		tagIndex = strings.LastIndex(imageTag[lastSlash:], ":")
		if tagIndex >= 0 {
			tagIndex += lastSlash // Adjust index to be relative to start of string
		}
	} else {
		// No '/', so the last ':' is the tag separator
		tagIndex = strings.LastIndex(imageTag, ":")
	}

	if tagIndex < 0 {
		// No tag found, use "latest" as default
		return []string{imageTag, "latest"}
	}

	// Split at the tag separator
	return []string{imageTag[:tagIndex], imageTag[tagIndex+1:]}
}
