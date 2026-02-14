package main

import (
	"reflect"
	"testing"

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
