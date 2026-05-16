package quota

import (
	"fmt"

	"github.com/user/portwatch/internal/alert"
)

// Stage is a pipeline stage that drops alerts once a port's quota is exhausted.
type Stage struct {
	q *Quota
}

// NewStage wraps a Quota as a pipeline stage.
// Panics if q is nil.
func NewStage(q *Quota) *Stage {
	if q == nil {
		panic("quota: NewStage requires a non-nil Quota")
	}
	return &Stage{q: q}
}

// Allow returns true if the alert's port is still within quota, false otherwise.
func (s *Stage) Allow(a alert.Alert) bool {
	return s.q.Allow(a.Port)
}

// String returns a description of the stage including quota configuration.
func (s *Stage) String() string {
	return fmt.Sprintf("QuotaStage(%s)", s.q)
}
