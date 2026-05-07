// Package ratelimit implements a concurrent-safe, per-port rate limiter used
// by portwatch to suppress alert floods when a port repeatedly transitions
// between open and closed states.
//
// Usage:
//
//	limiter := ratelimit.New(time.Minute, 3)
//	if limiter.Allow(port) {
//		// forward alert
//	}
//
// Call Purge periodically (e.g. from the daemon tick loop) to reclaim memory
// from buckets whose windows have expired.
package ratelimit
