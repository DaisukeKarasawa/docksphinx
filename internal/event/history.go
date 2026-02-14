package event

import "sync"

// History stores recent events in memory.
type History struct {
	mu      sync.RWMutex
	maxSize int
	events  []*Event
}

func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 1
	}
	return &History{
		maxSize: maxSize,
		events:  make([]*Event, 0, maxSize),
	}
}

func (h *History) Add(ev *Event) {
	if h == nil || ev == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, cloneEvent(ev))
	if len(h.events) > h.maxSize {
		h.events = h.events[len(h.events)-h.maxSize:]
	}
}

// Recent returns recent events from newest to oldest.
func (h *History) Recent(limit int) []*Event {
	if h == nil {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.events) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(h.events) {
		limit = len(h.events)
	}
	out := make([]*Event, 0, limit)
	for i := len(h.events) - 1; i >= len(h.events)-limit; i-- {
		out = append(out, cloneEvent(h.events[i]))
	}
	return out
}

func cloneEvent(in *Event) *Event {
	if in == nil {
		return nil
	}
	out := *in
	if in.Data != nil {
		out.Data = make(map[string]interface{}, len(in.Data))
		for k, v := range in.Data {
			out.Data[k] = cloneValue(v)
		}
	}
	return &out
}

func cloneValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, vv := range x {
			out[k] = cloneValue(vv)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, vv := range x {
			out[i] = cloneValue(vv)
		}
		return out
	case []string:
		return append([]string(nil), x...)
	case []int:
		return append([]int(nil), x...)
	case []int64:
		return append([]int64(nil), x...)
	case []float64:
		return append([]float64(nil), x...)
	case []bool:
		return append([]bool(nil), x...)
	default:
		return x
	}
}
