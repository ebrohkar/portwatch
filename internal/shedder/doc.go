// Package shedder provides a concurrency-based load shedder for the
// portwatch alert pipeline.
//
// When the number of concurrently processed alerts reaches the configured
// ceiling, additional calls to Allow return ErrShedded so that upstream
// stages can discard the alert rather than block indefinitely.
//
// Typical usage inside a pipeline stage:
//
//	if err := shed.Allow(); err != nil {
//		// drop the alert
//		return false
//	}
//	defer shed.Done()
//	// … process alert …
package shedder
