// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import "io"

// SanitizeFunc transforms log data before it is written, e.g. to mask
// secrets. It receives and returns a full write chunk (typically one line).
type SanitizeFunc func(string) string

// sanitizeWriter is an io.Writer decorator that runs a SanitizeFunc on the
// data before forwarding it to the wrapped writer. It is meant to wrap a
// line-oriented writer so that per-line transformations behave correctly.
type sanitizeWriter struct {
	dst      io.Writer
	sanitize SanitizeFunc
}

// NewSanitizeWriter wraps dst so that every write is passed through the given
// sanitize function first. A nil function passes data through unchanged. The
// returned writer reports the number of input bytes consumed, so it composes
// transparently with callers that check n against the input length.
func NewSanitizeWriter(dst io.Writer, sanitize SanitizeFunc) io.Writer {
	return &sanitizeWriter{
		dst:      dst,
		sanitize: sanitize,
	}
}

func (w *sanitizeWriter) Write(p []byte) (n int, err error) {
	data := p
	if w.sanitize != nil {
		data = []byte(w.sanitize(string(p)))
	}
	if _, err := w.dst.Write(data); err != nil {
		return 0, err
	}
	return len(p), nil
}
