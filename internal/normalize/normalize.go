// Package normalize provides port-event normalization for alert pipelines.
// It canonicalizes alert fields to ensure consistent downstream processing.
package normalize

import (
	"fmt"
	"strings"

	"github.com/user/portwatch/internal/alert"
)

// Normalizer canonicalizes alert fields before they enter the pipeline.
type Normalizer struct {
	defaultSeverity string
	defaultEvent    string
}

// Option configures a Normalizer.
type Option func(*Normalizer)

// WithDefaultSeverity sets the severity applied when an alert's severity is empty.
func WithDefaultSeverity(s string) Option {
	return func(n *Normalizer) {
		n.defaultSeverity = s
	}
}

// WithDefaultEvent sets the event applied when an alert's event is empty.
func WithDefaultEvent(e string) Option {
	return func(n *Normalizer) {
		n.defaultEvent = e
	}
}

// New returns a Normalizer with the given options.
// Defaults: severity="info", event="open".
func New(opts ...Option) *Normalizer {
	n := &Normalizer{
		defaultSeverity: "info",
		defaultEvent:    "open",
	}
	for _, o := range opts {
		o(n)
	}
	return n
}

// Apply returns a copy of a with all fields normalized.
// - Event and Severity are lower-cased and defaulted when blank.
// - Message is trimmed of surrounding whitespace.
// - Port must be in [1, 65535]; an error is returned otherwise.
func (n *Normalizer) Apply(a alert.Alert) (alert.Alert, error) {
	if a.Port < 1 || a.Port > 65535 {
		return alert.Alert{}, fmt.Errorf("normalize: port %d out of range [1, 65535]", a.Port)
	}

	a.Event = strings.ToLower(strings.TrimSpace(a.Event))
	if a.Event == "" {
		a.Event = n.defaultEvent
	}

	a.Severity = strings.ToLower(strings.TrimSpace(a.Severity))
	if a.Severity == "" {
		a.Severity = n.defaultSeverity
	}

	a.Message = strings.TrimSpace(a.Message)

	return a, nil
}

// Stage wraps Normalizer as a pipeline stage. Alerts that fail normalization
// are dropped (Allow returns false).
type Stage struct {
	n *Normalizer
}

// NewStage returns a Stage backed by n. Panics if n is nil.
func NewStage(n *Normalizer) *Stage {
	if n == nil {
		panic("normalize: NewStage requires a non-nil Normalizer")
	}
	return &Stage{n: n}
}

// Allow normalizes a in place and returns true when normalization succeeds.
func (s *Stage) Allow(a *alert.Alert) bool {
	norm, err := s.n.Apply(*a)
	if err != nil {
		return false
	}
	*a = norm
	return true
}

func (s *Stage) String() string { return "normalize" }
