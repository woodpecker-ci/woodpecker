// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tui

import (
	"sync"
)

// Ring is a FIFO line buffer with a byte-size limit and a truncation
// counter, intended to back a single step's log pane or the TUI's
// debug pane.
//
// When an Append would push the total byte count above the configured
// cap, the oldest lines are dropped until the new line fits. Every
// dropped line bumps the Truncated counter so the TUI can render a
// "[N lines truncated]" marker at the top of the pane.
//
// Ring is safe for concurrent use by one writer and one reader; this
// matches the TUI's producer-consumer model where tea.Msg handlers
// read while a pipeline logger goroutine writes.
type Ring struct {
	mu        sync.Mutex
	lines     []string
	bytes     int
	capBytes  int
	truncated uint64
}

// NewRing returns a Ring capped at capBytes. A cap of zero means
// unbounded — use with care; typically reserved for tests.
func NewRing(capBytes int) *Ring {
	return &Ring{capBytes: capBytes}
}

// Append adds a line to the ring. The line is stored as-is; trailing
// newlines are preserved because renderers may want to emit raw
// bytes. If the ring has a byte cap, the oldest lines are dropped
// until the incoming line fits.
//
// If the incoming line alone exceeds the cap, the line is stored
// and everything else is evicted. This is deliberate: a log line
// bigger than the buffer is a weird edge case, but dropping it
// silently would hide whatever produced it.
func (r *Ring) Append(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lineLen := len(line)
	if r.capBytes > 0 {
		for r.bytes+lineLen > r.capBytes && len(r.lines) > 0 {
			dropped := r.lines[0]
			r.lines = r.lines[1:]
			r.bytes -= len(dropped)
			r.truncated++
		}
	}
	r.lines = append(r.lines, line)
	r.bytes += lineLen
}

// Snapshot returns a copy of the currently retained lines, plus the
// number of lines that have been dropped since the ring was created.
// Callers may safely mutate the returned slice; it does not share
// backing storage with the ring.
func (r *Ring) Snapshot() (lines []string, truncated uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]string, len(r.lines))
	copy(out, r.lines)
	return out, r.truncated
}

// Bytes returns the current total byte count retained by the ring.
// Used by the budget controller when enforcing a global cap across
// multiple rings.
func (r *Ring) Bytes() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bytes
}

// Len returns the number of lines currently retained.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.lines)
}

// evictOldest drops the oldest line unconditionally, bumping the
// truncated counter. Returns the number of bytes freed and false if
// there was nothing to evict.
//
// Exposed for the global budget controller in budget.go.
func (r *Ring) evictOldest() (freed int, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.lines) == 0 {
		return 0, false
	}
	dropped := r.lines[0]
	r.lines = r.lines[1:]
	r.bytes -= len(dropped)
	r.truncated++
	return len(dropped), true
}
