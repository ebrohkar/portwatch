// Package window provides a sliding time-window counter used to track
// event frequency over a rolling duration (e.g. alerts per minute).
package window

import (
	"sync"
	"time"
)

// Clock allows injecting a custom time source for testing.
type Clock func() time.Time

// Window is a thread-safe sliding time-window counter keyed by an integer
// (typically a port number).
type Window struct {
	mu       sync.Mutex
	size     time.Duration
	clock    Clock
	buckets  map[int][]time.Time
}

// New returns a Window with the given duration. Panics if size is zero.
func New(size time.Duration) *Window {
	if size <= 0 {
		panic("window: size must be greater than zero")
	}
	return &Window{
		size:    size,
		clock:   time.Now,
		buckets: make(map[int][]time.Time),
	}
}

// WithClock returns a copy of w that uses the provided clock function.
func (w *Window) WithClock(c Clock) *Window {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clock = c
	return w
}

// Add records one event for the given key at the current time and returns
// the total number of events within the window after the addition.
func (w *Window) Add(key int) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.clock()
	w.buckets[key] = append(w.evict(key, now), now)
	return len(w.buckets[key])
}

// Count returns the number of events recorded for key within the window.
func (w *Window) Count(key int) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.clock()
	w.buckets[key] = w.evict(key, now)
	return len(w.buckets[key])
}

// Reset clears all recorded events for key.
func (w *Window) Reset(key int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.buckets, key)
}

// Keys returns all keys that currently have at least one event within the
// window. Expired entries are evicted before the check.
func (w *Window) Keys() []int {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.clock()
	keys := make([]int, 0, len(w.buckets))
	for key := range w.buckets {
		w.buckets[key] = w.evict(key, now)
		if len(w.buckets[key]) > 0 {
			keys = append(keys, key)
		} else {
			delete(w.buckets, key)
		}
	}
	return keys
}

// evict removes timestamps older than the window size and returns the
// remaining slice. Caller must hold w.mu.
func (w *Window) evict(key int, now time.Time) []time.Time {
	cutoff := now.Add(-w.size)
	ts := w.buckets[key]
	i := 0
	for i < len(ts) && ts[i].Before(cutoff) {
		i++
	}
	return ts[i:]
}
