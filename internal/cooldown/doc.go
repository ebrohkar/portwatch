// Package cooldown implements a per-port alert cooldown store for portwatch.
//
// It ensures that once an alert fires for a given port, subsequent alerts
// for that same port are suppressed until the configured quiet period
// (cooldown duration) has elapsed. This prevents alert storms when a port
// flaps or is repeatedly scanned within a short window.
//
// Usage:
//
//	store := cooldown.New(30 * time.Second)
//	if store.Allow(8080) {
//		// send alert
//	}
package cooldown
