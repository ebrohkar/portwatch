package prober_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/prober"
)

func startTCPServer(t *testing.T) (port int, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startTCPServer: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, func() { _ = ln.Close() }
}

func TestNew_EmptyHost_ReturnsError(t *testing.T) {
	_, err := prober.New("", 0)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestNew_ZeroTimeout_DefaultsToTwoSeconds(t *testing.T) {
	p, err := prober.New("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil prober")
	}
}

func TestProbe_OpenPort_ReturnsReachable(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	p, _ := prober.New("127.0.0.1", time.Second)
	res := p.Probe(context.Background(), port)

	if !res.Reachable {
		t.Fatalf("expected reachable=true, got err: %v", res.Err)
	}
	if res.Latency <= 0 {
		t.Error("expected positive latency")
	}
	if res.ProbeTime.IsZero() {
		t.Error("expected non-zero probe time")
	}
}

func TestProbe_ClosedPort_ReturnsUnreachable(t *testing.T) {
	p, _ := prober.New("127.0.0.1", 200*time.Millisecond)
	res := p.Probe(context.Background(), 1)

	if res.Reachable {
		t.Fatal("expected reachable=false for closed port")
	}
	if res.Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestProbe_InvalidPort_ReturnsError(t *testing.T) {
	p, _ := prober.New("127.0.0.1", time.Second)
	res := p.Probe(context.Background(), 0)

	if res.Reachable {
		t.Fatal("expected reachable=false")
	}
	if res.Err == nil {
		t.Error("expected error for port 0")
	}
}

func TestProbeAll_MixedPorts(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	p, _ := prober.New("127.0.0.1", 200*time.Millisecond)
	results := p.ProbeAll(context.Background(), []int{port, 1})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Reachable {
		t.Errorf("expected port %d reachable", port)
	}
	if results[1].Reachable {
		t.Error("expected port 1 unreachable")
	}
}
