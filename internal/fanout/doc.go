// Package fanout provides a broadcast primitive that delivers each alert to
// every registered sink in sequence.
//
// All sinks are attempted regardless of individual failures; the first error
// encountered is returned to the caller after all sinks have been tried.
//
// Typical usage:
//
//	f := fanout.New(emailSink, slackSink, logSink)
//	if err := f.Send(ctx, a); err != nil {
//		log.Printf("fanout partial failure: %v", err)
//	}
package fanout
