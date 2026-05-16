// Package mute implements time-bounded alert muting for portwatch.
//
// A mute window silences all alerts for a given port until its expiry
// time. Windows are registered via Add and queried via IsMuted.
// Expired windows are pruned lazily on each IsMuted call to avoid
// background goroutines.
//
// Typical usage during a maintenance window:
//
//	store := mute.New()
//	store.Add(8080, 2*time.Hour, "scheduled maintenance")
//
//	// later, inside the alert pipeline:
//	if store.IsMuted(alert.Port) {
//	    return // drop the alert
//	}
package mute
