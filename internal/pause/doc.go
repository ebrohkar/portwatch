// Package pause provides a Store that temporarily suppresses alert delivery
// for a (port, event) pair for a caller-specified duration.
//
// Pauses are automatically expired when the duration elapses; no background
// goroutine is required. Callers may also lift a pause early via Resume.
//
// Typical use: a human operator pauses noisy alerts during a maintenance
// window without restarting the daemon.
package pause
