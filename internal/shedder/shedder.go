// Package shedder implements load shedding for the alert pipeline.
// When the number of in-flight alerts exceeds a configured ceiling the
// shedder drops new arrivals and increments a counter so operators can
// observe backpressure in metrics.
package shedder

import (
	"errors"
	"fmt"
	"sync/atomic"
)

// ErrShedded is returned by Allow when the load ceiling is exceeded.
var ErrShedded = errors.New("shedder: load ceiling exceeded, alert dropped")

// Shedder tracks the number of concurrent in-flight alerts and rejects
// new ones once the ceiling is reached.
type Shedder struct {
	max     int64
	inflight atomic.Int64
	dropped  atomic.Int64
}

// New creates a Shedder with the given maximum concurrency ceiling.
// It panics when max is less than 1.
func New(max int) *Shedder {
	if max < 1 {
		panic("shedder: max must be >= 1")
	}
	return &Shedder{max: int64(max)}
}

// Allow returns nil and increments the in-flight counter when capacity is
// available.  The caller MUST call Done when processing is complete.
// Allow returns ErrShedded without touching the counter when the ceiling
// is already reached.
func (s *Shedder) Allow() error {
	if s.inflight.Load() >= s.max {
		s.dropped.Add(1)
		return ErrShedded
	}
	s.inflight.Add(1)
	return nil
}

// Done decrements the in-flight counter.  It must be called exactly once
// for every successful Allow call.
func (s *Shedder) Done() {
	s.inflight.Add(-1)
}

// Inflight returns the current number of in-flight alerts.
func (s *Shedder) Inflight() int64 {
	return s.inflight.Load()
}

// Dropped returns the cumulative number of alerts dropped since the
// Shedder was created.
func (s *Shedder) Dropped() int64 {
	return s.dropped.Load()
}

// String implements fmt.Stringer.
func (s *Shedder) String() string {
	return fmt.Sprintf("shedder(max=%d inflight=%d dropped=%d)",
		s.max, s.Inflight(), s.Dropped())
}
