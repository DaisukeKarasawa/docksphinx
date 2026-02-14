package grpc

import (
	"context"
	"sync"

	"docksphinx/internal/event"
)

const subscriberChanBuf = 32

// Broadcaster fans out events to multiple subscribers (e.g. Stream RPC handlers)
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan *event.Event]struct{}
}

// NewBroadcaster creates a new Broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan *event.Event]struct{}),
	}
}

// Subscribe adds a new subscriber and returns a channel and an unsubscribe function.
// Call the returned function (e.g. defer unsub()) when done receiving.
func (b *Broadcaster) Subscribe() (ch <-chan *event.Event, unsub func()) {
	if b == nil {
		closed := make(chan *event.Event)
		close(closed)
		return closed, func() {}
	}
	writable := make(chan *event.Event, subscriberChanBuf)
	b.mu.Lock()
	b.subscribers[writable] = struct{}{}
	b.mu.Unlock()
	return writable, func() { b.Unsubscribe(writable) }
}

// Unsubscribe removes the subscriber and closes its channel
func (b *Broadcaster) Unsubscribe(ch chan *event.Event) {
	if b == nil || ch == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Send sends the event to all current subscribers (non-blocking; drops if channel full)
func (b *Broadcaster) Send(ev *event.Event) {
	if b == nil {
		return
	}
	if ev == nil {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- ev:
		default:
			// subscriber too slow; skip this subscriber for this event
		}
	}
}

// Run reads from src and forwards to all subscribers. Call in a goroutine;
// stops when ctx is canceled or src is closed.
func (b *Broadcaster) Run(ctx context.Context, src <-chan *event.Event) {
	if b == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if src == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-src:
			if !ok {
				return
			}
			b.Send(ev)
		}
	}
}
