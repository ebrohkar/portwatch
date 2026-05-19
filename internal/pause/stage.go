package pause

import (
	"fmt"

	"github.com/user/portwatch/internal/alert"
)

// Stage is a pipeline stage that drops alerts whose port+event is paused.
type Stage struct {
	store *Store
}

// NewStage creates a Stage backed by the given Store.
// Panics if store is nil.
func NewStage(store *Store) *Stage {
	if store == nil {
		panic("pause: NewStage requires a non-nil Store")
	}
	return &Stage{store: store}
}

// Allow returns false (dropping the alert) when the port+event is paused,
// and true otherwise.
func (s *Stage) Allow(a alert.Alert) bool {
	return !s.store.IsPaused(a.Port, a.Event)
}

// String implements fmt.Stringer for pipeline introspection.
func (s *Stage) String() string {
	return fmt.Sprintf("PauseStage(active=%d)", s.store.Len())
}
