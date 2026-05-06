// Package daemon provides the core orchestration loop for portwatch.
//
// A Daemon periodically invokes the port scanner, computes the diff against
// the previously persisted state, evaluates configured rules against each
// changed port, and dispatches alert notifications for any rule violations.
//
// Typical usage:
//
//	d := daemon.New(daemon.Config{
//		Interval: 60 * time.Second,
//		RuleSet:  rs,
//		Scanner:  sc,
//		Store:    store,
//		Notifier: nt,
//	})
//	d.Run(ctx)
package daemon
