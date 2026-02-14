package main

import (
	"reflect"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
)

func TestFormatFloat1OrNA(t *testing.T) {
	if got := formatFloat1OrNA(12.34, true); got != "12.3" {
		t.Fatalf("expected 12.3, got %s", got)
	}
	if got := formatFloat1OrNA(12.34, false); got != "N/A" {
		t.Fatalf("expected N/A, got %s", got)
	}
}

func TestFormatInt64OrNA(t *testing.T) {
	if got := formatInt64OrNA(123, true); got != "123" {
		t.Fatalf("expected 123, got %s", got)
	}
	if got := formatInt64OrNA(123, false); got != "N/A" {
		t.Fatalf("expected N/A, got %s", got)
	}
}

func TestFilteredContainerRowsForDetailSortAndNonMutating(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Containers: []*pb.ContainerInfo{
			{ContainerId: "id-z", ContainerName: "z-web", ImageName: "z:latest", State: "running", UptimeSeconds: 10},
			{ContainerId: "id-a", ContainerName: "a-web", ImageName: "a:latest", State: "running", UptimeSeconds: 50},
			{ContainerId: "id-m", ContainerName: "m-web", ImageName: "m:latest", State: "running", UptimeSeconds: 30},
		},
		Metrics: map[string]*pb.ContainerMetrics{
			"id-z": {CpuPercent: 20, MemoryPercent: 40},
			"id-a": {CpuPercent: 90, MemoryPercent: 10},
			"id-m": {CpuPercent: 50, MemoryPercent: 80},
		},
	}

	before := []string{
		m.snapshot.Containers[0].GetContainerName(),
		m.snapshot.Containers[1].GetContainerName(),
		m.snapshot.Containers[2].GetContainerName(),
	}

	m.sortMode = sortCPU
	cpuRows := m.filteredContainerRowsForDetail()
	if got := []string{cpuRows[0].GetContainerName(), cpuRows[1].GetContainerName(), cpuRows[2].GetContainerName()}; !reflect.DeepEqual(got, []string{"a-web", "m-web", "z-web"}) {
		t.Fatalf("expected cpu-desc order [a-web m-web z-web], got %v", got)
	}

	m.sortMode = sortMemory
	memRows := m.filteredContainerRowsForDetail()
	if got := []string{memRows[0].GetContainerName(), memRows[1].GetContainerName(), memRows[2].GetContainerName()}; !reflect.DeepEqual(got, []string{"m-web", "z-web", "a-web"}) {
		t.Fatalf("expected mem-desc order [m-web z-web a-web], got %v", got)
	}

	m.sortMode = sortUptime
	uptimeRows := m.filteredContainerRowsForDetail()
	if got := []string{uptimeRows[0].GetContainerName(), uptimeRows[1].GetContainerName(), uptimeRows[2].GetContainerName()}; !reflect.DeepEqual(got, []string{"a-web", "m-web", "z-web"}) {
		t.Fatalf("expected uptime-desc order [a-web m-web z-web], got %v", got)
	}

	m.sortMode = sortName
	nameRows := m.filteredContainerRowsForDetail()
	if got := []string{nameRows[0].GetContainerName(), nameRows[1].GetContainerName(), nameRows[2].GetContainerName()}; !reflect.DeepEqual(got, []string{"a-web", "m-web", "z-web"}) {
		t.Fatalf("expected name-asc order [a-web m-web z-web], got %v", got)
	}

	after := []string{
		m.snapshot.Containers[0].GetContainerName(),
		m.snapshot.Containers[1].GetContainerName(),
		m.snapshot.Containers[2].GetContainerName(),
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("expected source snapshot container order unchanged, before=%v after=%v", before, after)
	}
}

func TestFormatDateTimeOrNA(t *testing.T) {
	if got := formatDateTimeOrNA(0); got != "N/A" {
		t.Fatalf("expected N/A for missing timestamp, got %q", got)
	}
	unix := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC).Unix()
	if got := formatDateTimeOrNA(unix); got != "2026-02-14 10:30" {
		t.Fatalf("expected formatted datetime, got %q", got)
	}
}

