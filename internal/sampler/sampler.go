// Package sampler provides probabilistic sampling for alerts,
// allowing a configurable fraction of alerts to pass through
// based on port and event type.
package sampler

import (
	"fmt"
	"math/rand"
	"sync"
)

// RandFunc is a function that returns a float64 in [0.0, 1.0).
type RandFunc func() float64

// Sampler decides whether an alert should be forwarded based on
// a per-port or global sampling rate.
type Sampler struct {
	mu       sync.Mutex
	rates    map[int]float64 // port-specific rates
	default_ float64
	randf    RandFunc
}

// New creates a Sampler with the given default sampling rate.
// rate must be in (0.0, 1.0]. A rate of 1.0 passes all alerts.
func New(defaultRate float64) (*Sampler, error) {
	if defaultRate <= 0 || defaultRate > 1.0 {
		return nil, fmt.Errorf("sampler: default rate must be in (0, 1], got %v", defaultRate)
	}
	return &Sampler{
		rates:    make(map[int]float64),
		default_: defaultRate,
		randf:    rand.Float64,
	}, nil
}

// withRand returns a Sampler using the provided RandFunc (for testing).
func withRand(defaultRate float64, randf RandFunc) (*Sampler, error) {
	s, err := New(defaultRate)
	if err != nil {
		return nil, err
	}
	s.randf = randf
	return s, nil
}

// SetPortRate sets a sampling rate for a specific port.
// rate must be in (0.0, 1.0].
func (s *Sampler) SetPortRate(port int, rate float64) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("sampler: invalid port %d", port)
	}
	if rate <= 0 || rate > 1.0 {
		return fmt.Errorf("sampler: rate must be in (0, 1], got %v", rate)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rates[port] = rate
	return nil
}

// Allow returns true if the alert for the given port should be forwarded.
func (s *Sampler) Allow(port int) bool {
	s.mu.Lock()
	rate, ok := s.rates[port]
	if !ok {
		rate = s.default_
	}
	s.mu.Unlock()
	return s.randf() < rate
}

// Rate returns the effective sampling rate for a port.
func (s *Sampler) Rate(port int) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rates[port]; ok {
		return r
	}
	return s.default_
}
