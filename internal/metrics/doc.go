// Package metrics provides lightweight atomic counters that track portwatch
// daemon activity across its lifetime.
//
// Usage:
//
//	ctr := metrics.New()
//	ctr.IncScans()
//	ctr.IncAlerts()
//	ctr.WriteTo(os.Stdout)
//
// All counter methods are safe for concurrent use from multiple goroutines.
package metrics
