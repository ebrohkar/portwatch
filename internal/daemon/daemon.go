package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/rules"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

// Config holds runtime configuration for the daemon.
type Config struct {
	Interval  time.Duration
	RuleSet   *rules.RuleSet
	Scanner   *scanner.Scanner
	Store     *state.Store
	Notifier  *notifier.Notifier
}

// Daemon orchestrates periodic port scanning, state diffing, and alerting.
type Daemon struct {
	cfg Config
}

// New creates a new Daemon with the provided configuration.
func New(cfg Config) *Daemon {
	if cfg.Interval <= 0 {
		cfg.Interval = 60 * time.Second
	}
	return &Daemon{cfg: cfg}
}

// Run starts the daemon loop, blocking until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	log.Printf("portwatch daemon starting (interval=%s)", d.cfg.Interval)

	if err := d.cfg.Store.Load(); err != nil {
		log.Printf("warn: could not load previous state: %v", err)
	}

	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	// Run immediately on start, then on each tick.
	if err := d.tick(); err != nil {
		log.Printf("scan error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch daemon stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := d.tick(); err != nil {
				log.Printf("scan error: %v", err)
			}
		}
	}
}

func (d *Daemon) tick() error {
	ports, err := d.cfg.Scanner.Scan()
	if err != nil {
		return err
	}

	diffs := d.cfg.Store.Diff(ports)
	for _, diff := range diffs {
		events := d.cfg.RuleSet.Evaluate(diff)
		for _, ev := range events {
			if err := d.cfg.Notifier.Send(ev); err != nil {
				log.Printf("notify error: %v", err)
			}
		}
	}

	d.cfg.Store.SetCurrent(ports)
	return d.cfg.Store.Save()
}
