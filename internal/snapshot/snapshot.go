// Package snapshot provides point-in-time captures of open port sets,
// enabling before/after comparisons across scan cycles.
package snapshot

import (
	"fmt"
	"sort"
	"time"
)

// Snapshot holds a sorted list of open ports captured at a specific time.
type Snapshot struct {
	Ports     []int     `json:"ports"`
	CapturedAt time.Time `json:"captured_at"`
}

// New creates a Snapshot from the provided port list, deduplicating and
// sorting entries. Returns an error if any port is out of the valid range.
func New(ports []int) (*Snapshot, error) {
	seen := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("snapshot: port %d out of range [1, 65535]", p)
		}
		seen[p] = struct{}{}
	}

	unique := make([]int, 0, len(seen))
	for p := range seen {
		unique = append(unique, p)
	}
	sort.Ints(unique)

	return &Snapshot{
		Ports:      unique,
		CapturedAt: time.Now().UTC(),
	}, nil
}

// Contains reports whether the given port is present in the snapshot.
func (s *Snapshot) Contains(port int) bool {
	for _, p := range s.Ports {
		if p == port {
			return true
		}
	}
	return false
}

// Added returns ports that are in s but not in prev (newly opened).
func (s *Snapshot) Added(prev *Snapshot) []int {
	if prev == nil {
		return append([]int(nil), s.Ports...)
	}
	var result []int
	for _, p := range s.Ports {
		if !prev.Contains(p) {
			result = append(result, p)
		}
	}
	return result
}

// Removed returns ports that are in prev but not in s (newly closed).
func (s *Snapshot) Removed(prev *Snapshot) []int {
	if prev == nil {
		return nil
	}
	var result []int
	for _, p := range prev.Ports {
		if !s.Contains(p) {
			result = append(result, p)
		}
	}
	return result
}

// Len returns the number of open ports in the snapshot.
func (s *Snapshot) Len() int { return len(s.Ports) }
