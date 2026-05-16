// Package window provides a sliding time-window counter keyed by port number.
//
// It is used by portwatch components (e.g. ratelimit, throttle) that need to
// track how many times an event has occurred within a rolling time period.
//
// Usage:
//
//	w := window.New(time.Minute)
//	count := w.Add(8080)   // record an event on port 8080
//	fmt.Println(count)     // events in the last minute for port 8080
package window
