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

package tui_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/tui"
)

func TestRingAppendWithinCap(t *testing.T) {
	r := tui.NewRing(100)
	r.Append("hello\n")
	r.Append("world\n")
	lines, truncated := r.Snapshot()
	assert.Equal(t, []string{"hello\n", "world\n"}, lines)
	assert.Equal(t, uint64(0), truncated)
	assert.Equal(t, 12, r.Bytes())
	assert.Equal(t, 2, r.Len())
}

func TestRingEvictsOldestWhenOverCap(t *testing.T) {
	// Cap exactly fits two 6-byte lines. A third forces the first out.
	r := tui.NewRing(12)
	r.Append("aaaaa\n") // 6 bytes
	r.Append("bbbbb\n") // 12 bytes, at cap
	r.Append("ccccc\n") // would be 18; evict first
	lines, truncated := r.Snapshot()
	assert.Equal(t, []string{"bbbbb\n", "ccccc\n"}, lines)
	assert.Equal(t, uint64(1), truncated)
}

func TestRingEvictsManyIfIncomingIsLarge(t *testing.T) {
	r := tui.NewRing(20)
	r.Append("a\n") // 2
	r.Append("b\n") // 4
	r.Append("c\n") // 6
	r.Append("d\n") // 8
	// Now append a 15-byte line; fits under the cap only after
	// evicting some (2+4+... until total <= 5 remaining slot).
	big := "0123456789abcd\n" // 15 bytes
	r.Append(big)
	lines, truncated := r.Snapshot()
	// The scheduler must have dropped enough to fit the new line.
	assert.LessOrEqual(t, r.Bytes(), 20)
	assert.Contains(t, lines, big, "the newest line must be retained")
	assert.Positive(t, truncated)
}

func TestRingOversizedLineEvictsEverythingAndStores(t *testing.T) {
	// If the incoming line alone is bigger than the cap, the
	// documented behavior is: evict all, then store the line. This
	// avoids silently dropping a line whose very size is the signal
	// the user wants to see.
	r := tui.NewRing(10)
	r.Append("old\n")
	big := "way-too-big-to-fit-in-cap\n" // 26 bytes
	r.Append(big)
	lines, truncated := r.Snapshot()
	assert.Equal(t, []string{big}, lines)
	assert.Equal(t, uint64(1), truncated)
}

func TestRingUnboundedCap(t *testing.T) {
	// Cap of 0 means no enforcement. Append a lot and verify nothing
	// is dropped.
	r := tui.NewRing(0)
	for i := 0; i < 1000; i++ {
		r.Append("x\n")
	}
	_, truncated := r.Snapshot()
	assert.Equal(t, uint64(0), truncated)
	assert.Equal(t, 1000, r.Len())
}

func TestRingSnapshotIsIndependent(t *testing.T) {
	r := tui.NewRing(0)
	r.Append("a\n")
	snap1, _ := r.Snapshot()
	r.Append("b\n")
	// snap1 must not reflect the later append.
	assert.Equal(t, []string{"a\n"}, snap1)
}

func TestRingConcurrentAppendAndSnapshot(t *testing.T) {
	// Ring is documented safe for one writer and one reader. Run both
	// under -race; the goroutines interleave however the scheduler
	// chooses, and the test passes iff no race fires.
	r := tui.NewRing(0)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			r.Append("x\n")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			_, _ = r.Snapshot()
		}
	}()
	wg.Wait()
}

func TestBudgetEvictsFromLargestRing(t *testing.T) {
	// Two rings: one spammy (100 * "x\n" = 200 bytes), one quiet
	// (1 * "q\n" = 2 bytes). Budget cap of 150 bytes. After Enforce
	// the quiet ring should be untouched and the spammy ring should
	// be trimmed.
	spam := tui.NewRing(0)
	quiet := tui.NewRing(0)
	b := tui.NewBudget(150)
	b.Register(spam)
	b.Register(quiet)

	for i := 0; i < 100; i++ {
		spam.Append("x\n")
	}
	quiet.Append("q\n")
	b.Enforce()

	assert.LessOrEqual(t, spam.Bytes()+quiet.Bytes(), 150)
	// Quiet ring's content must survive — this is the policy's point.
	quietLines, _ := quiet.Snapshot()
	assert.Equal(t, []string{"q\n"}, quietLines)
}

func TestBudgetZeroCapIsInert(t *testing.T) {
	r := tui.NewRing(0)
	b := tui.NewBudget(0)
	b.Register(r)
	for i := 0; i < 1000; i++ {
		r.Append("x\n")
	}
	b.Enforce()
	assert.Equal(t, 1000, r.Len())
}

func TestBudgetWithNoRegisteredRings(t *testing.T) {
	// Enforce on an empty budget is a no-op; no panic.
	b := tui.NewBudget(100)
	b.Enforce()
}

func TestRingWriterSplitsOnNewlines(t *testing.T) {
	r := tui.NewRing(0)
	w := tui.NewRingWriter(r)

	n, err := w.Write([]byte("line1\nline2\nline3\n"))
	require.NoError(t, err)
	assert.Equal(t, 18, n, "Write must return the full byte count per io.Writer contract")

	lines, _ := r.Snapshot()
	assert.Equal(t, []string{"line1\n", "line2\n", "line3\n"}, lines)
}

func TestRingWriterBuffersIncompleteLine(t *testing.T) {
	r := tui.NewRing(0)
	w := tui.NewRingWriter(r)

	// Split "hello world\n" across two writes with no newline in the
	// first half.
	_, err := w.Write([]byte("hello "))
	require.NoError(t, err)
	// Nothing emitted yet.
	lines, _ := r.Snapshot()
	assert.Empty(t, lines)

	_, err = w.Write([]byte("world\n"))
	require.NoError(t, err)
	lines, _ = r.Snapshot()
	assert.Equal(t, []string{"hello world\n"}, lines)
}

func TestRingWriterFlushEmitsPartialLine(t *testing.T) {
	r := tui.NewRing(0)
	w := tui.NewRingWriter(r)
	_, err := w.Write([]byte("no trailing newline"))
	require.NoError(t, err)
	// Before flush: buffered.
	lines, _ := r.Snapshot()
	assert.Empty(t, lines)
	w.Flush()
	lines, _ = r.Snapshot()
	assert.Equal(t, []string{"no trailing newline"}, lines)
}

func TestRingWriterFlushNoopWhenEmpty(t *testing.T) {
	r := tui.NewRing(0)
	w := tui.NewRingWriter(r)
	w.Flush()
	lines, _ := r.Snapshot()
	assert.Empty(t, lines)
}
