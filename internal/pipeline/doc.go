// Package pipeline provides a composable, stage-based processing chain
// for port-change alerts.
//
// Alerts flow through an ordered sequence of Stage functions.  Each
// stage may pass or drop the alert.  Only alerts that survive every
// stage reach the notifier.
//
// The standard chain (built by FromParts) applies the following stages
// in order:
//
//  1. filter    – allow/exclude specific ports
//  2. debounce  – suppress rapid repeated events for the same port
//  3. ratelimit – cap the number of alerts per port per time window
//  4. suppress  – honour explicit suppression rules
//
// Custom stages can be injected via New for testing or extension.
package pipeline
