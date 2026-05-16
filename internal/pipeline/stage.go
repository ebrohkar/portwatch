package pipeline

import (
	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/metrics"
)

// MetricsStage returns a Stage that increments the suppressed counter on
// the supplied Metrics whenever an alert is dropped by a subsequent
// stage.  Because stages are evaluated in order, wrap MetricsStage
// around a dropping stage by placing it before that stage and using
// the returned pair via WithMetrics.
//
// For simple accounting, InstrumentedStage wraps an existing Stage and
// records every dropped alert against m.
func InstrumentedStage(inner Stage, m *metrics.Metrics) Stage {
	if m == nil {
		return inner
	}
	return func(a alert.Alert) bool {
		if ok := inner(a); !ok {
			m.IncSuppressed()
			return false
		}
		return true
	}
}

// AlertingStage returns a Stage that always passes but calls onAlert
// with each alert that reaches it.  Useful for side-effects such as
// writing to history or the audit log without blocking the pipeline.
func AlertingStage(onAlert func(alert.Alert)) Stage {
	if onAlert == nil {
		return func(alert.Alert) bool { return true }
	}
	return func(a alert.Alert) bool {
		onAlert(a)
		return true
	}
}
