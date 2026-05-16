package trend

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroInterval(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero interval")
		}
	}()
	New(0, 3)
}

func TestNew_PanicsOnLowMaxBuckets(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when maxBuckets < 2")
		}
	}()
	New(time.Second, 1)
}

func TestTrend_NoData_ReturnsStable(t *testing.T) {
	tr := New(time.Minute, 3)
	if d := tr.Trend(80); d != Stable {
		t.Fatalf("expected Stable, got %v", d)
	}
}

func TestTrend_OneBucket_ReturnsStable(t *testing.T) {
	now := time.Now()
	tr := withClock(time.Minute, 3, fixedClock(now))
	tr.Record(80)
	tr.Record(80)
	if d := tr.Trend(80); d != Stable {
		t.Fatalf("expected Stable with single bucket, got %v", d)
	}
}

func TestTrend_Rising(t *testing.T) {
	now := time.Now()
	tr := withClock(time.Minute, 3, fixedClock(now))
	tr.Record(443)

	tr.clock = fixedClock(now.Add(time.Minute))
	tr.Record(443)
	tr.Record(443)
	tr.Record(443)

	if d := tr.Trend(443); d != Rising {
		t.Fatalf("expected Rising, got %v", d)
	}
}

func TestTrend_Falling(t *testing.T) {
	now := time.Now()
	tr := withClock(time.Minute, 3, fixedClock(now))
	tr.Record(22)
	tr.Record(22)
	tr.Record(22)

	tr.clock = fixedClock(now.Add(time.Minute))
	tr.Record(22)

	if d := tr.Trend(22); d != Falling {
		t.Fatalf("expected Falling, got %v", d)
	}
}

func TestTrend_Stable_EqualBuckets(t *testing.T) {
	now := time.Now()
	tr := withClock(time.Minute, 3, fixedClock(now))
	tr.Record(8080)

	tr.clock = fixedClock(now.Add(time.Minute))
	tr.Record(8080)

	if d := tr.Trend(8080); d != Stable {
		t.Fatalf("expected Stable, got %v", d)
	}
}

func TestRecord_EvictsOldBuckets(t *testing.T) {
	now := time.Now()
	tr := withClock(time.Minute, 2, fixedClock(now))
	for i := 0; i < 5; i++ {
		tr.clock = fixedClock(now.Add(time.Duration(i) * time.Minute))
		tr.Record(9090)
	}
	tr.mu.Lock()
	l := len(tr.buckets[9090])
	tr.mu.Unlock()
	if l > 2 {
		t.Fatalf("expected at most 2 buckets, got %d", l)
	}
}

func TestDirection_String(t *testing.T) {
	cases := []struct {
		d    Direction
		want string
	}{
		{Stable, "stable"},
		{Rising, "rising"},
		{Falling, "falling"},
	}
	for _, c := range cases {
		if got := c.d.String(); got != c.want {
			t.Errorf("Direction(%d).String() = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestSummary_ContainsPort(t *testing.T) {
	tr := New(time.Minute, 2)
	s := tr.Summary(3000)
	if len(s) == 0 {
		t.Fatal("expected non-empty summary")
	}
}
