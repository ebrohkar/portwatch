// Package acknowledge provides a thread-safe store for tracking
// acknowledged alerts so that repeated notifications for known issues
// can be suppressed until the acknowledgement expires or is revoked.
package acknowledge

import (
	"sync"
	"time"
)

// clock allows time to be faked in tests.
type clock func() time.Time

// entry holds the expiry time for a single acknowledgement.
type entry struct {
	expiresAt time.Time
}

// Store tracks acknowledged (port, event) pairs.
type Store struct {
	mu    sync.RWMutex
	items map[string]entry
	now   clock
}

// New returns a Store that uses the real wall clock.
func New() *Store {
	return &Store{
		items: make(map[string]entry),
		now:   time.Now,
	}
}

// withClock returns a Store with a custom clock (for testing).
func withClock(c clock) *Store {
	return &Store{
		items: make(map[string]entry),
		now:   c,
	}
}

func key(port int, event string) string {
	return event + ":" + itoa(port)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 8)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// Acknowledge records that the given (port, event) pair is acknowledged
// for the specified duration. A zero or negative duration is a no-op.
func (s *Store) Acknowledge(port int, event string, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key(port, event)] = entry{expiresAt: s.now().Add(ttl)}
}

// IsAcknowledged reports whether (port, event) is currently acknowledged.
func (s *Store) IsAcknowledged(port int, event string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[key(port, event)]
	if !ok {
		return false
	}
	return s.now().Before(e.expiresAt)
}

// Revoke removes an acknowledgement before it would naturally expire.
func (s *Store) Revoke(port int, event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key(port, event))
}

// Len returns the number of active (non-expired) acknowledgements.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	count := 0
	for _, e := range s.items {
		if now.Before(e.expiresAt) {
			count++
		}
	}
	return count
}
