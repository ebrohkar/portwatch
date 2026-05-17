// Package prober implements active TCP reachability probing for portwatch.
//
// Unlike the passive scanner which enumerates open ports on the local host,
// the prober dials out to a target host and measures whether specific ports
// respond within a configurable timeout. Results include latency and are
// suitable for feeding into the alert pipeline.
package prober
