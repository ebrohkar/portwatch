// Package alert provides the Alert type used to represent port change
// detection events within portwatch.
//
// Alerts carry a severity level (info, warning, critical), the affected port
// and protocol, the type of event (opened or closed), a human-readable
// message, and a UTC timestamp set at construction time.
//
// Use alert.New to construct a valid Alert and IsValid to verify that all
// required fields are present before forwarding to a notifier.
package alert
