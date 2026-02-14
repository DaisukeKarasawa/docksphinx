package grpc

import (
	"context"
	"testing"
	"time"

	"docksphinx/internal/event"
)

func TestBroadcasterSendAndUnsubscribe(t *testing.T) {
	b := NewBroadcaster()
	ch, unsub := b.Subscribe()

	ev := event.NewEvent(event.EventTypeStarted, "id1", "c1", "img")
	b.Send(ev)

	select {
	case got := <-ch:
		if got == nil || got.ID != ev.ID {
			t.Fatalf("unexpected event received: %#v", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}

	unsub()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed channel after unsubscribe")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("unsubscribe did not close subscriber channel")
	}
}

func TestBroadcasterRunStopsOnContextCancel(t *testing.T) {
	b := NewBroadcaster()
	src := make(chan *event.Event)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		b.Run(ctx, src)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("broadcaster run did not stop on context cancellation")
	}
}
