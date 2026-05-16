// Package digest provides a rolling-window digest of port alert activity.
//
// A Digest accumulates opened/closed event counts per port and periodically
// writes a human-readable summary to any io.Writer.  Use NewStage to insert
// a non-dropping observation stage into an alert pipeline.
//
// Example:
//
//	dig := digest.New(os.Stdout, 5*time.Minute)
//	stage := digest.NewStage(dig)
//	// ... attach stage to pipeline ...
//	// periodically call dig.Flush() from a ticker goroutine
package digest
