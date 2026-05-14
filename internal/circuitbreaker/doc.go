// Package circuitbreaker provides a thread-safe circuit breaker for use in
// portwatch components that call external or unreliable subsystems (e.g. the
// notifier, audit writer, or rotation backend).
//
// The breaker moves through three states:
//
//   - Closed  – normal operation; failures are counted.
//   - Open    – calls are rejected with ErrOpen after the failure threshold is
//     reached; the breaker remains open for the configured reset timeout.
//   - HalfOpen – one probe call is allowed after the timeout elapses; a
//     successful RecordSuccess transitions back to Closed, while another
//     RecordFailure re-opens the circuit.
//
// Usage:
//
//	b := circuitbreaker.New(5, 30*time.Second)
//	if err := b.Allow(); err != nil {
//	    // circuit is open – skip the call
//	}
//	if err := doCall(); err != nil {
//	    b.RecordFailure()
//	} else {
//	    b.RecordSuccess()
//	}
package circuitbreaker
