package grpc

import (
	"testing"
	"time"

	"docksphinx/internal/event"
)

func TestBroadcasterRunClosesSubscribersWhenSourceCloses(t *testing.T) {
	b := NewBroadcaster()
	sub, _ := b.Subscribe()
	src := make(chan *event.Event)

	done := make(chan struct{})
	go func() {
		b.Run(src)
		close(done)
	}()

	close(src)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Run did not return after source channel close")
	}

	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("subscriber channel should be closed")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("subscriber channel was not closed")
	}
}

// 複数の購読者がいる場合のテストを追加
func TestBroadcasterRunClosesMultipleSubscribers(t *testing.T) {
	b := NewBroadcaster()
	sub1, _ := b.Subscribe()
	sub2, _ := b.Subscribe()
	src := make(chan *event.Event)

	go b.Run(src)
	close(src)
	time.Sleep(100 * time.Millisecond) // Runが完了するまで少し待つ

	// 両方のチャネルが閉じられているか確認
	if _, ok := <-sub1; ok {
		t.Error("sub1 should be closed")
	}
	if _, ok := <-sub2; ok {
		t.Error("sub2 should be closed")
	}
}

// Runがチャネルを閉じた後、Unsubscribeが呼ばれるとpanicする可能性があるテスト
func TestBroadcasterDoubleClosePanic(t *testing.T) {
	b := NewBroadcaster()
	_, unsub := b.Subscribe()
	src := make(chan *event.Event)

	go b.Run(src)
	close(src)
	time.Sleep(50 * time.Millisecond) // Runがチャネルを閉じるまで待つ

	// Runが既にチャネルを閉じている状態でUnsubscribeを呼ぶとpanicする可能性
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Panic occurred as expected: %v", r)
		}
	}()
	unsub() // 既に閉じられたチャネルを再度閉じようとする
}
