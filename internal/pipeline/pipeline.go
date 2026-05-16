// Package pipeline wires alert-processing stages together into a
// linear chain: filter → debounce → rate-limit → suppress → notify.
package pipeline

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/debounce"
	"github.com/user/portwatch/internal/filter"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/ratelimit"
	"github.com/user/portwatch/internal/suppress"
)

// Stage is a single processing step.  It returns false when the alert
// should be dropped.
type Stage func(a alert.Alert) bool

// Pipeline runs an alert through an ordered list of stages and, if all
// stages pass, forwards the alert to the notifier.
type Pipeline struct {
	stages   []Stage
	notifier *notifier.Notifier
}

// New constructs a Pipeline from the supplied stages and notifier.
// It panics when notifier is nil.
func New(n *notifier.Notifier, stages ...Stage) *Pipeline {
	if n == nil {
		panic("pipeline: notifier must not be nil")
	}
	return &Pipeline{stages: stages, notifier: n}
}

// Run passes a through every stage in order.  The first stage that
// returns false short-circuits evaluation and the alert is dropped.
// If all stages pass, the alert is sent via the notifier.
// Run returns an error only when the notifier fails.
func (p *Pipeline) Run(_ context.Context, a alert.Alert) error {
	for _, s := range p.stages {
		if !s(a) {
			return nil
		}
	}
	return p.notifier.Send(a)
}

// FromParts is a convenience constructor that builds the standard
// filter → debounce → ratelimit → suppress chain and returns a ready
// Pipeline.
func FromParts(
	n *notifier.Notifier,
	f *filter.Filter,
	d *debounce.Debouncer,
	rl *ratelimit.RateLimiter,
	su *suppress.Suppressor,
) *Pipeline {
	if n == nil {
		panic("pipeline: notifier must not be nil")
	}

	stages := []Stage{
		func(a alert.Alert) bool { return f.Allow(a.Port) },
		func(a alert.Alert) bool {
			ok, err := d.Allow(a.Port, a.Event)
			if err != nil {
				_ = fmt.Sprintf("debounce: %v", err) // best-effort
			}
			return ok
		},
		func(a alert.Alert) bool { return rl.Allow(a.Port) },
		func(a alert.Alert) bool { return !su.IsSuppressed(a.Port, a.Event) },
	}
	return New(n, stages...)
}
