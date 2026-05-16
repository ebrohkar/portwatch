// Package dispatcher routes alerts to one or more named notifier channels
// based on configurable severity-to-channel mappings.
package dispatcher

import (
	"errors"
	"fmt"

	"github.com/example/portwatch/internal/alert"
)

// Sender is the interface satisfied by a notifier.
type Sender interface {
	Send(a alert.Alert) error
}

// Rule maps a minimum severity level to a named channel.
type Rule struct {
	// MinSeverity is the lowest severity that triggers this rule (inclusive).
	// Accepted values: "low", "medium", "high", "critical".
	MinSeverity string
	// Channel is the key into the channels map supplied to New.
	Channel string
}

// Dispatcher routes an alert to every channel whose rule is satisfied.
type Dispatcher struct {
	channels map[string]Sender
	rules    []Rule
}

var severityRank = map[string]int{
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}

// New creates a Dispatcher. channels must not be nil; each Rule.Channel must
// exist in channels and each Rule.MinSeverity must be a known level.
func New(channels map[string]Sender, rules []Rule) (*Dispatcher, error) {
	if len(channels) == 0 {
		return nil, errors.New("dispatcher: channels map must not be empty")
	}
	for i, r := range rules {
		if _, ok := severityRank[r.MinSeverity]; !ok {
			return nil, fmt.Errorf("dispatcher: rule %d: unknown severity %q", i, r.MinSeverity)
		}
		if _, ok := channels[r.Channel]; !ok {
			return nil, fmt.Errorf("dispatcher: rule %d: unknown channel %q", i, r.Channel)
		}
	}
	return &Dispatcher{channels: channels, rules: rules}, nil
}

// Dispatch sends a to every channel whose MinSeverity is ≤ the alert's
// severity. Errors from individual senders are joined and returned together.
func (d *Dispatcher) Dispatch(a alert.Alert) error {
	rank, ok := severityRank[a.Severity]
	if !ok {
		return fmt.Errorf("dispatcher: unknown alert severity %q", a.Severity)
	}

	seen := make(map[string]bool)
	var errs []error
	for _, r := range d.rules {
		if severityRank[r.MinSeverity] <= rank && !seen[r.Channel] {
			seen[r.Channel] = true
			if err := d.channels[r.Channel].Send(a); err != nil {
				errs = append(errs, fmt.Errorf("channel %q: %w", r.Channel, err))
			}
		}
	}
	return errors.Join(errs...)
}
