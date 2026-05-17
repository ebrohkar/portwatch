// Package prober provides active probing of specific ports to verify
// reachability beyond a simple open/closed scan. It attempts a TCP dial
// and records latency alongside the result.
package prober

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Result holds the outcome of a single probe attempt.
type Result struct {
	Port      int
	Reachable bool
	Latency   time.Duration
	Err       error
	ProbeTime time.Time
}

// Prober actively dials ports and returns reachability results.
type Prober struct {
	host    string
	timeout time.Duration
	dial    func(ctx context.Context, network, addr string) (net.Conn, error)
}

// New creates a Prober targeting host with the given dial timeout.
// A zero timeout defaults to 2 seconds.
func New(host string, timeout time.Duration) (*Prober, error) {
	if host == "" {
		return nil, fmt.Errorf("prober: host must not be empty")
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	d := &net.Dialer{Timeout: timeout}
	return &Prober{
		host:    host,
		timeout: timeout,
		dial:    d.DialContext,
	}, nil
}

// Probe attempts a TCP connection to the given port and returns a Result.
func (p *Prober) Probe(ctx context.Context, port int) Result {
	if port < 1 || port > 65535 {
		return Result{
			Port:      port,
			Reachable: false,
			Err:       fmt.Errorf("prober: invalid port %d", port),
			ProbeTime: time.Now().UTC(),
		}
	}
	addr := fmt.Sprintf("%s:%d", p.host, port)
	start := time.Now()
	conn, err := p.dial(ctx, "tcp", addr)
	latency := time.Since(start)
	if err != nil {
		return Result{Port: port, Reachable: false, Latency: latency, Err: err, ProbeTime: start.UTC()}
	}
	_ = conn.Close()
	return Result{Port: port, Reachable: true, Latency: latency, ProbeTime: start.UTC()}
}

// ProbeAll probes each port in the slice and returns all results.
func (p *Prober) ProbeAll(ctx context.Context, ports []int) []Result {
	results := make([]Result, 0, len(ports))
	for _, port := range ports {
		results = append(results, p.Probe(ctx, port))
	}
	return results
}
