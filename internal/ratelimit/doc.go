// Package ratelimit implements a concurrent-safe, per-port rate limiter used
// by portwatch to suppress alert floods when a port repeatedly transitions
// between open and closed states.
//
// # Design
//
// Each port tracked by the limiter maintains a sliding-window counter. When
// the number of events observed within the configured window duration exceeds
// the configured limit, [Limiter.Allow] returns false, suppressing the alert.
// Counters are stored in a map protected by a sync.Mutex, so the limiter is
// safe for concurrent use by multiple goroutines.
//
// # Usage
//
//	limiter := ratelimit.New(time.Minute, 3)
//	if limiter.Allow(port) {
//		// forward alert
//	}
//
// # Memory management
//
// Call [Limiter.Purge] periodically (e.g. from the daemon tick loop) to
// reclaim memory from buckets whose windows have expired and are no longer
// receiving events.
package ratelimit
