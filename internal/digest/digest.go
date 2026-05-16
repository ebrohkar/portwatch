// Package digest produces periodic scan digest reports summarising port
// activity over a configurable rolling window.
package digest

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/example/portwatch/internal/alert"
)

// Entry holds aggregated counts for a single port within the window.
type Entry struct {
	Port   int
	Opened int
	Closed int
	Total  int
}

// Digest accumulates alert events and writes summary reports.
type Digest struct {
	mu       sync.Mutex
	counts   map[int]*Entry
	window   time.Duration
	lastFlush time.Time
	w        io.Writer
	clock    func() time.Time
}

// New returns a Digest that flushes summaries to w every window duration.
// Panics if window is zero or w is nil.
func New(w io.Writer, window time.Duration) *Digest {
	if window == 0 {
		panic("digest: window must be non-zero")
	}
	if w == nil {
		w = os.Stdout
	}
	return &Digest{
		counts:    make(map[int]*Entry),
		window:    window,
		w:         w,
		clock:     time.Now,
		lastFlush: time.Now(),
	}
}

// withClock replaces the clock — for testing only.
func withClock(w io.Writer, window time.Duration, clk func() time.Time) *Digest {
	d := New(w, window)
	d.clock = clk
	d.lastFlush = clk()
	return d
}

// Add records an alert into the rolling window.
func (d *Digest) Add(a alert.Alert) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e := d.entry(a.Port)
	switch a.Event {
	case "opened":
		e.Opened++
	case "closed":
		e.Closed++
	}
	e.Total++
}

// Flush writes the accumulated digest to the writer and resets counters.
// It is a no-op if the window has not elapsed since the last flush.
func (d *Digest) Flush() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.clock().Sub(d.lastFlush) < d.window {
		return false
	}
	d.write()
	d.counts = make(map[int]*Entry)
	d.lastFlush = d.clock()
	return true
}

// Entries returns a sorted snapshot of current counts without flushing.
func (d *Digest) Entries() []Entry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]Entry, 0, len(d.counts))
	for _, e := range d.counts {
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func (d *Digest) entry(port int) *Entry {
	if d.counts[port] == nil {
		d.counts[port] = &Entry{Port: port}
	}
	return d.counts[port]
}

func (d *Digest) write() {
	entries := make([]Entry, 0, len(d.counts))
	for _, e := range d.counts {
		entries = append(entries, *e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Port < entries[j].Port })
	fmt.Fprintf(d.w, "=== digest @ %s ===\n", d.clock().UTC().Format(time.RFC3339))
	if len(entries) == 0 {
		fmt.Fprintln(d.w, "  (no activity)")
		return
	}
	for _, e := range entries {
		fmt.Fprintf(d.w, "  port %d: total=%d opened=%d closed=%d\n",
			e.Port, e.Total, e.Opened, e.Closed)
	}
}
