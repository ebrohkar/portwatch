package dispatcher_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/dispatcher"
)

// stubSender records every alert it receives and optionally returns an error.
type stubSender struct {
	alerts []alert.Alert
	errOnSend error
}

func (s *stubSender) Send(a alert.Alert) error {
	s.alerts = append(s.alerts, a)
	return s.errOnSend
}

func makeAlert(severity string) alert.Alert {
	return alert.Alert{
		Port:      8080,
		Event:     "open",
		Severity:  severity,
		Timestamp: time.Now(),
	}
}

func TestNew_EmptyChannels_ReturnsError(t *testing.T) {
	_, err := dispatcher.New(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil channels")
	}
}

func TestNew_UnknownSeverity_ReturnsError(t *testing.T) {
	ch := map[string]dispatcher.Sender{"a": &stubSender{}}
	_, err := dispatcher.New(ch, []dispatcher.Rule{{MinSeverity: "extreme", Channel: "a"}})
	if err == nil {
		t.Fatal("expected error for unknown severity")
	}
}

func TestNew_UnknownChannel_ReturnsError(t *testing.T) {
	ch := map[string]dispatcher.Sender{"a": &stubSender{}}
	_, err := dispatcher.New(ch, []dispatcher.Rule{{MinSeverity: "low", Channel: "missing"}})
	if err == nil {
		t.Fatal("expected error for unknown channel")
	}
}

func TestDispatch_RoutesToMatchingChannel(t *testing.T) {
	slack := &stubSender{}
	pager := &stubSender{}
	ch := map[string]dispatcher.Sender{"slack": slack, "pager": pager}
	rules := []dispatcher.Rule{
		{MinSeverity: "low", Channel: "slack"},
		{MinSeverity: "critical", Channel: "pager"},
	}
	d, err := dispatcher.New(ch, rules)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// medium alert → only slack
	if err := d.Dispatch(makeAlert("medium")); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(slack.alerts) != 1 {
		t.Errorf("slack: want 1 alert, got %d", len(slack.alerts))
	}
	if len(pager.alerts) != 0 {
		t.Errorf("pager: want 0 alerts, got %d", len(pager.alerts))
	}

	// critical alert → both channels
	if err := d.Dispatch(makeAlert("critical")); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if len(slack.alerts) != 2 {
		t.Errorf("slack: want 2 alerts, got %d", len(slack.alerts))
	}
	if len(pager.alerts) != 1 {
		t.Errorf("pager: want 1 alert, got %d", len(pager.alerts))
	}
}

func TestDispatch_DeduplicatesChannels(t *testing.T) {
	slack := &stubSender{}
	ch := map[string]dispatcher.Sender{"slack": slack}
	// Two rules pointing at the same channel.
	rules := []dispatcher.Rule{
		{MinSeverity: "low", Channel: "slack"},
		{MinSeverity: "medium", Channel: "slack"},
	}
	d, _ := dispatcher.New(ch, rules)
	_ = d.Dispatch(makeAlert("high"))
	if len(slack.alerts) != 1 {
		t.Errorf("want 1 delivery, got %d", len(slack.alerts))
	}
}

func TestDispatch_SenderError_Propagated(t *testing.T) {
	sendErr := errors.New("network timeout")
	ch := map[string]dispatcher.Sender{"a": &stubSender{errOnSend: sendErr}}
	d, _ := dispatcher.New(ch, []dispatcher.Rule{{MinSeverity: "low", Channel: "a"}})
	err := d.Dispatch(makeAlert("high"))
	if err == nil {
		t.Fatal("expected error from failing sender")
	}
	if !errors.Is(err, sendErr) {
		t.Errorf("want wrapped sendErr, got %v", err)
	}
}

func TestDispatch_UnknownAlertSeverity_ReturnsError(t *testing.T) {
	ch := map[string]dispatcher.Sender{"a": &stubSender{}}
	d, _ := dispatcher.New(ch, []dispatcher.Rule{{MinSeverity: "low", Channel: "a"}})
	a := makeAlert("unknown")
	if err := d.Dispatch(a); err == nil {
		t.Fatal("expected error for unknown alert severity")
	}
}
