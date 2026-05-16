package mute

import (
	"context"

	"github.com/example/portwatch/internal/alert"
)

// Stage is a pipeline stage that drops alerts for muted ports.
type Stage struct {
	store *Store
}

// NewStage returns a Stage backed by the given Store.
func NewStage(store *Store) *Stage {
	if store == nil {
		panic("mute: store must not be nil")
	}
	return &Stage{store: store}
}

// Allow returns the alert unchanged when the port is not muted.
// When the port is muted it returns (zero Alert, false, nil), signalling to
// the pipeline that the alert should be silently dropped without treating it
// as an error condition.
func (s *Stage) Allow(_ context.Context, a alert.Alert) (alert.Alert, bool, error) {
	if s.store.IsMuted(a.Port) {
		return alert.Alert{}, false, nil
	}
	return a, true, nil
}
