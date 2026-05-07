// Package retry provides a configurable retry Policy with exponential
// backoff for use across portwatch subsystems.
//
// Basic usage:
//
//	p := retry.DefaultPolicy()
//	err := p.Do(ctx, func() error {
//		return someUnreliableCall()
//	})
//	if errors.Is(err, retry.ErrExhausted) {
//		// all attempts consumed
//	}
package retry
