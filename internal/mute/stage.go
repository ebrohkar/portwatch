package mute

import (
	"context"
	"fmt"

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

// Allow returns false (dropping the alert) when the port is muted.
func (s *Stage) Allow(_ context.Context, a alert.Alert) (alert.Alert, bool, error) {
	if s.store.IsMuted(a.Port) {
		return alert.Alert{}, false, fmt.Errorf("mute: port %d is muted", a.Port)
	}
	return a, true, nil
}
