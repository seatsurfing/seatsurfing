package util

import "sync"

// KeyedQueue runs functions asynchronously, one at a time per key, in the
// order they were enqueued. Functions for different keys run concurrently.
// The zero value is ready to use.
type KeyedQueue struct {
	mu      sync.Mutex
	pending map[string][]func()
}

// Enqueue schedules f to run after all functions previously enqueued for key
// have finished. It never blocks on f.
func (q *KeyedQueue) Enqueue(key string, f func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.pending == nil {
		q.pending = make(map[string][]func())
	}
	queue, running := q.pending[key]
	q.pending[key] = append(queue, f)
	if !running {
		go q.run(key)
	}
}

// run works off the functions for key until none are left. A key is present
// in pending exactly as long as its worker is running.
func (q *KeyedQueue) run(key string) {
	for {
		q.mu.Lock()
		queue := q.pending[key]
		if len(queue) == 0 {
			delete(q.pending, key)
			q.mu.Unlock()
			return
		}
		f := queue[0]
		q.pending[key] = queue[1:]
		q.mu.Unlock()
		f()
	}
}
