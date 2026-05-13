// Package throttle implements a sliding-window token-bucket throttle for
// portwatch alert dispatch.
//
// Unlike ratelimit (which enforces per-port limits), throttle enforces a
// global cap across all ports and event types. This prevents alert storms
// when many ports change state simultaneously.
//
// The sliding window approach means that the rate limit is evaluated over
// a rolling time window rather than a fixed interval, providing smoother
// rate limiting behaviour without the burst spikes that fixed windows can
// allow at interval boundaries.
//
// Usage:
//
//	th := throttle.New(100, time.Minute)
//	if th.Allow() {
//		// dispatch alert
//	}
//
// To inspect current state without consuming a token, use Remaining:
//
//	remaining := th.Remaining()
package throttle
