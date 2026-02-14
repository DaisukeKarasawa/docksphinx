package event

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

type structuredPayload struct {
	Name  string
	Items []string
	Meta  map[string]string
}

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

func TestNewHistoryEnforcesMinSize(t *testing.T) {
	h := NewHistory(0)
	for i := 0; i < 3; i++ {
		h.Add(&Event{ID: "e", Timestamp: time.Now()})
	}
	got := h.Recent(10)
	if len(got) != 1 {
		t.Fatalf("expected min capacity behavior to keep 1 event, got %d", len(got))
	}
}

func TestHistoryRecentLimitContract(t *testing.T) {
	h := NewHistory(5)
	h.Add(&Event{ID: "e1", Timestamp: time.Unix(1, 0)})
	h.Add(&Event{ID: "e2", Timestamp: time.Unix(2, 0)})
	h.Add(&Event{ID: "e3", Timestamp: time.Unix(3, 0)})

	if got := h.Recent(0); len(got) != 3 {
		t.Fatalf("expected full length for zero limit, got %d", len(got))
	}
	if got := h.Recent(-1); len(got) != 3 {
		t.Fatalf("expected full length for negative limit, got %d", len(got))
	}
	if got := h.Recent(99); len(got) != 3 {
		t.Fatalf("expected full length for large limit, got %d", len(got))
	}
}

