package digest

import (
	"fmt"

	"github.com/example/portwatch/internal/alert"
)

// Stage is a pipeline stage that feeds every passing alert into a Digest.
// It never drops alerts — it only observes them.
type Stage struct {
	d *Digest
}

// NewStage returns a Stage backed by d. Panics if d is nil.
func NewStage(d *Digest) *Stage {
	if d == nil {
		panic("digest: NewStage requires a non-nil Digest")
	}
	return &Stage{d: d}
}

// Allow records the alert in the digest and always returns true.
func (s *Stage) Allow(a alert.Alert) bool {
	s.d.Add(a)
	return true
}

// String implements fmt.Stringer.
func (s *Stage) String() string {
	return fmt.Sprintf("DigestStage(entries=%d)", len(s.d.Entries()))
}
