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

package exec

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failWriter struct{ err error }

func (f *failWriter) Write([]byte) (int, error) { return 0, f.err }

func TestLineWriterPropagatesError(t *testing.T) {
	wantErr := errors.New("broken pipe")
	w := NewLineWriter("step", "uuid")
	lw, ok := w.(*LineWriter)
	require.True(t, ok)
	lw.out = &failWriter{err: wantErr}

	n, err := w.Write([]byte("line\n"))
	assert.ErrorIs(t, err, wantErr)
	assert.Zero(t, n)
}

func TestLineWriterWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewLineWriter("step", "uuid")
	lw, ok := w.(*LineWriter)
	require.True(t, ok)
	lw.out = &buf

	n, err := w.Write([]byte("hello\n"))
	assert.NoError(t, err)
	assert.Equal(t, len("hello\n"), n)
	assert.Contains(t, buf.String(), "[step:L0:0s] hello")
}
