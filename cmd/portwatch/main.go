package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/portwatch/internal/daemon"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/rules"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

func main() {
	var (
		rulesFile  = flag.String("rules", "rules.yaml", "path to rules YAML file")
		stateFile  = flag.String("state", "/var/lib/portwatch/state.json", "path to state file")
		startPort  = flag.Int("start", 1, "start of port range")
		endPort    = flag.Int("end", 65535, "end of port range")
		interval   = flag.Duration("interval", 60*time.Second, "scan interval")
		timeout    = flag.Duration("timeout", 500*time.Millisecond, "per-port dial timeout")
	)
	flag.Parse()

	rs, err := rules.LoadFromFile(*rulesFile)
	if err != nil {
		log.Fatalf("failed to load rules: %v", err)
	}

	sc, err := scanner.NewScanner(scanner.Options{
		StartPort: *startPort,
		EndPort:   *endPort,
		Timeout:   *timeout,
	})
	if err != nil {
		log.Fatalf("failed to create scanner: %v", err)
	}

	store := state.NewStore(*stateFile)
	nt := notifier.New(os.Stdout)

	d := daemon.New(daemon.Config{
		Interval: *interval,
		RuleSet:  rs,
		Scanner:  sc,
		Store:    store,
		Notifier: nt,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := d.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("daemon exited with error: %v", err)
	}
}
