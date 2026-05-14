// Package cooldown provides a per-port cooldown tracker that prevents
// repeated alerts from firing until a configurable quiet period has elapsed.
package cooldown

import (
	"sync"
	"time"
)

// Clock abstracts time for testing.
type Clock func() time.Time

// Store tracks the last alert time for each port.
type Store struct {
	mu       sync.Mutex
	cooldown time.Duration
	entries  map[int]time.Time
	clock    Clock
}

// New creates a Store with the given cooldown duration.
// Panics if cooldown is zero or negative.
func New(cooldown time.Duration) *Store {
	return WithClock(cooldown, time.Now)
}

// WithClock creates a Store with an injectable clock for testing.
func WithClock(cooldown time.Duration, clock Clock) *Store {
	if cooldown <= 0 {
		panic("cooldown: duration must be positive")
	}
	if clock == nil {
		panic("cooldown: clock must not be nil")
	}
	return &Store{
		cooldown: cooldown,
		entries:  make(map[int]time.Time),
		clock:    clock,
	}
}

// Allow returns true and records the current time if the port has not fired
// an alert within the cooldown window. Returns false otherwise.
func (s *Store) Allow(port int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	if last, ok := s.entries[port]; ok {
		if now.Sub(last) < s.cooldown {
			return false
		}
	}
	s.entries[port] = now
	return true
}

// Reset clears the cooldown entry for the given port, allowing the next
// alert through immediately.
func (s *Store) Reset(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, port)
}

// Len returns the number of ports currently tracked.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}
