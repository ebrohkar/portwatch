// Package jitter provides a simple utility for adding randomised offsets to
// durations. Spreading out periodic scans across multiple portwatch instances
// prevents simultaneous bursts of port-scan activity that could saturate
// local resources or produce correlated spikes in alert output.
//
// Usage:
//
//	j := jitter.New(30*time.Second, 0.20) // ±20 % of 30 s
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-time.After(j.Duration()):
//			scan()
//		}
//	}
package jitter
