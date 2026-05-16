// Package escalation implements automatic severity promotion for repeated
// port-change alerts.
//
// An Escalator tracks how many times a given port+event combination has fired
// within a sliding time window. Once the count reaches the configured
// threshold the alert severity is promoted one level:
//
//	info    → warning
//	warning → critical
//	critical stays critical
//
// State for a key is reset automatically when its window expires, or
// explicitly via Reset.
package escalation
