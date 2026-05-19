// Package scorecard provides a per-port health scoring mechanism for portwatch.
//
// Scores are incremented each time an alert is recorded for a port and decay
// exponentially over a configurable half-life period. This allows the daemon
// to distinguish ports that are consistently problematic from those that
// produced only a transient alert.
//
// Example usage:
//
//	sc := scorecard.New(5 * time.Minute)
//	sc.Record(8080, 1.0)
//	fmt.Println(sc.Get(8080)) // slightly below 1.0 after a moment
package scorecard
