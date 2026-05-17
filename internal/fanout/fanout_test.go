package fanout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/fanout"
)

// recordSink captures every alert it receives.
type recordSink struct {
	got []alert.Alert
	err error
}

func (r *recordSink) Send(_ context.Context, a alert.Alert) error {
	r.got = append(r.got, a)
	return r.err
}

func makeAlert() alert.Alert {
	return alert.New(8080, "open", "high", time.Now())
}

func TestNew_PanicsOnNoSinks(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic with no sinks")
		}
	}()
	fanout.New()
}

func TestNew_StoresAllSinks(t *testing.T) {
	s1, s2 := &recordSink{}, &recordSink{}
	f := fanout.New(s1, s2)
	if f.Len() != 2 {
		t.Fatalf("expected 2 sinks, got %d", f.Len())
	}
}

func TestSend_DeliveriesToAllSinks(t *testing.T) {
	s1, s2 := &recordSink{}, &recordSink{}
	f := fanout.New(s1, s2)
	a := makeAlert()

	if err := f.Send(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s1.got) != 1 || len(s2.got) != 1 {
		t.Fatal("expected both sinks to receive the alert")
	}
}

func TestSend_ContinuesAfterSinkError(t *testing.T) {
	s1 := &recordSink{err: errors.New("boom")}
	s2 := &recordSink{}
	f := fanout.New(s1, s2)

	err := f.Send(context.Background(), makeAlert())
	if err == nil {
		t.Fatal("expected error from failing sink")
	}
	if len(s2.got) != 1 {
		t.Fatal("second sink should still receive the alert")
	}
}

func TestSend_ReturnsFirstError(t *testing.T) {
	e1 := errors.New("first")
	s1 := &recordSink{err: e1}
	s2 := &recordSink{err: errors.New("second")}
	f := fanout.New(s1, s2)

	err := f.Send(context.Background(), makeAlert())
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(err, e1) {
		t.Fatalf("expected first error to be wrapped, got: %v", err)
	}
}

func TestSend_SingleSink_NoError(t *testing.T) {
	s := &recordSink{}
	f := fanout.New(s)
	if err := f.Send(context.Background(), makeAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.got) != 1 {
		t.Fatal("sink should have received the alert")
	}
}