func TestRenderImagesShowsNAForMissingCreatedTimestamp(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Images: []*pb.ImageInfo{
			{Repository: "busybox", Tag: "latest", Size: 123, CreatedUnix: 0},
		},
	}

	m.renderImages()
	cell := m.center.GetCell(1, 3)
	if cell == nil {
		t.Fatal("expected created column cell to exist")
	}
	if got := cell.Text; got != "N/A" {
		t.Fatalf("expected created column to render N/A, got %q", got)
	}
}

func TestFilteredContainerRowsForDetailUsesNameTieBreakForStableOrdering(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Containers: []*pb.ContainerInfo{
			{ContainerId: "id-b", ContainerName: "b-web", ImageName: "b:latest", State: "running", UptimeSeconds: 20},
			{ContainerId: "id-a", ContainerName: "a-web", ImageName: "a:latest", State: "running", UptimeSeconds: 20},
		},
		Metrics: map[string]*pb.ContainerMetrics{
			"id-b": {CpuPercent: 50, MemoryPercent: 60},
			"id-a": {CpuPercent: 50, MemoryPercent: 60},
		},
	}

	assertOrder := func(mode sortMode, want []string) {
		t.Helper()
		m.sortMode = mode
		rows := m.filteredContainerRowsForDetail()
		if len(rows) != len(want) {
			t.Fatalf("expected %d rows, got %d", len(want), len(rows))
		}
		got := make([]string, 0, len(rows))
		for _, r := range rows {
			got = append(got, r.GetContainerName())
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected order for mode=%v: got=%v want=%v", mode, got, want)
		}
	}

	assertOrder(sortCPU, []string{"a-web", "b-web"})
	assertOrder(sortMemory, []string{"a-web", "b-web"})
	assertOrder(sortUptime, []string{"a-web", "b-web"})
}

func TestFilteredContainerRowsForDetailUsesContainerIDTieBreakWhenNamesEqual(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Containers: []*pb.ContainerInfo{
			{ContainerId: "id-b", ContainerName: "same", ImageName: "b:latest", State: "running", UptimeSeconds: 20},
			{ContainerId: "id-a", ContainerName: "same", ImageName: "a:latest", State: "running", UptimeSeconds: 20},
		},
		Metrics: map[string]*pb.ContainerMetrics{
			"id-b": {CpuPercent: 50, MemoryPercent: 60},
			"id-a": {CpuPercent: 50, MemoryPercent: 60},
		},
	}

	assertOrder := func(mode sortMode, want []string) {
		t.Helper()
		m.sortMode = mode
		rows := m.filteredContainerRowsForDetail()
		if len(rows) != len(want) {
			t.Fatalf("expected %d rows, got %d", len(want), len(rows))
		}
		got := make([]string, 0, len(rows))
		for _, r := range rows {
			got = append(got, r.GetContainerId())
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected id order for mode=%v: got=%v want=%v", mode, got, want)
		}
	}

	assertOrder(sortCPU, []string{"id-a", "id-b"})
	assertOrder(sortMemory, []string{"id-a", "id-b"})
	assertOrder(sortUptime, []string{"id-a", "id-b"})
	assertOrder(sortName, []string{"id-a", "id-b"})
}

