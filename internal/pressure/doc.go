// Package pressure provides a sliding-window load gauge for alert pipelines.
//
// A Gauge accumulates event counts over a configurable time window and
// returns a normalised pressure level in [0.0, 1.0]. Downstream components
// can use this value to shed load, throttle notifications, or trigger
// escalation when the system is under stress.
//
// Example:
//
//	g := pressure.New(time.Minute, 100)
//	g.Record(1)
//	fmt.Println(g.Level()) // e.g. 0.01
package pressure
