package event

import "sync"

// History keeps recent events in memory.
type History struct {
	mu      sync.RWMutex
	maxSize int
	events  []*Event
}

// NewHistory creates an in-memory event history with a max size.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 1
	}
	return &History{
		maxSize: maxSize,
		events:  make([]*Event, 0, maxSize),
	}
}

// Add appends an event and evicts old entries if needed.
func (h *History) Add(ev *Event) {
	if ev == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = append(h.events, ev)
	if len(h.events) > h.maxSize {
		drop := len(h.events) - h.maxSize
		h.events = append([]*Event(nil), h.events[drop:]...)
	}
}

// Recent returns up to n recent events from newest to oldest.
// If n <= 0 or n > len(history), returns all available events.
func (h *History) Recent(n int) []*Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := len(h.events)
	if total == 0 {
		return nil
	}
	if n <= 0 || n > total {
		n = total
	}

	out := make([]*Event, 0, n)
	for i := total - 1; i >= total-n; i-- {
		out = append(out, h.events[i])
	}
	return out
}
