package grouper

import (
	"context"
	"fmt"

	"github.com/example/portwatch/internal/alert"
)

// Stage is a pipeline stage that accumulates alerts into the Grouper.
// It always permits the alert to continue downstream; grouping is a
// side-effect so that other stages still receive every alert.
type Stage struct {
	g *Grouper
}

// NewStage returns a Stage backed by g.
// It panics if g is nil.
func NewStage(g *Grouper) *Stage {
	if g == nil {
		panic("grouper: NewStage called with nil Grouper")
	}
	return &Stage{g: g}
}

// Allow records the alert in the grouper and always returns true.
func (s *Stage) Allow(_ context.Context, a alert.Alert) (bool, error) {
	s.g.Add(a)
	return true, nil
}

// String implements fmt.Stringer for pipeline introspection.
func (s *Stage) String() string {
	return fmt.Sprintf("grouper.Stage(groups=%d)", len(s.g.rules))
}
