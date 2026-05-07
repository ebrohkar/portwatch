package notifier

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/retry"
)

// errWriter fails on the first N writes then succeeds.
type errWriter struct {
	failFor int
	calls   int
	buf     bytes.Buffer
}

func (e *errWriter) Write(p []byte) (int, error) {
	e.calls++
	if e.calls <= e.failFor {
		return 0, errors.New("transient write error")
	}
	return e.buf.Write(p)
}

func TestSend_WithRetry_EventuallySucceeds(t *testing.T) {
	w := &errWriter{failFor: 2}
	n := New(w)

	policy := retry.Policy{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		Multiplier:   1.0,
		MaxDelay:     10 * time.Millisecond,
	}

	a := alert.New(8080, "open", "warn")
	var lastErr error
	err := policy.Do(context.Background(), func() error {
		lastErr = n.Send(a)
		return lastErr
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if !strings.Contains(w.buf.String(), "8080") {
		t.Errorf("expected port 8080 in output, got: %s", w.buf.String())
	}
	if w.calls != 3 {
		t.Errorf("expected 3 write calls (2 failures + 1 success), got %d", w.calls)
	}
}

func TestSend_WithRetry_ExhaustedReturnsError(t *testing.T) {
	w := &errWriter{failFor: 99}
	n := New(w)

	policy := retry.Policy{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		Multiplier:   1.0,
		MaxDelay:     5 * time.Millisecond,
	}

	a := alert.New(9090, "closed", "info")
	err := policy.Do(context.Background(), func() error {
		return n.Send(a)
	})
	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
}
