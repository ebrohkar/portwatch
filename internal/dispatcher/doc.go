// Package dispatcher routes alert.Alert values to one or more named notifier
// channels according to severity-based rules.
//
// A Rule pairs a minimum severity level with a channel name. When Dispatch is
// called the alert's severity is compared against every rule; all channels
// whose threshold is met (or exceeded) receive the alert exactly once, even
// if multiple rules reference the same channel.
//
// Example:
//
//	ch := map[string]dispatcher.Sender{
//		"slack": slackNotifier,
//		"pager": pagerNotifier,
//	}
//	rules := []dispatcher.Rule{
//		{MinSeverity: "low",      Channel: "slack"},
//		{MinSeverity: "critical", Channel: "pager"},
//	}
//	d, err := dispatcher.New(ch, rules)
//	if err != nil { … }
//	_ = d.Dispatch(someAlert)
package dispatcher
