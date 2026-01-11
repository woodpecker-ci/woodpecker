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

package config

import (
	"os"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextManagement(t *testing.T) {
	// Create a temporary directory for test contexts
	tmpDir := t.TempDir()

	// Override xdg directories for testing
	t.Setenv("HOME", tmpDir)
	xdg.Reload()
	contextsFile, err := xdg.ConfigFile("woodpecker/contexts.json")
	require.NoError(t, err)

	t.Run("LoadContexts returns empty when file doesn't exist", func(t *testing.T) {
		contexts, err := LoadContexts()
		require.NoError(t, err)
		assert.NotNil(t, contexts)
		assert.Empty(t, contexts.Contexts)
		assert.Empty(t, contexts.CurrentContext)
	})

	t.Run("SaveContexts creates valid JSON", func(t *testing.T) {
		contexts := &Contexts{
			CurrentContext: "test",
			Contexts: map[string]Context{
				"test": {
					Name:      "test",
					ServerURL: "https://test.example.com",
					LogLevel:  "info",
				},
			},
		}

		err := SaveContexts(contexts)
		require.NoError(t, err)

		// Verify file exists and contains valid JSON
		data, err := os.ReadFile(contextsFile)
		require.NoError(t, err)
		assert.Contains(t, string(data), "test.example.com")
	})

	t.Run("LoadContexts reads saved contexts", func(t *testing.T) {
		contexts, err := LoadContexts()
		require.NoError(t, err)
		assert.Equal(t, "test", contexts.CurrentContext)
		assert.Len(t, contexts.Contexts, 1)
		assert.Equal(t, "https://test.example.com", contexts.Contexts["test"].ServerURL)
	})

	t.Run("SetCurrentContext updates current context", func(t *testing.T) {
		contexts := &Contexts{
			CurrentContext: "test",
			Contexts: map[string]Context{
				"test": {
					Name:      "test",
					ServerURL: "https://test.example.com",
				},
				"prod": {
					Name:      "prod",
					ServerURL: "https://prod.example.com",
				},
			},
		}
		err := SaveContexts(contexts)
		require.NoError(t, err)

		err = SetCurrentContext("prod")
		require.NoError(t, err)

		contexts, err = LoadContexts()
		require.NoError(t, err)
		assert.Equal(t, "prod", contexts.CurrentContext)
	})

	t.Run("SetCurrentContext fails for non-existent context", func(t *testing.T) {
		err := SetCurrentContext("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("RenameContext updates context name", func(t *testing.T) {
		contexts := &Contexts{
			CurrentContext: "old",
			Contexts: map[string]Context{
				"old": {
					Name:      "old",
					ServerURL: "https://test.example.com",
				},
			},
		}
		err := SaveContexts(contexts)
		require.NoError(t, err)

		err = RenameContext("old", "new")
		require.NoError(t, err)

		contexts, err = LoadContexts()
		require.NoError(t, err)
		assert.Equal(t, "new", contexts.CurrentContext)
		assert.Contains(t, contexts.Contexts, "new")
		assert.NotContains(t, contexts.Contexts, "old")
		assert.Equal(t, "new", contexts.Contexts["new"].Name)
	})

	t.Run("RenameContext fails if target exists", func(t *testing.T) {
		contexts := &Contexts{
			Contexts: map[string]Context{
				"ctx1": {Name: "ctx1", ServerURL: "https://test1.example.com"},
				"ctx2": {Name: "ctx2", ServerURL: "https://test2.example.com"},
			},
		}
		err := SaveContexts(contexts)
		require.NoError(t, err)

		err = RenameContext("ctx1", "ctx2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}
