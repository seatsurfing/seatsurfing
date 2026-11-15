package test

import (
	"sync"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestKeyedQueueRunsInOrderPerKey(t *testing.T) {
	q := &KeyedQueue{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	order := []int{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		q.Enqueue("booking", func() {
			defer wg.Done()
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
		})
	}
	wg.Wait()
	CheckTestInt(t, 100, len(order))
	for i, v := range order {
		CheckTestInt(t, i, v)
	}
}

func TestKeyedQueueWaitsForRunningFunction(t *testing.T) {
	q := &KeyedQueue{}
	release := make(chan struct{})
	done := make(chan string, 2)
	q.Enqueue("booking", func() {
		<-release
		done <- "first"
	})
	q.Enqueue("booking", func() {
		done <- "second"
	})
	select {
	case v := <-done:
		t.Fatalf("expected nothing to finish before release, got %s", v)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	CheckTestString(t, "first", <-done)
	CheckTestString(t, "second", <-done)
}

func TestKeyedQueueDifferentKeysRunConcurrently(t *testing.T) {
	q := &KeyedQueue{}
	release := make(chan struct{})
	defer close(release)
	done := make(chan struct{})
	q.Enqueue("a", func() {
		<-release
	})
	q.Enqueue("b", func() {
		close(done)
	})
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("function for key b was blocked by key a")
	}
}

func TestKeyedQueueRestartsAfterDrain(t *testing.T) {
	q := &KeyedQueue{}
	done := make(chan int, 1)
	q.Enqueue("booking", func() { done <- 1 })
	CheckTestInt(t, 1, <-done)
	// Give the worker time to drain and exit before enqueueing again.
	time.Sleep(10 * time.Millisecond)
	q.Enqueue("booking", func() { done <- 2 })
	select {
	case v := <-done:
		CheckTestInt(t, 2, v)
	case <-time.After(time.Second):
		t.Fatal("function enqueued after drain never ran")
	}
}
