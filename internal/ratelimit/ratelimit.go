// Package ratelimit provides a token-bucket style rate limiter for
// suppressing repeated alerts for the same port within a configurable window.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks per-port alert counts and suppresses duplicates within a
// sliding window.
type Limiter struct {
	mu      sync.Mutex
	window  time.Duration
	maxHits int
	buckets map[int]*bucket
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// New returns a Limiter that allows at most maxHits alerts per port within
// the given window duration. maxHits < 1 is treated as 1.
func New(window time.Duration, maxHits int) *Limiter {
	if maxHits < 1 {
		maxHits = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{
		window:  window,
		maxHits: maxHits,
		buckets: make(map[int]*bucket),
	}
}

// Allow reports whether an alert for the given port should be forwarded.
// It returns true the first maxHits times within the window, false thereafter.
func (l *Limiter) Allow(port int) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[port]
	if !ok || now.After(b.windowEnd) {
		l.buckets[port] = &bucket{count: 1, windowEnd: now.Add(l.window)}
		return true
	}

	if b.count < l.maxHits {
		b.count++
		return true
	}
	return false
}

// Reset clears the rate-limit state for a specific port.
func (l *Limiter) Reset(port int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, port)
}

// Purge removes all expired buckets to reclaim memory.
func (l *Limiter) Purge() {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	for port, b := range l.buckets {
		if now.After(b.windowEnd) {
			delete(l.buckets, port)
		}
	}
}
