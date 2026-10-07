// Copyright 2026 Woodpecker Authors
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

package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPipelineConfig(t *testing.T) {
	t.Run("PrefersWoodpeckerDirectory", func(t *testing.T) {
		repoRoot := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(repoRoot, ".woodpecker"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repoRoot, ".woodpecker.yaml"), []byte("steps: {}\n"), 0o644))

		isDir, config, found, err := FindPipelineConfig(repoRoot)

		require.NoError(t, err)
		assert.True(t, found)
		assert.True(t, isDir)
		assert.Equal(t, filepath.Join(repoRoot, ".woodpecker"), config)
	})

	t.Run("FindsFileConfig", func(t *testing.T) {
		repoRoot := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(repoRoot, ".woodpecker.yml"), []byte("steps: {}\n"), 0o644))

		isDir, config, found, err := FindPipelineConfig(repoRoot)

		require.NoError(t, err)
		assert.True(t, found)
		assert.False(t, isDir)
		assert.Equal(t, filepath.Join(repoRoot, ".woodpecker.yml"), config)
	})

	t.Run("MissingConfig", func(t *testing.T) {
		isDir, config, found, err := FindPipelineConfig(t.TempDir())

		require.NoError(t, err)
		assert.False(t, found)
		assert.False(t, isDir)
		assert.Empty(t, config)
	})
}
