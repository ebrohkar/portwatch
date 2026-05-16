// Package correlator groups related alerts into correlated events
// based on time proximity and shared attributes such as port range or event type.
package correlator

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/alert"
)

// Group holds a set of alerts that have been correlated together.
type Group struct {
	ID     string
	Alerts []alert.Alert
	First  time.Time
	Last   time.Time
}

// Correlator accumulates alerts and groups them when they share the same
// event type and arrive within the configured window duration.
type Correlator struct {
	mu     sync.Mutex
	window time.Duration
	groups map[string]*Group
	clock  func() time.Time
}

// New returns a Correlator with the given grouping window.
// Panics if window is zero.
func New(window time.Duration) *Correlator {
	if window == 0 {
		panic("correlator: window must be greater than zero")
	}
	return &Correlator{
		window: window,
		groups: make(map[string]*Group),
		clock:  time.Now,
	}
}

// withClock returns a Correlator using the provided clock function (for testing).
func withClock(window time.Duration, clock func() time.Time) *Correlator {
	c := New(window)
	c.clock = clock
	return c
}

// Add inserts an alert into an existing group if one exists for the same event
// type within the active window, otherwise it starts a new group.
// Returns the group ID the alert was assigned to.
func (c *Correlator) Add(a alert.Alert) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	key := a.Event

	if g, ok := c.groups[key]; ok {
		if now.Sub(g.Last) <= c.window {
			g.Alerts = append(g.Alerts, a)
			g.Last = now
			return g.ID
		}
	}

	id := key + "-" + now.Format("20060102150405.000000000")
	c.groups[key] = &Group{
		ID:     id,
		Alerts: []alert.Alert{a},
		First:  now,
		Last:   now,
	}
	return id
}

// Flush returns all current groups and resets internal state.
func (c *Correlator) Flush() []Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]Group, 0, len(c.groups))
	for _, g := range c.groups {
		out = append(out, *g)
	}
	c.groups = make(map[string]*Group)
	return out
}

// Len returns the number of active groups.
func (c *Correlator) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.groups)
}
