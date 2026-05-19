// Package decay implements exponential half-life score decay for port-based
// alert scoring. Each port maintains an independent score that is reduced
// by half every configured half-life duration when no new events are added.
//
// Typical usage:
//
//	d := decay.New(5 * time.Minute)
//	score := d.Add(8080, 1.0)  // increment score for port 8080
//	current := d.Get(8080)     // read decayed score at any time
//	d.Reset(8080)              // clear score entirely
package decay
