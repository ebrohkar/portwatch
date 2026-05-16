package rollup

import (
	"strings"
	"testing"
	"time"

	"portwatch/internal/alert"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) clock { return func() time.Time { return t } }

func makeAlert(port int, event string) alert.Alert {
	a, _ := alert.New(port, event, "info")
	return a
}

func TestNew_PanicsOnLowThreshold(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for threshold < 2")
		}
	}()
	New(1, time.Second)
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	New(3, 0)
}

func TestAdd_BelowThreshold_ReturnsFalse(t *testing.T) {
	r := withClock(3, time.Minute, fixedClock(epoch))
	a := makeAlert(8080, "open")

	for i := 0; i < 2; i++ {
		_, ok := r.Add(a)
		if ok {
			t.Fatalf("iteration %d: expected false before threshold", i)
		}
	}
}

func TestAdd_AtThreshold_ReturnsSummary(t *testing.T) {
	r := withClock(3, time.Minute, fixedClock(epoch))
	a := makeAlert(9090, "open")

	var got alert.Alert
	var fired bool
	for i := 0; i < 3; i++ {
		got, fired = r.Add(a)
	}

	if !fired {
		t.Fatal("expected summary alert at threshold")
	}
	if !strings.Contains(got.Message, "[rollup]") {
		t.Errorf("summary message missing [rollup] prefix: %q", got.Message)
	}
	if got.Port != 9090 {
		t.Errorf("expected port 9090, got %d", got.Port)
	}
}

func TestAdd_WindowExpiry_ResetsCount(t *testing.T) {
	now := epoch
	r := withClock(3, 10*time.Second, func() time.Time { return now })
	a := makeAlert(443, "close")

	r.Add(a)
	r.Add(a)

	// advance past the window
	now = epoch.Add(11 * time.Second)

	_, ok := r.Add(a) // should restart the bucket, not fire
	if ok {
		t.Fatal("expected no summary after window expiry reset")
	}
}

func TestAdd_DifferentPortsAreIndependent(t *testing.T) {
	r := withClock(2, time.Minute, fixedClock(epoch))
	a1 := makeAlert(80, "open")
	a2 := makeAlert(443, "open")

	r.Add(a1)
	r.Add(a2)

	_, ok1 := r.Add(a1)
	_, ok2 := r.Add(a2)

	if !ok1 || !ok2 {
		t.Errorf("expected both ports to reach threshold independently: ok1=%v ok2=%v", ok1, ok2)
	}
}

func TestReset_ClearsPortState(t *testing.T) {
	r := withClock(2, time.Minute, fixedClock(epoch))
	a := makeAlert(8080, "open")

	r.Add(a)
	r.Reset(8080, "open")

	_, ok := r.Add(a) // bucket was cleared; count is 1 again
	if ok {
		t.Fatal("expected no summary after reset")
	}
}
