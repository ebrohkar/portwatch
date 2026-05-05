package scanner

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of a single port.
type PortState struct {
	Port     int
	Protocol string
	Open     bool
}

// ScanResult holds all open ports found during a scan.
type ScanResult struct {
	Timestamp time.Time
	Ports     []PortState
}

// Scanner scans local ports within a given range.
type Scanner struct {
	StartPort int
	EndPort   int
	Protocol  string
	Timeout   time.Duration
}

// NewScanner creates a Scanner with sensible defaults.
func NewScanner(start, end int) *Scanner {
	return &Scanner{
		StartPort: start,
		EndPort:   end,
		Protocol:  "tcp",
		Timeout:   500 * time.Millisecond,
	}
}

// Scan checks every port in the configured range and returns open ones.
func (s *Scanner) Scan() (*ScanResult, error) {
	if s.StartPort < 1 || s.EndPort > 65535 || s.StartPort > s.EndPort {
		return nil, fmt.Errorf("invalid port range: %d-%d", s.StartPort, s.EndPort)
	}

	result := &ScanResult{
		Timestamp: time.Now(),
	}

	for port := s.StartPort; port <= s.EndPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout(s.Protocol, address, s.Timeout)
		if err == nil {
			conn.Close()
			result.Ports = append(result.Ports, PortState{
				Port:     port,
				Protocol: s.Protocol,
				Open:     true,
			})
		}
	}

	return result, nil
}
