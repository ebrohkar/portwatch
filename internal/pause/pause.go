// Package pause provides a mechanism to temporarily pause alert delivery
// for a specific port and event type for a fixed duration.
package pause

import (
	"fmt"
	"sync"
	"time"
)

// Clock abstracts time for testability.
type Clock func() time.Time

// entry holds the expiry time for a paused key.
type entry struct {
	expiry time.Time
}

// Store tracks paused port/event combinations.
type Store struct {
	mu    sync.Mutex
	items map[string]entry
	clock Clock
}

// New creates a new Store using the real clock.
func New() *Store {
	return withClock(time.Now)
}

func withClock(c Clock) *Store {
	return &Store{
		items: make(map[string]entry),
		clock: c,
	}
}

// Pause marks the given port+event as paused until now+duration.
// Calling Pause on an already-paused key extends the expiry.
func (s *Store) Pause(port int, event string, d time.Duration) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("pause: invalid port %d", port)
	}
	if event == "" {
		return fmt.Errorf("pause: event must not be empty")
	}
	if d <= 0 {
		return fmt.Errorf("pause: duration must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key(port, event)] = entry{expiry: s.clock().Add(d)}
	return nil
}

// IsPaused reports whether the port+event combination is currently paused.
func (s *Store) IsPaused(port int, event string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[key(port, event)]
	if !ok {
		return false
	}
	if s.clock().After(e.expiry) {
		delete(s.items, key(port, event))
		return false
	}
	return true
}

// Resume removes a pause entry immediately, regardless of expiry.
func (s *Store) Resume(port int, event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key(port, event))
}

// Len returns the number of active (non-expired) pause entries.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	count := 0
	for k, e := range s.items {
		if now.After(e.expiry) {
			delete(s.items, k)
		} else {
			count++
		}
	}
	return count
}

func key(port int, event string) string {
	return fmt.Sprintf("%d:%s", port, event)
}
