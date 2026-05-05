package scanner

import (
	"net"
	"testing"
	"time"
)

// startTestServer opens a TCP listener on an OS-assigned port and returns it.
func startTestServer(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port
}

func TestNewScanner_Defaults(t *testing.T) {
	s := NewScanner(1, 1024)
	if s.StartPort != 1 || s.EndPort != 1024 {
		t.Errorf("unexpected port range: %d-%d", s.StartPort, s.EndPort)
	}
	if s.Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", s.Protocol)
	}
	if s.Timeout != 500*time.Millisecond {
		t.Errorf("unexpected timeout: %v", s.Timeout)
	}
}

func TestScan_DetectsOpenPort(t *testing.T) {
	ln, port := startTestServer(t)
	defer ln.Close()

	s := NewScanner(port, port)
	s.Timeout = 200 * time.Millisecond

	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(result.Ports) != 1 {
		t.Fatalf("expected 1 open port, got %d", len(result.Ports))
	}
	if result.Ports[0].Port != port {
		t.Errorf("expected port %d, got %d", port, result.Ports[0].Port)
	}
	if !result.Ports[0].Open {
		t.Error("expected port to be marked open")
	}
}

func TestScan_InvalidRange(t *testing.T) {
	s := NewScanner(9000, 8000) // reversed range
	_, err := s.Scan()
	if err == nil {
		t.Error("expected error for invalid port range, got nil")
	}
}

func TestScan_NoOpenPorts(t *testing.T) {
	// Use a very narrow range unlikely to have anything listening.
	s := NewScanner(1, 1)
	s.Timeout = 100 * time.Millisecond

	result, err := s.Scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// We simply verify the result is non-nil; port 1 is almost never open.
	if result == nil {
		t.Error("expected non-nil result")
	}
}