func TestRenderImagesUsesDeterministicTieBreakersAndNonMutating(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Images: []*pb.ImageInfo{
			{ImageId: "img-b", Repository: "same", Tag: "latest", Size: 20, CreatedUnix: 2},
			{ImageId: "img-a", Repository: "same", Tag: "latest", Size: 10, CreatedUnix: 1},
			{ImageId: "img-z", Repository: "alpha", Tag: "latest", Size: 30, CreatedUnix: 3},
		},
	}
	before := []string{
		m.snapshot.Images[0].GetImageId(),
		m.snapshot.Images[1].GetImageId(),
		m.snapshot.Images[2].GetImageId(),
	}

	m.renderImages()
	gotOrder := []string{
		m.center.GetCell(1, 0).Text + ":" + m.center.GetCell(1, 1).Text + ":" + m.center.GetCell(1, 2).Text,
		m.center.GetCell(2, 0).Text + ":" + m.center.GetCell(2, 1).Text + ":" + m.center.GetCell(2, 2).Text,
		m.center.GetCell(3, 0).Text + ":" + m.center.GetCell(3, 1).Text + ":" + m.center.GetCell(3, 2).Text,
	}
	wantOrder := []string{
		"alpha:latest:30",
		"same:latest:10",
		"same:latest:20",
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected rendered image row order: got=%v want=%v", gotOrder, wantOrder)
	}

	after := []string{
		m.snapshot.Images[0].GetImageId(),
		m.snapshot.Images[1].GetImageId(),
		m.snapshot.Images[2].GetImageId(),
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("expected source image order unchanged, before=%v after=%v", before, after)
	}
}

func TestRenderNetworksUsesDeterministicTieBreakersAndNonMutating(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Networks: []*pb.NetworkInfo{
			{NetworkId: "n2", Name: "same", Driver: "bridge", Scope: "local", ContainerCount: 2},
			{NetworkId: "n1", Name: "same", Driver: "bridge", Scope: "local", ContainerCount: 1},
			{NetworkId: "n0", Name: "alpha", Driver: "bridge", Scope: "local", ContainerCount: 3},
		},
	}
	before := []string{
		m.snapshot.Networks[0].GetNetworkId(),
		m.snapshot.Networks[1].GetNetworkId(),
		m.snapshot.Networks[2].GetNetworkId(),
	}

	m.renderNetworks()
	gotOrder := []string{
		m.center.GetCell(1, 0).Text + ":" + m.center.GetCell(1, 4).Text,
		m.center.GetCell(2, 0).Text + ":" + m.center.GetCell(2, 4).Text,
		m.center.GetCell(3, 0).Text + ":" + m.center.GetCell(3, 4).Text,
	}
	wantOrder := []string{
		"alpha:3",
		"same:1",
		"same:2",
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected rendered network row order: got=%v want=%v", gotOrder, wantOrder)
	}

	after := []string{
		m.snapshot.Networks[0].GetNetworkId(),
		m.snapshot.Networks[1].GetNetworkId(),
		m.snapshot.Networks[2].GetNetworkId(),
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("expected source network order unchanged, before=%v after=%v", before, after)
	}
}

func TestRenderVolumesUsesDeterministicTieBreakersAndNonMutating(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Volumes: []*pb.VolumeInfo{
			{Name: "same", Driver: "local", Mountpoint: "/m", RefCount: 2, UsageNote: "metadata-only"},
			{Name: "same", Driver: "local", Mountpoint: "/m", RefCount: 1, UsageNote: "metadata-only"},
			{Name: "alpha", Driver: "local", Mountpoint: "/a", RefCount: 3, UsageNote: "metadata-only"},
		},
	}
	before := []int32{
		m.snapshot.Volumes[0].GetRefCount(),
		m.snapshot.Volumes[1].GetRefCount(),
		m.snapshot.Volumes[2].GetRefCount(),
	}

	m.renderVolumes()
	gotOrder := []string{
		m.center.GetCell(1, 0).Text + ":" + m.center.GetCell(1, 2).Text,
		m.center.GetCell(2, 0).Text + ":" + m.center.GetCell(2, 2).Text,
		m.center.GetCell(3, 0).Text + ":" + m.center.GetCell(3, 2).Text,
	}
	wantOrder := []string{
		"alpha:3",
		"same:1",
		"same:2",
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected rendered volume row order: got=%v want=%v", gotOrder, wantOrder)
	}

	after := []int32{
		m.snapshot.Volumes[0].GetRefCount(),
		m.snapshot.Volumes[1].GetRefCount(),
		m.snapshot.Volumes[2].GetRefCount(),
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("expected source volume order unchanged, before=%v after=%v", before, after)
	}
}

