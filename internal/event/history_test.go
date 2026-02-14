package event

import (
	"reflect"
	"testing"
	"time"
)

func TestHistoryRecentReturnsNewestFirstWithinLimit(t *testing.T) {
	h := NewHistory(2)

	ev1 := &Event{ID: "e1", Timestamp: time.Unix(1, 0)}
	ev2 := &Event{ID: "e2", Timestamp: time.Unix(2, 0)}
	ev3 := &Event{ID: "e3", Timestamp: time.Unix(3, 0)}
	h.Add(ev1)
	h.Add(ev2)
	h.Add(ev3)

	got := h.Recent(10)
	if len(got) != 2 {
		t.Fatalf("expected 2 events with max size cap, got %d", len(got))
	}
	if got[0].ID != "e3" || got[1].ID != "e2" {
		t.Fatalf("expected newest-first [e3 e2], got [%s %s]", got[0].ID, got[1].ID)
	}
}

func TestHistoryAddAndRecentAreMutationSafe(t *testing.T) {
	h := NewHistory(5)

	nestedMap := map[string]interface{}{"name": "alpha"}
	nestedSlice := []interface{}{"x", "y"}
	original := &Event{
		ID:        "e1",
		Type:      EventTypeCPUThreshold,
		Timestamp: time.Unix(100, 0),
		Data: map[string]interface{}{
			"cpu":  95.5,
			"meta": nestedMap,
			"tags": nestedSlice,
		},
		Message: "high cpu",
	}
	h.Add(original)

	// mutate original after Add; history must be isolated.
	original.ID = "mutated-input"
	original.Message = "changed"
	original.Data["cpu"] = 1.0
	original.Data["new"] = "x"
	nestedMap["name"] = "mutated"
	nestedSlice[0] = "changed"

	got := h.Recent(1)
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].ID != "e1" || got[0].Message != "high cpu" {
		t.Fatalf("expected stored event to keep original values, got id=%q message=%q", got[0].ID, got[0].Message)
	}
	if gotMeta, ok := got[0].Data["meta"].(map[string]interface{}); !ok || gotMeta["name"] != "alpha" {
		t.Fatalf("expected nested map to be isolated, got %#v", got[0].Data["meta"])
	}
	if gotTags, ok := got[0].Data["tags"].([]interface{}); !ok || len(gotTags) != 2 || gotTags[0] != "x" {
		t.Fatalf("expected nested slice to be isolated, got %#v", got[0].Data["tags"])
	}
	if !reflect.DeepEqual(got[0].Data["cpu"], 95.5) {
		t.Fatalf("expected stored data to be unchanged, got %#v", got[0].Data)
	}

	// mutate returned event; history must stay isolated on next read.
	got[0].ID = "mutated-output"
	got[0].Data["cpu"] = 0.0
	if meta, ok := got[0].Data["meta"].(map[string]interface{}); ok {
		meta["name"] = "out-mutated"
	}
	if tags, ok := got[0].Data["tags"].([]interface{}); ok && len(tags) > 0 {
		tags[0] = "out-changed"
	}

	gotAgain := h.Recent(1)
	if gotAgain[0].ID != "e1" {
		t.Fatalf("expected history to be immune from output mutation, got id=%q", gotAgain[0].ID)
	}
	if gotAgain[0].Data["cpu"] != 95.5 {
		t.Fatalf("expected history data to be immune from output mutation, got cpu=%v", gotAgain[0].Data["cpu"])
	}
	if gotMeta, ok := gotAgain[0].Data["meta"].(map[string]interface{}); !ok || gotMeta["name"] != "alpha" {
		t.Fatalf("expected nested map in history to stay unchanged, got %#v", gotAgain[0].Data["meta"])
	}
	if gotTags, ok := gotAgain[0].Data["tags"].([]interface{}); !ok || len(gotTags) != 2 || gotTags[0] != "x" {
		t.Fatalf("expected nested slice in history to stay unchanged, got %#v", gotAgain[0].Data["tags"])
	}
}
