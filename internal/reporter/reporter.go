// Package reporter ties together scanning, rule evaluation, and notification
// into a single Report cycle used by the daemon.
package reporter

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/rules"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

// Reporter orchestrates a single scan-evaluate-notify cycle.
type Reporter struct {
	scanner  *scanner.Scanner
	ruleSet  *rules.RuleSet
	store    *state.Store
	notifier *notifier.Notifier
}

// New creates a Reporter wired to the provided dependencies.
func New(sc *scanner.Scanner, rs *rules.RuleSet, st *state.Store, n *notifier.Notifier) *Reporter {
	return &Reporter{
		scanner:  sc,
		ruleSet:  rs,
		store:    st,
		notifier: n,
	}
}

// Run performs one full scan cycle: scan ports, diff against previous state,
// evaluate rules, emit alerts for any violations, then persist the new state.
func (r *Reporter) Run(ctx context.Context) error {
	openPorts, err := r.scanner.Scan(ctx)
	if err != nil {
		return fmt.Errorf("reporter: scan failed: %w", err)
	}

	prev := r.store.Current()
	added, removed := state.Diff(prev, openPorts)

	for _, port := range added {
		alerts := r.ruleSet.Evaluate(port, rules.EventOpened)
		for _, a := range alerts {
			if err := r.notifier.Send(a); err != nil {
				return fmt.Errorf("reporter: notify failed: %w", err)
			}
		}
	}

	for _, port := range removed {
		alerts := r.ruleSet.Evaluate(port, rules.EventClosed)
		for _, a := range alerts {
			if err := r.notifier.Send(a); err != nil {
				return fmt.Errorf("reporter: notify failed: %w", err)
			}
		}
	}

	if err := r.store.Save(openPorts); err != nil {
		return fmt.Errorf("reporter: save state failed: %w", err)
	}

	return nil
}

// AlertCount returns the number of alerts that would be generated for the
// provided port sets without persisting any state — useful for dry-runs.
func (r *Reporter) AlertCount(added, removed []int) int {
	count := 0
	for _, p := range added {
		count += len(r.ruleSet.Evaluate(p, rules.EventOpened))
	}
	for _, p := range removed {
		count += len(r.ruleSet.Evaluate(p, rules.EventClosed))
	}
	return count
}

// ensure alert import is used
var _ = alert.New
