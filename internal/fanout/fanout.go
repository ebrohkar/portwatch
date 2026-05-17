// Package fanout broadcasts a single alert to multiple independent pipelines.
package fanout

import (
	"context"
	"fmt"

	"github.com/example/portwatch/internal/alert"
)

// Sink is any destination that can receive an alert.
type Sink interface {
	Send(ctx context.Context, a alert.Alert) error
}

// Fanout distributes each alert to every registered sink.
type Fanout struct {
	sinks []Sink
}

// New returns a Fanout that broadcasts to the provided sinks.
// It panics if no sinks are provided.
func New(sinks ...Sink) *Fanout {
	if len(sinks) == 0 {
		panic("fanout: at least one sink is required")
	}
	return &Fanout{sinks: sinks}
}

// Send delivers the alert to every sink. All sinks are attempted even if
// earlier ones fail. The first non-nil error encountered is returned.
func (f *Fanout) Send(ctx context.Context, a alert.Alert) error {
	var first error
	for i, s := range f.sinks {
		if err := s.Send(ctx, a); err != nil {
			if first == nil {
				first = fmt.Errorf("fanout: sink %d: %w", i, err)
			}
		}
	}
	return first
}

// Len returns the number of registered sinks.
func (f *Fanout) Len() int { return len(f.sinks) }
