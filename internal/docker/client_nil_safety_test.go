package docker

import (
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestClientMethodsReturnExplicitErrorWhenReceiverNil(t *testing.T) {
	var c *Client

	tests := []struct {
		name string
		call func() error
	}{
		{name: "Ping", call: func() error { return c.Ping(context.Background()) }},
		{name: "ListContainers", call: func() error {
			_, err := c.ListContainers(context.Background(), ListContainersOptions{All: true})
			return err
		}},
		{name: "GetContainer", call: func() error {
			_, err := c.GetContainer(context.Background(), "cid")
			return err
		}},
		{name: "GetContainerDetails", call: func() error {
			_, err := c.GetContainerDetails(context.Background(), "cid")
			return err
		}},
		{name: "ListImages", call: func() error {
			_, err := c.ListImages(context.Background())
			return err
		}},
		{name: "GetImage", call: func() error {
			_, err := c.GetImage(context.Background(), "img")
			return err
		}},
		{name: "ListNetworks", call: func() error {
			_, err := c.ListNetworks(context.Background())
			return err
		}},
		{name: "GetNetwork", call: func() error {
			_, err := c.GetNetwork(context.Background(), "net")
			return err
		}},
		{name: "ListVolumes", call: func() error {
			_, err := c.ListVolumes(context.Background())
			return err
		}},
		{name: "GetVolume", call: func() error {
			_, err := c.GetVolume(context.Background(), "vol")
			return err
		}},
		{name: "GetContainerStats", call: func() error {
			_, err := c.GetContainerStats(context.Background(), "cid")
			return err
		}},
		{name: "GetMemoryUsage", call: func() error {
			_, _, err := c.GetMemoryUsage(context.Background(), "cid")
			return err
		}},
		{name: "GetNetworkStats", call: func() error {
			_, _, err := c.GetNetworkStats(context.Background(), "cid")
			return err
		}},
		{name: "GetVolumeUsage", call: func() error {
			_, err := c.GetVolumeUsage(context.Background(), "cid")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatalf("expected %s to return an error for nil receiver", tt.name)
			}
			if !strings.Contains(err.Error(), "docker client is nil") {
				t.Fatalf("expected nil receiver error for %s, got: %v", tt.name, err)
			}
		})
	}

	if got := c.GetAPIClient(); got != nil {
		t.Fatalf("expected GetAPIClient on nil receiver to return nil, got %#v", got)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("expected Close on nil receiver to be no-op, got %v", err)
	}
}

func TestClientMethodsReturnExplicitErrorWhenAPIClientMissing(t *testing.T) {
	c := &Client{}

	tests := []struct {
		name string
		call func() error
	}{
		{name: "Ping", call: func() error { return c.Ping(context.Background()) }},
		{name: "ListContainers", call: func() error {
			_, err := c.ListContainers(context.Background(), ListContainersOptions{All: true})
			return err
		}},
		{name: "GetContainerStats", call: func() error {
			_, err := c.GetContainerStats(context.Background(), "cid")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatalf("expected %s to return an error for missing api client", tt.name)
			}
			if !strings.Contains(err.Error(), "docker api client is nil") {
				t.Fatalf("expected missing api client error for %s, got: %v", tt.name, err)
			}
		})
	}

	if err := c.Close(); err != nil {
		t.Fatalf("expected Close on missing api client to be no-op, got %v", err)
	}
}

func TestCalculateStatusNilState(t *testing.T) {
	if got := calculateStatus(nil); got != "Unknown" {
		t.Fatalf("expected nil container state to map to Unknown, got %q", got)
	}
}

func TestCalculateStatusKnownStates(t *testing.T) {
	if got := calculateStatus(&container.State{Status: "created"}); got != "Created" {
		t.Fatalf("expected created status, got %q", got)
	}
	if got := calculateStatus(&container.State{Status: "paused"}); got != "Paused" {
		t.Fatalf("expected paused status, got %q", got)
	}
}