func TestRenderGroupsUsesDeterministicTieBreakersAndNonMutating(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Groups: []*pb.ComposeGroup{
			{Project: "same", Service: "svc", ContainerIds: []string{"b", "a"}, ContainerNames: []string{"web-2", "web-1"}, NetworkNames: []string{"net-b", "net-a"}},
			{Project: "same", Service: "svc", ContainerIds: []string{"a"}, ContainerNames: []string{"api-1"}, NetworkNames: []string{"net-c", "net-a"}},
			{Project: "alpha", Service: "svc", ContainerIds: []string{"z"}, ContainerNames: []string{"alpha-1"}, NetworkNames: []string{"net-z"}},
		},
	}
	beforeGroupOrder := []string{
		m.snapshot.Groups[0].GetProject() + "/" + m.snapshot.Groups[0].GetService(),
		m.snapshot.Groups[1].GetProject() + "/" + m.snapshot.Groups[1].GetService(),
		m.snapshot.Groups[2].GetProject() + "/" + m.snapshot.Groups[2].GetService(),
	}
	beforeNames := append([]string(nil), m.snapshot.Groups[0].GetContainerNames()...)
	beforeNets := append([]string(nil), m.snapshot.Groups[0].GetNetworkNames()...)

	m.renderGroups()
	gotOrder := []string{
		m.center.GetCell(1, 0).Text + "/" + m.center.GetCell(1, 1).Text + ":" + m.center.GetCell(1, 2).Text + ":" + m.center.GetCell(1, 3).Text,
		m.center.GetCell(2, 0).Text + "/" + m.center.GetCell(2, 1).Text + ":" + m.center.GetCell(2, 2).Text + ":" + m.center.GetCell(2, 3).Text,
		m.center.GetCell(3, 0).Text + "/" + m.center.GetCell(3, 1).Text + ":" + m.center.GetCell(3, 2).Text + ":" + m.center.GetCell(3, 3).Text,
	}
	wantOrder := []string{
		"alpha/svc:alpha-1:net-z",
		"same/svc:api-1:net-a,net-c",
		"same/svc:web-1,web-2:net-a,net-b",
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected rendered group row order: got=%v want=%v", gotOrder, wantOrder)
	}

	afterGroupOrder := []string{
		m.snapshot.Groups[0].GetProject() + "/" + m.snapshot.Groups[0].GetService(),
		m.snapshot.Groups[1].GetProject() + "/" + m.snapshot.Groups[1].GetService(),
		m.snapshot.Groups[2].GetProject() + "/" + m.snapshot.Groups[2].GetService(),
	}
	afterNames := m.snapshot.Groups[0].GetContainerNames()
	afterNets := m.snapshot.Groups[0].GetNetworkNames()
	if !reflect.DeepEqual(beforeGroupOrder, afterGroupOrder) {
		t.Fatalf("expected source group order unchanged, before=%v after=%v", beforeGroupOrder, afterGroupOrder)
	}
	if !reflect.DeepEqual(beforeNames, afterNames) {
		t.Fatalf("expected source group container names unchanged, before=%v after=%v", beforeNames, afterNames)
	}
	if !reflect.DeepEqual(beforeNets, afterNets) {
		t.Fatalf("expected source group network names unchanged, before=%v after=%v", beforeNets, afterNets)
	}
}

