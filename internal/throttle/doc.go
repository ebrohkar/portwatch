// Package throttle implements a sliding-window token-bucket throttle for
// portwatch alert dispatch.
//
// Unlike ratelimit (which enforces per-port limits), throttle enforces a
// global cap across all ports and event types. This prevents alert storms
// when many ports change state simultaneously.
//
// Usage:
//
//	th := throttle.New(100, time.Minute)
//	if th.Allow() {
//		// dispatch alert
//	}
package throttle
