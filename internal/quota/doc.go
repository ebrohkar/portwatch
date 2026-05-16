// Package quota provides a per-port alert quota enforcer for portwatch.
//
// A Quota tracks how many alerts have been emitted for each monitored port
// within a configurable rolling time window. Once a port reaches the
// configured maximum, subsequent Allow calls return false until the window
// expires and the counter resets automatically.
//
// Usage:
//
//	q := quota.New(5, time.Minute)
//	if q.Allow(8080) {
//	    // emit alert
//	}
package quota
