package docker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

// Error types for better error handling
var (
	ErrDockerNotRunning  = errors.New("Docker daemon is not running")
	ErrPermissionDenied  = errors.New("permission denied: unable to access Docker daemon")
	ErrContainerNotFound = errors.New("container not found")
	ErrImageNotFound     = errors.New("image not found")
	ErrNetworkNotFound   = errors.New("network not found")
	ErrVolumeNotFound    = errors.New("volume not found")
)

// HandleAPIError converts Docker API errors to more user-friendly errors
func HandleAPIError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific Docker client errors
	if client.IsErrNotFound(err) {
		// Try to determine what type of resource was not found
		// This is a simplified check; in practice, you might want more specific checks
		return fmt.Errorf("%w: %v", ErrContainerNotFound, err)
	}

	if client.IsErrConnectionFailed(err) {
		return fmt.Errorf("%w: %v", ErrDockerNotRunning, err)
	}

	// Check for permission errors (typically manifests as connection refused or permission denied)
	if isPermissionError(err) {
		return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
	}

	// Return the original errors if we can't categorize it
	return fmt.Errorf("Docker API error: %w", err)
}

// IsNotFoundError checks if an error indicates a resource was not found
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrContainerNotFound) ||
		errors.Is(err, ErrImageNotFound) ||
		errors.Is(err, ErrNetworkNotFound) ||
		errors.Is(err, ErrVolumeNotFound)
}

// isPermissionError checks if an error is related to permissions
func isPermissionError(err error) bool {
	errStr := err.Error()
	// Common permission error patterns
	permissionPatterns := []string{
		"permission denied",
		"access denied",
		"connection refused",
		"cannot connect to the Docker daemon",
	}

	for _, pattern := range permissionPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple case-insensitive check
	// In production, you might want to use strings.Contains with strings.ToLower
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