func TestHistoryAddAndRecentAreMutationSafe(t *testing.T) {
	h := NewHistory(5)

	nestedMap := map[string]interface{}{"name": "alpha"}
	nestedSlice := []interface{}{"x", "y"}
	stringMap := map[string]string{"owner": "team-a"}
	stringSlice := []string{"svc-a", "svc-b"}
	arrayPayload := [2][]string{{"arr-a"}, {"arr-b"}}
	keyInt := 7
	pointerKeyMap := map[*int]string{&keyInt: "seven"}
	structured := structuredPayload{
		Name:  "payload",
		Items: []string{"it-1", "it-2"},
		Meta:  map[string]string{"k": "v"},
	}
	ptrStructured := &structuredPayload{
		Name:  "ptr-payload",
		Items: []string{"pit-1", "pit-2"},
		Meta:  map[string]string{"pk": "pv"},
	}
	original := &Event{
		ID:        "e1",
		Type:      EventTypeCPUThreshold,
		Timestamp: time.Unix(100, 0),
		Data: map[string]interface{}{
			"cpu":           95.5,
			"meta":          nestedMap,
			"tags":          nestedSlice,
			"labels":        stringMap,
			"names":         stringSlice,
			"arrayPayload":  arrayPayload,
			"pointerKeyMap": pointerKeyMap,
			"structured":    structured,
			"structuredPtr": ptrStructured,
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
	stringMap["owner"] = "mutated-team"
	stringSlice[0] = "changed-svc"
	arrayPayload[0][0] = "mutated-arr-a"
	pointerKeyMap[&keyInt] = "mutated-seven"
	structured.Items[0] = "mutated-item"
	structured.Meta["k"] = "mutated-meta"
	ptrStructured.Items[0] = "mutated-ptr-item"
	ptrStructured.Meta["pk"] = "mutated-ptr-meta"

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
	if gotLabels, ok := got[0].Data["labels"].(map[string]string); !ok || gotLabels["owner"] != "team-a" {
		t.Fatalf("expected typed map to be isolated, got %#v", got[0].Data["labels"])
	}
	if gotNames, ok := got[0].Data["names"].([]string); !ok || len(gotNames) != 2 || gotNames[0] != "svc-a" {
		t.Fatalf("expected typed slice to be isolated, got %#v", got[0].Data["names"])
	}
	if gotArrayPayload, ok := got[0].Data["arrayPayload"].([2][]string); !ok ||
		len(gotArrayPayload[0]) != 1 || gotArrayPayload[0][0] != "arr-a" {
		t.Fatalf("expected array payload to be isolated, got %#v", got[0].Data["arrayPayload"])
	}
	if gotPointerKeyMap, ok := got[0].Data["pointerKeyMap"].(map[*int]string); !ok || gotPointerKeyMap[&keyInt] != "seven" {
		t.Fatalf("expected pointer-key map to be isolated and key-preserved, got %#v", got[0].Data["pointerKeyMap"])
	}
	if gotStructured, ok := got[0].Data["structured"].(structuredPayload); !ok ||
		len(gotStructured.Items) != 2 || gotStructured.Items[0] != "it-1" ||
		gotStructured.Meta["k"] != "v" {
		t.Fatalf("expected struct payload to be isolated, got %#v", got[0].Data["structured"])
	}
	if gotStructuredPtr, ok := got[0].Data["structuredPtr"].(*structuredPayload); !ok ||
		gotStructuredPtr == nil ||
		len(gotStructuredPtr.Items) != 2 || gotStructuredPtr.Items[0] != "pit-1" ||
		gotStructuredPtr.Meta["pk"] != "pv" {
		t.Fatalf("expected struct pointer payload to be isolated, got %#v", got[0].Data["structuredPtr"])
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
	if labels, ok := got[0].Data["labels"].(map[string]string); ok {
		labels["owner"] = "out-mutated-team"
	}
	if names, ok := got[0].Data["names"].([]string); ok && len(names) > 0 {
		names[0] = "out-changed-svc"
	}
	if arr, ok := got[0].Data["arrayPayload"].([2][]string); ok {
		arr[0][0] = "out-mutated-arr-a"
		got[0].Data["arrayPayload"] = arr
	}
	if pointerMap, ok := got[0].Data["pointerKeyMap"].(map[*int]string); ok {
		pointerMap[&keyInt] = "out-mutated-seven"
	}
	if payload, ok := got[0].Data["structured"].(structuredPayload); ok {
		payload.Items[0] = "out-mutated-item"
		payload.Meta["k"] = "out-mutated-meta"
		got[0].Data["structured"] = payload
	}
	if payloadPtr, ok := got[0].Data["structuredPtr"].(*structuredPayload); ok && payloadPtr != nil {
		payloadPtr.Items[0] = "out-mutated-ptr-item"
		payloadPtr.Meta["pk"] = "out-mutated-ptr-meta"
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
	if gotLabels, ok := gotAgain[0].Data["labels"].(map[string]string); !ok || gotLabels["owner"] != "team-a" {
		t.Fatalf("expected typed map in history to stay unchanged, got %#v", gotAgain[0].Data["labels"])
	}
	if gotNames, ok := gotAgain[0].Data["names"].([]string); !ok || len(gotNames) != 2 || gotNames[0] != "svc-a" {
		t.Fatalf("expected typed slice in history to stay unchanged, got %#v", gotAgain[0].Data["names"])
	}
	if gotArrayPayload, ok := gotAgain[0].Data["arrayPayload"].([2][]string); !ok ||
		len(gotArrayPayload[0]) != 1 || gotArrayPayload[0][0] != "arr-a" {
		t.Fatalf("expected array payload in history to stay unchanged, got %#v", gotAgain[0].Data["arrayPayload"])
	}
	if gotPointerKeyMap, ok := gotAgain[0].Data["pointerKeyMap"].(map[*int]string); !ok || gotPointerKeyMap[&keyInt] != "seven" {
		t.Fatalf("expected pointer-key map in history to stay unchanged, got %#v", gotAgain[0].Data["pointerKeyMap"])
	}
	if gotStructured, ok := gotAgain[0].Data["structured"].(structuredPayload); !ok ||
		len(gotStructured.Items) != 2 || gotStructured.Items[0] != "it-1" ||
		gotStructured.Meta["k"] != "v" {
		t.Fatalf("expected struct payload in history to stay unchanged, got %#v", gotAgain[0].Data["structured"])
	}
	if gotStructuredPtr, ok := gotAgain[0].Data["structuredPtr"].(*structuredPayload); !ok ||
		gotStructuredPtr == nil ||
		len(gotStructuredPtr.Items) != 2 || gotStructuredPtr.Items[0] != "pit-1" ||
		gotStructuredPtr.Meta["pk"] != "pv" {
		t.Fatalf("expected struct pointer payload in history to stay unchanged, got %#v", gotAgain[0].Data["structuredPtr"])
	}
}

func TestHistoryConcurrentAddAndRecent(t *testing.T) {
	h := NewHistory(50)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			h.Add(&Event{
				ID:        "ev",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"i": i,
				},
			})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			got := h.Recent(10)
			if len(got) > 10 {
				t.Errorf("recent size exceeded limit: %d", len(got))
				return
			}
			for _, ev := range got {
				if ev == nil {
					t.Errorf("recent returned nil event")
					return
				}
			}
		}
	}()

	wg.Wait()
}
