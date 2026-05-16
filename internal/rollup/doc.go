// Package rollup provides a threshold-based alert aggregation mechanism for
// portwatch. When the same (port, event) combination is observed repeatedly
// within a sliding time window, the individual alerts are suppressed and a
// single summary alert is emitted once the configured threshold is reached.
//
// This prevents alert storms — for example, a flapping port that repeatedly
// opens and closes will generate one rolled-up alert rather than dozens of
// individual notifications.
//
// Usage:
//
//	r := rollup.New(5, 30*time.Second)
//	if summary, ok := r.Add(a); ok {
//		// emit summary
//	}
package rollup
