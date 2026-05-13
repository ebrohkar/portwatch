// Package watchdog tracks consecutive scan failures and triggers
// a configurable callback when the failure threshold is exceeded.
package watchdog

import (
	"fmt"
	"sync"
	"time"
)

// AlertFunc is called when the failure threshold is breached.
type AlertFunc func(failures int, lastErr error)

// Watchdog monitors scan health and fires an alert after a
// configurable number of consecutive failures.
type Watchdog struct {
	mu           sync.Mutex
	threshold    int
	consecutive  int
	lastErr      error
	lastFailure  time.Time
	alertFn      AlertFunc
	alerted      bool
}

// New creates a Watchdog with the given failure threshold and alert
// callback. threshold must be >= 1; alertFn must not be nil.
func New(threshold int, alertFn AlertFunc) (*Watchdog, error) {
	if threshold < 1 {
		return nil, fmt.Errorf("watchdog: threshold must be >= 1, got %d", threshold)
	}
	if alertFn == nil {
		return nil, fmt.Errorf("watchdog: alertFn must not be nil")
	}
	return &Watchdog{
		threshold: threshold,
		alertFn:   alertFn,
	}, nil
}

// RecordSuccess resets the consecutive failure counter and clears
// the alerted flag so future failure runs can trigger again.
func (w *Watchdog) RecordSuccess() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.consecutive = 0
	w.lastErr = nil
	w.alerted = false
}

// RecordFailure increments the consecutive failure counter. When the
// counter reaches the threshold for the first time the alert callback
// is invoked synchronously.
func (w *Watchdog) RecordFailure(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.consecutive++
	w.lastErr = err
	w.lastFailure = time.Now().UTC()
	if w.consecutive >= w.threshold && !w.alerted {
		w.alerted = true
		w.alertFn(w.consecutive, err)
	}
}

// Consecutive returns the current run of consecutive failures.
func (w *Watchdog) Consecutive() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.consecutive
}

// LastFailure returns the timestamp of the most recent failure, or
// the zero value if no failure has been recorded.
func (w *Watchdog) LastFailure() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastFailure
}
