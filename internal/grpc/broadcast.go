package grpc

import (
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
	writable := make(chan *event.Event, subscriberChanBuf)
	b.mu.Lock()
	b.subscribers[writable] = struct{}{}
	b.mu.Unlock()
	return writable, func() { b.Unsubscribe(writable) }
}

// Unsubscribe removes the subscriber and closes its channel
func (b *Broadcaster) Unsubscribe(ch chan *event.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch) // 既に閉じられたチャネルを再度閉じる可能性がある（panicの原因）
	}
}

// Send sends the event to all current subscribers (non-blocking; drops if channel full)
func (b *Broadcaster) Send(ev *event.Event) {
	if ev == nil {
		return
	}
	// Create a snapshot of subscriber channels while holding the lock
	// to prevent sending to channels that are closed by cleanup
	b.mu.RLock()
	subs := make([]chan *event.Event, 0, len(b.subscribers))
	for ch := range b.subscribers {
		subs = append(subs, ch)
	}
	b.mu.RUnlock()
	
	// Iterate over the snapshot without holding the lock
	// Use recover to handle send-on-closed-channel panics
	for _, ch := range subs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Channel was closed; ignore the panic
				}
			}()
			select {
			case ch <- ev:
			default:
				// subscriber too slow; skip this subscriber for this event
			}
		}()
	}
}

// Run reads from src and forwards to all subscribers. Call in a goroutine; stops when src is closed.
func (b *Broadcaster) Run(src <-chan *event.Event) {
	for ev := range src {
		b.Send(ev)
	}

	// ソースチャネルが閉じられたので、すべての購読者チャネルを閉じる
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, ch)
		// 注意: ここでチャネルを閉じた後、Unsubscribeが呼ばれると再度閉じようとしてpanicする可能性がある
	}
}
