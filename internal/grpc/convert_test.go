package grpc

import (
	"math"
	"reflect"
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

func TestStateToSnapshotSortsComposeGroupsAndFields(t *testing.T) {
	sm := monitor.NewStateManager()
	sm.UpdateResources(monitor.ResourceState{
		Groups: []monitor.ComposeGroup{
			{
				Project:        "zeta",
				Service:        "api",
				ContainerIDs:   []string{"cid-2", "cid-1"},
				ContainerNames: []string{"web-2", "web-1"},
				NetworkNames:   []string{"net-b", "net-a"},
			},
			{
				Project:        "alpha",
				Service:        "worker",
				ContainerIDs:   []string{"cid-4", "cid-3"},
				ContainerNames: []string{"job-2", "job-1"},
				NetworkNames:   []string{"net-d", "net-c"},
			},
		},
	})

	snapshot := StateToSnapshot(sm)
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if len(snapshot.GetGroups()) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(snapshot.GetGroups()))
	}

	first := snapshot.GetGroups()[0]
	second := snapshot.GetGroups()[1]
	if first.GetProject() != "alpha" || first.GetService() != "worker" {
		t.Fatalf("expected first group alpha/worker, got %s/%s", first.GetProject(), first.GetService())
	}
	if second.GetProject() != "zeta" || second.GetService() != "api" {
		t.Fatalf("expected second group zeta/api, got %s/%s", second.GetProject(), second.GetService())
	}

	if !reflect.DeepEqual(first.GetContainerIds(), []string{"cid-3", "cid-4"}) {
		t.Fatalf("unexpected sorted container ids for first group: %#v", first.GetContainerIds())
	}
	if !reflect.DeepEqual(first.GetContainerNames(), []string{"job-1", "job-2"}) {
		t.Fatalf("unexpected sorted container names for first group: %#v", first.GetContainerNames())
	}
	if !reflect.DeepEqual(first.GetNetworkNames(), []string{"net-c", "net-d"}) {
		t.Fatalf("unexpected sorted network names for first group: %#v", first.GetNetworkNames())
	}

	if !reflect.DeepEqual(second.GetContainerIds(), []string{"cid-1", "cid-2"}) {
		t.Fatalf("unexpected sorted container ids for second group: %#v", second.GetContainerIds())
	}
	if !reflect.DeepEqual(second.GetContainerNames(), []string{"web-1", "web-2"}) {
		t.Fatalf("unexpected sorted container names for second group: %#v", second.GetContainerNames())
	}
	if !reflect.DeepEqual(second.GetNetworkNames(), []string{"net-a", "net-b"}) {
		t.Fatalf("unexpected sorted network names for second group: %#v", second.GetNetworkNames())
	}
}
