package eventorder

import (
	"reflect"
	"sort"
	"testing"
	"time"

	pb "docksphinx/api/docksphinx/v1"
	"docksphinx/internal/event"
)

func TestLessPBAndLessInternalProduceSameOrder(t *testing.T) {
	base := time.Unix(1700000500, 0)

	pbEvents := []*pb.Event{
		{TimestampUnix: base.Unix(), Id: "", ContainerName: "same", Type: "cpu_threshold", Message: "a", ContainerId: "cid-2", ImageName: "img-b"},
		{TimestampUnix: base.Unix(), Id: "", ContainerName: "same", Type: "cpu_threshold", Message: "a", ContainerId: "cid-1", ImageName: "img-c"},
		{TimestampUnix: base.Unix(), Id: "", ContainerName: "same", Type: "cpu_threshold", Message: "a", ContainerId: "cid-1", ImageName: "img-a"},
		{TimestampUnix: base.Unix(), Id: "b", ContainerName: "same-second", Type: "cpu_threshold", Message: "same-second", ContainerId: "cid-b", ImageName: "img-b"},
		{TimestampUnix: base.Unix(), Id: "a", ContainerName: "same-second", Type: "cpu_threshold", Message: "same-second", ContainerId: "cid-a", ImageName: "img-a"},
		{TimestampUnix: base.Add(-time.Second).Unix(), Id: "z", ContainerName: "older", Type: "cpu_threshold", Message: "z", ContainerId: "cid-z", ImageName: "img-z"},
	}

	internalEvents := []*event.Event{
		{Timestamp: base.Add(900 * time.Millisecond), ID: "", ContainerName: "same", Type: event.EventTypeCPUThreshold, Message: "a", ContainerID: "cid-2", ImageName: "img-b"},
		{Timestamp: base.Add(100 * time.Millisecond), ID: "", ContainerName: "same", Type: event.EventTypeCPUThreshold, Message: "a", ContainerID: "cid-1", ImageName: "img-c"},
		{Timestamp: base.Add(800 * time.Millisecond), ID: "", ContainerName: "same", Type: event.EventTypeCPUThreshold, Message: "a", ContainerID: "cid-1", ImageName: "img-a"},
		{Timestamp: base.Add(900 * time.Millisecond), ID: "b", ContainerName: "same-second", Type: event.EventTypeCPUThreshold, Message: "same-second", ContainerID: "cid-b", ImageName: "img-b"},
		{Timestamp: base.Add(100 * time.Millisecond), ID: "a", ContainerName: "same-second", Type: event.EventTypeCPUThreshold, Message: "same-second", ContainerID: "cid-a", ImageName: "img-a"},
		{Timestamp: base.Add(-time.Second), ID: "z", ContainerName: "older", Type: event.EventTypeCPUThreshold, Message: "z", ContainerID: "cid-z", ImageName: "img-z"},
	}

	sort.Slice(pbEvents, func(i, j int) bool { return LessPB(pbEvents[i], pbEvents[j]) })
	sort.Slice(internalEvents, func(i, j int) bool { return LessInternal(internalEvents[i], internalEvents[j]) })

	keyPB := func(e *pb.Event) string {
		return stringsJoin(
			e.GetId(),
			e.GetContainerName(),
			e.GetType(),
			e.GetMessage(),
			e.GetContainerId(),
			e.GetImageName(),
		)
	}
	keyInternal := func(e *event.Event) string {
		return stringsJoin(
			e.ID,
			e.ContainerName,
			string(e.Type),
			e.Message,
			e.ContainerID,
			e.ImageName,
		)
	}

	gotPB := make([]string, 0, len(pbEvents))
	for _, e := range pbEvents {
		gotPB = append(gotPB, keyPB(e))
	}
	gotInternal := make([]string, 0, len(internalEvents))
	for _, e := range internalEvents {
		gotInternal = append(gotInternal, keyInternal(e))
	}

	if !reflect.DeepEqual(gotPB, gotInternal) {
		t.Fatalf("expected pb/internal ordering parity:\n pb=%v\n internal=%v", gotPB, gotInternal)
	}
}

func TestLessInternalUsesSecondLevelTimestampBeforeID(t *testing.T) {
	base := time.Unix(1700000600, 0)
	a := &event.Event{
		Timestamp:     base.Add(900 * time.Millisecond),
		ID:            "a",
		ContainerName: "same-second",
		Type:          event.EventTypeCPUThreshold,
		Message:       "same-second",
		ContainerID:   "cid-a",
		ImageName:     "img-a",
	}
	b := &event.Event{
		Timestamp:     base.Add(100 * time.Millisecond),
		ID:            "b",
		ContainerName: "same-second",
		Type:          event.EventTypeCPUThreshold,
		Message:       "same-second",
		ContainerID:   "cid-b",
		ImageName:     "img-b",
	}

	if !LessInternal(a, b) {
		t.Fatalf("expected ID tie-break to win within same second; a should be before b")
	}
	if LessInternal(b, a) {
		t.Fatalf("expected b not to be before a")
	}
}

func stringsJoin(parts ...string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "|"
		}
		out += p
	}
	return out
}
