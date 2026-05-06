// Package filter provides port filtering utilities for portwatch.
// It allows callers to include or exclude specific ports or ranges
// from scan results before evaluation.
package filter

import "fmt"

// Filter holds inclusion and exclusion sets for port numbers.
type Filter struct {
	include map[int]struct{}
	exclude map[int]struct{}
}

// New returns a Filter populated from the provided include and exclude port lists.
// An empty include list means "include all" (exclusions still apply).
func New(include, exclude []int) (*Filter, error) {
	f := &Filter{
		include: make(map[int]struct{}, len(include)),
		exclude: make(map[int]struct{}, len(exclude)),
	}
	for _, p := range include {
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("filter: invalid include port %d", p)
		}
		f.include[p] = struct{}{}
	}
	for _, p := range exclude {
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("filter: invalid exclude port %d", p)
		}
		f.exclude[p] = struct{}{}
	}
	return f, nil
}

// Allow reports whether the given port should be included in results.
// A port is allowed when it is not excluded and either the include list
// is empty (allow-all) or the port is explicitly included.
func (f *Filter) Allow(port int) bool {
	if _, excluded := f.exclude[port]; excluded {
		return false
	}
	if len(f.include) == 0 {
		return true
	}
	_, included := f.include[port]
	return included
}

// Apply returns only the ports from the input slice that pass the filter.
func (f *Filter) Apply(ports []int) []int {
	out := make([]int, 0, len(ports))
	for _, p := range ports {
		if f.Allow(p) {
			out = append(out, p)
		}
	}
	return out
}
