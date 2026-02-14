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
	h.events = append(h.events, ev)
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
		out = append(out, h.events[i])
	}
	return out
}
