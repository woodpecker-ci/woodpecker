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

package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestNewLogStore(t *testing.T) {
	t.Parallel()

	t.Run("empty base path is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := NewLogStore("")
		assert.Error(t, err)
	})

	t.Run("creates missing base directory", func(t *testing.T) {
		t.Parallel()
		base := filepath.Join(t.TempDir(), "logs", "nested")
		s, err := NewLogStore(base)
		require.NoError(t, err)
		assert.NotNil(t, s)

		info, statErr := os.Stat(base)
		require.NoError(t, statErr)
		assert.True(t, info.IsDir())
	})

	t.Run("accepts existing base directory", func(t *testing.T) {
		t.Parallel()
		s, err := NewLogStore(t.TempDir())
		require.NoError(t, err)
		assert.NotNil(t, s)
	})
}

func TestLogStoreAppendFindDelete(t *testing.T) {
	t.Parallel()

	s, err := NewLogStore(t.TempDir())
	require.NoError(t, err)

	step := &model.Step{ID: 42}
	first := []*model.LogEntry{
		{StepID: 42, Line: 0, Data: []byte("hello"), Type: model.LogEntryStdout},
		{StepID: 42, Line: 1, Data: []byte("world"), Type: model.LogEntryStdout},
	}

	require.NoError(t, s.LogAppend(step, first))

	got, err := s.LogFind(step)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, []byte("hello"), got[0].Data)
	assert.Equal(t, 1, got[1].Line)

	// append must add to the existing file, not overwrite it
	require.NoError(t, s.LogAppend(step, []*model.LogEntry{
		{StepID: 42, Line: 2, Data: []byte("again"), Type: model.LogEntryStdout},
	}))

	got, err = s.LogFind(step)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, []byte("again"), got[2].Data)

	require.NoError(t, s.LogDelete(step))

	// StepFinished is a no-op for the file store but part of the interface
	s.StepFinished(step)

	// after delete the file is gone -> find returns no entries, no error
	got, err = s.LogFind(step)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestLogFindMissingFile(t *testing.T) {
	t.Parallel()

	s, err := NewLogStore(t.TempDir())
	require.NoError(t, err)

	got, err := s.LogFind(&model.Step{ID: 999})
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestLogFindMalformedJSON(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	s, err := NewLogStore(base)
	require.NoError(t, err)

	// write a file with one valid line and one broken line
	path := filepath.Join(base, "7.json")
	require.NoError(t, os.WriteFile(path, []byte("{\"line\":0}\nnot-json\n"), 0o600))

	_, err = s.LogFind(&model.Step{ID: 7})
	assert.Error(t, err)
}

func TestLogFindSkipsBlankLines(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	s, err := NewLogStore(base)
	require.NoError(t, err)

	path := filepath.Join(base, "8.json")
	require.NoError(t, os.WriteFile(path, []byte("{\"line\":0}\n\n  \n{\"line\":1}\n"), 0o600))

	got, err := s.LogFind(&model.Step{ID: 8})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestLogDeleteMissingFile(t *testing.T) {
	t.Parallel()

	s, err := NewLogStore(t.TempDir())
	require.NoError(t, err)

	// removing a non-existent log file surfaces the os.Remove error
	assert.Error(t, s.LogDelete(&model.Step{ID: 12345}))
}
