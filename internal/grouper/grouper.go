// Package grouper batches alerts into named groups based on port ranges or
// explicit port lists, enabling downstream consumers to process related alerts
// together rather than one-by-one.
package grouper

import (
	"errors"
	"fmt"
	"sync"

	"github.com/example/portwatch/internal/alert"
)

// Group is a named collection of alerts.
type Group struct {
	Name   string
	Alerts []alert.Alert
}

// Rule maps a group name to the set of ports that belong to it.
type Rule struct {
	Name  string
	Ports map[int]struct{}
}

// Grouper assigns incoming alerts to named groups.
type Grouper struct {
	mu    sync.Mutex
	rules []Rule
	groups map[string]*Group
}

// New creates a Grouper from the provided rules.
// Each rule must have a non-empty name and at least one port.
func New(rules []Rule) (*Grouper, error) {
	if len(rules) == 0 {
		return nil, errors.New("grouper: at least one rule is required")
	}
	for i, r := range rules {
		if r.Name == "" {
			return nil, fmt.Errorf("grouper: rule %d has empty name", i)
		}
		if len(r.Ports) == 0 {
			return nil, fmt.Errorf("grouper: rule %q has no ports", r.Name)
		}
	}
	groups := make(map[string]*Group, len(rules))
	for _, r := range rules {
		groups[r.Name] = &Group{Name: r.Name}
	}
	return &Grouper{rules: rules, groups: groups}, nil
}

// Add places the alert into every matching group.
// An alert may belong to multiple groups if its port appears in several rules.
func (g *Grouper) Add(a alert.Alert) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, r := range g.rules {
		if _, ok := r.Ports[a.Port]; ok {
			g.groups[r.Name].Alerts = append(g.groups[r.Name].Alerts, a)
		}
	}
}

// Flush returns all groups that contain at least one alert and resets them.
func (g *Grouper) Flush() []Group {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []Group
	for name, grp := range g.groups {
		if len(grp.Alerts) == 0 {
			continue
		}
		out = append(out, Group{Name: grp.Name, Alerts: grp.Alerts})
		g.groups[name] = &Group{Name: name}
	}
	return out
}

// Len returns the total number of buffered alerts across all groups.
func (g *Grouper) Len() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	total := 0
	for _, grp := range g.groups {
		total += len(grp.Alerts)
	}
	return total
}
