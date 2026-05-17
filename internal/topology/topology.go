// Package topology maps ports to logical service names and tiers,
// enabling richer alert context when a port change is detected.
package topology

import (
	"errors"
	"fmt"
	"strings"
)

// Tier represents the logical layer a service belongs to.
type Tier string

const (
	TierFrontend Tier = "frontend"
	TierBackend  Tier = "backend"
	TierDatabase Tier = "database"
	TierInternal Tier = "internal"
)

var validTiers = map[Tier]struct{}{
	TierFrontend: {},
	TierBackend:  {},
	TierDatabase: {},
	TierInternal: {},
}

// Entry describes a single port-to-service mapping.
type Entry struct {
	Port    int
	Service string
	Tier    Tier
}

// Topology holds the full port-to-service map.
type Topology struct {
	entries map[int]Entry
}

// New builds a Topology from a slice of entries.
// Returns an error if any entry is invalid.
func New(entries []Entry) (*Topology, error) {
	if len(entries) == 0 {
		return nil, errors.New("topology: at least one entry is required")
	}
	m := make(map[int]Entry, len(entries))
	for _, e := range entries {
		if e.Port < 1 || e.Port > 65535 {
			return nil, fmt.Errorf("topology: invalid port %d", e.Port)
		}
		if strings.TrimSpace(e.Service) == "" {
			return nil, fmt.Errorf("topology: empty service name for port %d", e.Port)
		}
		if _, ok := validTiers[e.Tier]; !ok {
			return nil, fmt.Errorf("topology: unknown tier %q for port %d", e.Tier, e.Port)
		}
		m[e.Port] = e
	}
	return &Topology{entries: m}, nil
}

// Lookup returns the Entry for the given port and true if found,
// or a zero Entry and false if the port is not registered.
func (t *Topology) Lookup(port int) (Entry, bool) {
	e, ok := t.entries[port]
	return e, ok
}

// Len returns the number of registered entries.
func (t *Topology) Len() int {
	return len(t.entries)
}