func TestRenderResourcesSkipNilEntries(t *testing.T) {
	t.Run("images", func(t *testing.T) {
		m := newTUIModel()
		m.snapshot = &pb.Snapshot{
			Images: []*pb.ImageInfo{
				nil,
				{Repository: "busybox", Tag: "latest", Size: 123, CreatedUnix: 1},
			},
		}
		m.renderImages()
		if got := m.center.GetCell(1, 0).Text; got != "busybox" {
			t.Fatalf("expected first image row to be busybox, got %q", got)
		}
		if rows := m.center.GetRowCount(); rows != 2 {
			t.Fatalf("expected header+1 data row after nil-skip, got rowCount=%d", rows)
		}
	})

	t.Run("networks", func(t *testing.T) {
		m := newTUIModel()
		m.snapshot = &pb.Snapshot{
			Networks: []*pb.NetworkInfo{
				nil,
				{Name: "net1", Driver: "bridge", Scope: "local", ContainerCount: 1},
			},
		}
		m.renderNetworks()
		if got := m.center.GetCell(1, 0).Text; got != "net1" {
			t.Fatalf("expected first network row to be net1, got %q", got)
		}
		if rows := m.center.GetRowCount(); rows != 2 {
			t.Fatalf("expected header+1 data row after nil-skip, got rowCount=%d", rows)
		}
	})

	t.Run("volumes", func(t *testing.T) {
		m := newTUIModel()
		m.snapshot = &pb.Snapshot{
			Volumes: []*pb.VolumeInfo{
				nil,
				{Name: "vol1", Driver: "local", RefCount: 1, UsageNote: "metadata-only", Mountpoint: "/v"},
			},
		}
		m.renderVolumes()
		if got := m.center.GetCell(1, 0).Text; got != "vol1" {
			t.Fatalf("expected first volume row to be vol1, got %q", got)
		}
		if rows := m.center.GetRowCount(); rows != 2 {
			t.Fatalf("expected header+1 data row after nil-skip, got rowCount=%d", rows)
		}
	})

	t.Run("groups", func(t *testing.T) {
		m := newTUIModel()
		m.snapshot = &pb.Snapshot{
			Groups: []*pb.ComposeGroup{
				nil,
				{Project: "proj", Service: "svc", ContainerNames: []string{"c1"}, NetworkNames: []string{"n1"}},
			},
		}
		m.renderGroups()
		if got := m.center.GetCell(1, 0).Text; got != "proj" {
			t.Fatalf("expected first group row project to be proj, got %q", got)
		}
		if rows := m.center.GetRowCount(); rows != 2 {
			t.Fatalf("expected header+1 data row after nil-skip, got rowCount=%d", rows)
		}
	})
}

func TestFilteredContainerRowsForDetailSkipsNilEntries(t *testing.T) {
	m := newTUIModel()
	m.snapshot = &pb.Snapshot{
		Containers: []*pb.ContainerInfo{
			nil,
			{ContainerId: "id-a", ContainerName: "a-web", ImageName: "a:latest", State: "running", UptimeSeconds: 10},
		},
	}
	rows := m.filteredContainerRowsForDetail()
	if len(rows) != 1 {
		t.Fatalf("expected one non-nil container row, got %d", len(rows))
	}
	if rows[0].GetContainerId() != "id-a" {
		t.Fatalf("expected id-a row, got %#v", rows[0])
	}
}

func TestCompactEventsFiltersNilAndAppliesLimit(t *testing.T) {
	events := []*pb.Event{
		nil,
		{Id: "a"},
		nil,
		{Id: "b"},
		{Id: "c"},
	}

	got := compactEvents(events, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 compacted events, got %d", len(got))
	}
	if got[0].GetId() != "a" || got[1].GetId() != "b" {
		t.Fatalf("unexpected compacted order: %#v", got)
	}

	if got := compactEvents(events, 0); got != nil {
		t.Fatalf("expected nil when max=0, got %#v", got)
	}
	if got := compactEvents(events, -1); len(got) != 3 {
		t.Fatalf("expected all non-nil events when max<0, got %#v", got)
	}
	if got := compactEvents([]*pb.Event{nil, nil}, 10); got != nil {
		t.Fatalf("expected nil when all events are nil, got %#v", got)
	}
}

func TestLastEventTypeSkipsNilEntries(t *testing.T) {
	m := newTUIModel()
	m.events = []*pb.Event{
		nil,
		{ContainerId: "id-a", Type: "started"},
	}
	if got := m.lastEventType("id-a"); got != "started" {
		t.Fatalf("expected last event type started, got %q", got)
	}
}
