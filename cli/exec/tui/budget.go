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

// GlobalLogCapBytes is the TUI's shared memory budget for per-step
// log rings. When the combined retained size of all registered rings
// exceeds this cap, the oldest line from the single largest ring is
// dropped until the total fits. This policy preserves cheap-to-keep
// history from quiet steps while trimming the one that is actually
// spamming.
//
// The value is a deliberate compromise: large enough for reasonable
// CI output (~200 MiB typically fits the logs of dozens of steps),
// small enough not to invite accidental OOM kills in constrained
// environments. A flag to tune it can be added later if needed.
const GlobalLogCapBytes = 200 * 1024 * 1024

// DebugLogCapBytes is the separate cap for the zerolog debug tab.
// It is small because zerolog output is diagnostic noise, not the
// user's primary signal. Counted separately from the step budget so
// debug spam cannot crowd out step logs.
const DebugLogCapBytes = 5 * 1024 * 1024

// Budget tracks a set of rings against a shared byte cap. Call
// Register for each ring when it is created, then Enforce after each
// batch of appends. Enforce evicts lines from the largest ring first
// until the total fits, which preserves useful history from quiet
// steps while trimming the step that is actually growing.
type Budget struct {
	mu       sync.Mutex
	rings    []*Ring
	capBytes int
}

// NewBudget returns a Budget with the given byte cap. Zero means no
// enforcement — the Budget is inert and Enforce is a no-op.
func NewBudget(capBytes int) *Budget {
	return &Budget{capBytes: capBytes}
}

// Register adds a ring to the budget. The ring may still enforce its
// own per-ring cap independently; the budget enforces the shared cap
// across all registered rings.
func (b *Budget) Register(r *Ring) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rings = append(b.rings, r)
}

// Enforce evicts oldest-from-largest-ring until the total byte count
// across all registered rings is at or below the cap. Safe to call
// from any goroutine.
//
// The caller typically invokes Enforce on a timer (debounced) or
// after a batch of appends, rather than after every single line — the
// map and loop overhead per call is more than an eviction otherwise
// saves.
func (b *Budget) Enforce() {
	if b.capBytes <= 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	total := 0
	for _, r := range b.rings {
		total += r.Bytes()
	}

	for total > b.capBytes {
		var biggest *Ring
		var biggestBytes int
		for _, r := range b.rings {
			if size := r.Bytes(); size > biggestBytes {
				biggestBytes = size
				biggest = r
			}
		}
		if biggest == nil {
			// No ring has any content; nothing we can do.
			return
		}
		freed, ok := biggest.evictOldest()
		if !ok {
			return
		}
		total -= freed
	}
}
