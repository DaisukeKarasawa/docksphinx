package grpc

import (
	"math"
	"testing"
	"time"

	"docksphinx/internal/docker"
	"docksphinx/internal/monitor"
)

func TestStateToSnapshotClampsLargeValues(t *testing.T) {
	sm := monitor.NewStateManager()

	sm.UpdateState("c1", &monitor.ContainerState{
		ContainerID:      "c1",
		ContainerName:    "web",
		ImageName:        "nginx:latest",
		State:            "running",
		Status:           "Up",
		LastSeen:         time.Now(),
		RestartCount:     int(math.MaxInt32) + 99,
		VolumeMountCount: int(math.MaxInt32) + 88,
		CPUPercent:       10,
		MemoryPercent:    20,
		UptimeSeconds:    123,
		ComposeProject:   "proj",
		ComposeService:   "svc",
	})

	sm.UpdateResources(monitor.ResourceState{
		Networks: []docker.Network{
			{
				ID:             "n1",
				Name:           "net1",
				Driver:         "bridge",
				Scope:          "local",
				ContainerCount: int(math.MaxInt32) + 77,
			},
		},
		Volumes: []docker.Volume{
			{
				Name:       "v1",
				Driver:     "local",
				Mountpoint: "/tmp/v1",
				RefCount:   math.MaxInt32 + 66,
			},
		},
	})

	snapshot := StateToSnapshot(sm)
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if len(snapshot.GetContainers()) != 1 {
		t.Fatalf("expected 1 container, got %d", len(snapshot.GetContainers()))
	}
	c := snapshot.GetContainers()[0]
	if c.GetRestartCount() != math.MaxInt32 {
		t.Fatalf("expected restart_count clamp to MaxInt32, got %d", c.GetRestartCount())
	}
	if c.GetVolumeMountCount() != math.MaxInt32 {
		t.Fatalf("expected volume_mount_count clamp to MaxInt32, got %d", c.GetVolumeMountCount())
	}
	if len(snapshot.GetNetworks()) != 1 || snapshot.GetNetworks()[0].GetContainerCount() != math.MaxInt32 {
		t.Fatalf("expected network container_count clamp to MaxInt32, got %#v", snapshot.GetNetworks())
	}
	if len(snapshot.GetVolumes()) != 1 || snapshot.GetVolumes()[0].GetRefCount() != math.MaxInt32 {
		t.Fatalf("expected volume ref_count clamp to MaxInt32, got %#v", snapshot.GetVolumes())
	}
}
