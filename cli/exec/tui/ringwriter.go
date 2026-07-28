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
	"bytes"
	"io"
	"sync"
)

// RingWriter is an io.Writer that appends incoming bytes to a Ring,
// one Ring line per input newline-delimited record. Intended as a
// zerolog destination when the TUI is active: stderr is owned by the
// alt-screen buffer, so zerolog is redirected here instead.
//
// RingWriter buffers incomplete lines across Write calls. A partial
// final line (no trailing newline) stays in the internal buffer
// until the next Write completes it — zerolog always writes one
// complete JSON event per call, so in practice this buffering is
// defensive.
type RingWriter struct {
	ring *Ring

	mu  sync.Mutex
	buf []byte
}

// NewRingWriter returns an io.Writer that appends into ring.
func NewRingWriter(ring *Ring) *RingWriter {
	return &RingWriter{ring: ring}
}

// Write implements io.Writer. Each newline-terminated segment of p is
// appended to the underlying ring as a separate line (with the
// trailing newline retained so renderers can emit raw bytes). A
// trailing fragment without a newline is buffered for the next call.
//
// Returns len(p) and nil on success, per the io.Writer contract.
func (w *RingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Combine any carried-over fragment with the new bytes. We build
	// a fresh slice here rather than `append(w.buf, p...)` because
	// that pattern aliases w.buf's backing array on the fast path
	// (len(w.buf)==0 appends in place) and that's exactly the kind of
	// bug gocritic's appendAssign rule exists to catch.
	var data []byte
	if len(w.buf) == 0 {
		data = p
	} else {
		data = make([]byte, 0, len(w.buf)+len(p))
		data = append(data, w.buf...)
		data = append(data, p...)
		w.buf = w.buf[:0]
	}

	for len(data) > 0 {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			// No newline yet; stash and wait for the rest.
			w.buf = append(w.buf, data...)
			break
		}
		// i+1 keeps the newline attached to the emitted line, matching
		// the CopyLineByLine convention used elsewhere in the CLI.
		w.ring.Append(string(data[:i+1]))
		data = data[i+1:]
	}
	return len(p), nil
}

// Flush appends any buffered fragment as a final line. Call this
// during teardown to avoid losing the last (unterminated) line of a
// stream — in practice relevant only if the producer crashes
// mid-write.
func (w *RingWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) == 0 {
		return
	}
	w.ring.Append(string(w.buf))
	w.buf = w.buf[:0]
}

// Static type assertion: RingWriter is an io.Writer.
var _ io.Writer = (*RingWriter)(nil)
