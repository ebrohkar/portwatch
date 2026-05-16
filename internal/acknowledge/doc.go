// Package acknowledge provides a thread-safe acknowledgement store for
// portwatch. Operators can acknowledge a (port, event) pair for a
// configurable TTL so that the alert pipeline suppresses further
// notifications for that combination until the acknowledgement expires
// or is explicitly revoked.
package acknowledge
