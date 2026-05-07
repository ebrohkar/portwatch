// Package suppress implements a time-bounded suppression list for
// portwatch alerts. Suppressions can target a specific port+event
// combination or all events on a port (using an empty event string).
//
// Typical usage:
//
//	list := suppress.New()
//	list.Add(8080, "open", 30*time.Minute)
//
//	if !list.IsSuppressed(port, event) {
//		// forward alert to notifier
//	}
//
// Expired entries are pruned lazily on each IsSuppressed or Len call,
// keeping memory usage proportional to active suppressions.
package suppress
