package ratelimit_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/ratelimit"
)

func TestNew_DefaultsApplied(t *testing.T) {
	l := ratelimit.New(0, 0)
	if l == nil {
		t.Fatal("expected non-nil Limiter")
	}
	// With defaults (maxHits=1, window=1m) the first call must be allowed.
	if !l.Allow(80) {
		t.Error("first Allow should return true")
	}
}

func TestAllow_PermitsUpToMaxHits(t *testing.T) {
	l := ratelimit.New(time.Minute, 3)

	for i := 0; i < 3; i++ {
		if !l.Allow(443) {
			t.Errorf("call %d: expected Allow to return true", i+1)
		}
	}
	if l.Allow(443) {
		t.Error("4th call: expected Allow to return false")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	l := ratelimit.New(time.Minute, 1)

	if !l.Allow(80) {
		t.Error("port 80 first call should be allowed")
	}
	if !l.Allow(443) {
		t.Error("port 443 first call should be allowed independently")
	}
	if l.Allow(80) {
		t.Error("port 80 second call should be suppressed")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	l := ratelimit.New(50*time.Millisecond, 1)

	if !l.Allow(8080) {
		t.Fatal("first call should be allowed")
	}
	if l.Allow(8080) {
		t.Fatal("second call within window should be suppressed")
	}

	time.Sleep(60 * time.Millisecond)

	if !l.Allow(8080) {
		t.Error("call after window expiry should be allowed again")
	}
}

func TestReset_ClearsPortState(t *testing.T) {
	l := ratelimit.New(time.Minute, 1)

	l.Allow(9090)
	l.Reset(9090)

	if !l.Allow(9090) {
		t.Error("after Reset, Allow should return true again")
	}
}

func TestPurge_RemovesExpiredBuckets(t *testing.T) {
	l := ratelimit.New(30*time.Millisecond, 1)
	l.Allow(1234)

	time.Sleep(40 * time.Millisecond)
	l.Purge()

	// After purge the bucket is gone; next Allow should start a fresh window.
	if !l.Allow(1234) {
		t.Error("after Purge of expired bucket, Allow should return true")
	}
}
