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

func TestBroadcasterSendDoesNotBlockWhenSubscriberIsSlow(t *testing.T) {
	b := NewBroadcaster()
	ch, unsub := b.Subscribe()
	defer unsub()

	// Fill subscriber buffer without receiving.
	for i := 0; i < subscriberChanBuf; i++ {
		b.Send(event.NewEvent(event.EventTypeStarted, "id", "name", "img"))
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		b.Send(event.NewEvent(event.EventTypeStarted, "id2", "name2", "img2"))
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Send blocked on full subscriber channel")
	}

	// Drain one item to avoid goroutine leak in test.
	select {
	case <-ch:
	default:
	}
}

func TestBroadcasterNilSafetyContracts(t *testing.T) {
	var b *Broadcaster

	ch, unsub := b.Subscribe()
	unsub() // should not panic
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected closed channel from nil broadcaster subscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected nil broadcaster subscribe channel to be closed immediately")
	}

	b.Unsubscribe(nil) // should not panic
	b.Send(nil)        // should not panic

	done := make(chan struct{})
	go func() {
		defer close(done)
		b.Run(context.Background(), make(chan *event.Event))
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected nil broadcaster run to return immediately")
	}
}
