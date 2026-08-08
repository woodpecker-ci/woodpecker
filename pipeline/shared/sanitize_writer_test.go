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

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeWriterCustomFunc(t *testing.T) {
	var buf bytes.Buffer
	w := NewSanitizeWriter(&buf, strings.ToUpper)

	n, err := w.Write([]byte("hello\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("hello\n"), n)
	assert.Equal(t, "HELLO\n", buf.String())
}

func TestSanitizeWriterNilFuncPassesThrough(t *testing.T) {
	var buf bytes.Buffer
	w := NewSanitizeWriter(&buf, nil)

	_, err := w.Write([]byte("as is\n"))
	assert.NoError(t, err)
	assert.Equal(t, "as is\n", buf.String())
}

type errWriter struct{ err error }

func (e *errWriter) Write([]byte) (int, error) { return 0, e.err }

func TestSanitizeWriterPropagatesError(t *testing.T) {
	wantErr := errors.New("sink closed")
	w := NewSanitizeWriter(&errWriter{err: wantErr}, nil)

	n, err := w.Write([]byte("x"))
	assert.ErrorIs(t, err, wantErr)
	assert.Zero(t, n)
}
