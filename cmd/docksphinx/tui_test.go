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
