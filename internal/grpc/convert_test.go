package grpc

import (
	"math"
	"reflect"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/docker"
	"docksphinx/internal/event"
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
	inputGroups := []monitor.ComposeGroup{
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
	}
	sm.UpdateResources(monitor.ResourceState{
		Groups: inputGroups,
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

	// Non-mutating contract: StateToSnapshot sorting must not alter source resources.
	if !reflect.DeepEqual(inputGroups[0].ContainerIDs, []string{"cid-2", "cid-1"}) {
		t.Fatalf("expected caller input groups unchanged, got %#v", inputGroups[0].ContainerIDs)
	}
	resourcesAfter := sm.GetResources()
	if len(resourcesAfter.Groups) != 2 {
		t.Fatalf("expected 2 source groups after snapshot conversion, got %d", len(resourcesAfter.Groups))
	}
	if resourcesAfter.Groups[0].Project != "zeta" || resourcesAfter.Groups[1].Project != "alpha" {
		t.Fatalf("expected source group order unchanged [zeta alpha], got [%s %s]", resourcesAfter.Groups[0].Project, resourcesAfter.Groups[1].Project)
	}
	if !reflect.DeepEqual(resourcesAfter.Groups[0].ContainerIDs, []string{"cid-2", "cid-1"}) {
		t.Fatalf("expected source group container id order unchanged, got %#v", resourcesAfter.Groups[0].ContainerIDs)
	}
	if !reflect.DeepEqual(resourcesAfter.Groups[0].ContainerNames, []string{"web-2", "web-1"}) {
		t.Fatalf("expected source group container name order unchanged, got %#v", resourcesAfter.Groups[0].ContainerNames)
	}
	if !reflect.DeepEqual(resourcesAfter.Groups[0].NetworkNames, []string{"net-b", "net-a"}) {
		t.Fatalf("expected source group network order unchanged, got %#v", resourcesAfter.Groups[0].NetworkNames)
	}
}

func TestStateToSnapshotSortsResourcesWithoutMutatingSource(t *testing.T) {
	sm := monitor.NewStateManager()
	inputImages := []docker.Image{
		{ID: "img-z", Repository: "zrepo", Tag: "latest", Size: 2, Created: 2},
		{ID: "img-a", Repository: "arepo", Tag: "latest", Size: 1, Created: 1},
	}
	inputNetworks := []docker.Network{
		{ID: "n-z", Name: "znet", Driver: "bridge", Scope: "local", ContainerCount: 1},
		{ID: "n-a", Name: "anet", Driver: "bridge", Scope: "local", ContainerCount: 2},
	}
	inputVolumes := []docker.Volume{
		{Name: "zvol", Driver: "local", Mountpoint: "/z", RefCount: 1, UsageNote: "metadata-only"},
		{Name: "avol", Driver: "local", Mountpoint: "/a", RefCount: 2, UsageNote: "metadata-only"},
	}
	sm.UpdateResources(monitor.ResourceState{
		Images:   inputImages,
		Networks: inputNetworks,
		Volumes:  inputVolumes,
	})

	snapshot := StateToSnapshot(sm)
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}

	if got := snapshot.GetImages(); len(got) != 2 || got[0].GetRepository() != "arepo" || got[1].GetRepository() != "zrepo" {
		t.Fatalf("expected sorted images [arepo zrepo], got %#v", got)
	}
	if got := snapshot.GetNetworks(); len(got) != 2 || got[0].GetName() != "anet" || got[1].GetName() != "znet" {
		t.Fatalf("expected sorted networks [anet znet], got %#v", got)
	}
	if got := snapshot.GetVolumes(); len(got) != 2 || got[0].GetName() != "avol" || got[1].GetName() != "zvol" {
		t.Fatalf("expected sorted volumes [avol zvol], got %#v", got)
	}

	if !reflect.DeepEqual([]string{inputImages[0].Repository, inputImages[1].Repository}, []string{"zrepo", "arepo"}) {
		t.Fatalf("expected caller image input order unchanged, got %#v", inputImages)
	}
	if !reflect.DeepEqual([]string{inputNetworks[0].Name, inputNetworks[1].Name}, []string{"znet", "anet"}) {
		t.Fatalf("expected caller network input order unchanged, got %#v", inputNetworks)
	}
	if !reflect.DeepEqual([]string{inputVolumes[0].Name, inputVolumes[1].Name}, []string{"zvol", "avol"}) {
		t.Fatalf("expected caller volume input order unchanged, got %#v", inputVolumes)
	}

	resourcesAfter := sm.GetResources()
	if !reflect.DeepEqual([]string{resourcesAfter.Images[0].Repository, resourcesAfter.Images[1].Repository}, []string{"zrepo", "arepo"}) {
		t.Fatalf("expected source image order unchanged, got %#v", resourcesAfter.Images)
	}
	if !reflect.DeepEqual([]string{resourcesAfter.Networks[0].Name, resourcesAfter.Networks[1].Name}, []string{"znet", "anet"}) {
		t.Fatalf("expected source network order unchanged, got %#v", resourcesAfter.Networks)
	}
	if !reflect.DeepEqual([]string{resourcesAfter.Volumes[0].Name, resourcesAfter.Volumes[1].Name}, []string{"zvol", "avol"}) {
		t.Fatalf("expected source volume order unchanged, got %#v", resourcesAfter.Volumes)
	}
}

func TestEventsToProtoSkipsNilAndConvertsFields(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	events := []*event.Event{
		nil,
		{
			ID:            "ev-1",
			Type:          event.EventTypeCPUThreshold,
			Timestamp:     ts,
			ContainerID:   "c1",
			ContainerName: "web",
			ImageName:     "nginx:latest",
			Message:       "cpu high",
			Data: map[string]interface{}{
				"threshold": 80,
				"actual":    92.5,
				"note":      "critical",
			},
		},
		nil,
	}

	got := EventsToProto(events)
	if len(got) != 1 {
		t.Fatalf("expected nil events to be skipped, got len=%d", len(got))
	}

	ev := got[0]
	if ev.GetId() != "ev-1" {
		t.Fatalf("unexpected id: %q", ev.GetId())
	}
	if ev.GetType() != string(event.EventTypeCPUThreshold) {
		t.Fatalf("unexpected type: %q", ev.GetType())
	}
	if ev.GetTimestampUnix() != ts.Unix() {
		t.Fatalf("unexpected timestamp: %d", ev.GetTimestampUnix())
	}
	if ev.GetContainerId() != "c1" || ev.GetContainerName() != "web" || ev.GetImageName() != "nginx:latest" {
		t.Fatalf("unexpected container/image fields: %#v", ev)
	}
	if ev.GetMessage() != "cpu high" {
		t.Fatalf("unexpected message: %q", ev.GetMessage())
	}
	wantData := map[string]string{
		"threshold": "80",
		"actual":    "92.5",
		"note":      "critical",
	}
	if !reflect.DeepEqual(ev.GetData(), wantData) {
		t.Fatalf("unexpected data conversion:\n got=%#v\nwant=%#v", ev.GetData(), wantData)
	}
}

func TestEventToProtoNil(t *testing.T) {
	var in *event.Event
	var out *pb.Event = EventToProto(in)
	if out != nil {
		t.Fatalf("expected nil for nil input, got %#v", out)
	}
}
