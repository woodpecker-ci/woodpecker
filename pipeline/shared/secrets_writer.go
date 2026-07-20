// Copyright 2024 Woodpecker Authors
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

import (
	"io"
	"strings"
)

// secretsWriter is an io.Writer decorator that masks secret values before
// forwarding data to the wrapped writer. It is meant to wrap a line-oriented
// writer so that multi-line secrets are masked correctly per line.
type secretsWriter struct {
	dst      io.Writer
	replacer *strings.Replacer
}

// NewSecretsWriter wraps dst so that any of the given secret values are
// replaced with asterisks before being written. The returned writer reports
// the number of input bytes consumed, so it composes transparently with
// callers that check n against the input length.
func NewSecretsWriter(dst io.Writer, secrets []string) io.Writer {
	return &secretsWriter{
		dst:      dst,
		replacer: NewSecretsReplacer(secrets),
	}
}

func (w *secretsWriter) Write(p []byte) (n int, err error) {
	data := p
	if w.replacer != nil {
		data = []byte(w.replacer.Replace(string(p)))
	}
	if _, err := w.dst.Write(data); err != nil {
		return 0, err
	}
	return len(p), nil
}
