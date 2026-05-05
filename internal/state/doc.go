// Package state provides persistent snapshot storage for port scan results
// and utilities for computing differences between consecutive scans.
//
// A Store serialises snapshots as JSON to a configurable file path, allowing
// portwatch to detect changes across daemon restarts. The Diff helper returns
// the ports that appeared or disappeared between two successive scans so that
// the rules engine and notifier can act on meaningful changes only.
package state
