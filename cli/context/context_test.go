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

package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func runContextCommand(t *testing.T, args ...string) error {
	t.Helper()

	root := &cli.Command{
		Name:     "woodpecker-cli",
		Flags:    []cli.Flag{&cli.StringFlag{Name: "contexts-file"}},
		Commands: []*cli.Command{Command},
	}
	return root.Run(t.Context(), append([]string{"woodpecker-cli"}, args...))
}

func readContexts(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var contexts map[string]any
	require.NoError(t, json.Unmarshal(data, &contexts))
	return contexts
}

func TestContextsFileFlag(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	xdg.Reload()

	contextsFile := filepath.Join(t.TempDir(), "custom.json")
	require.NoError(t, os.WriteFile(contextsFile, []byte(`{"current_context":"prod","contexts":{"prod":{"name":"prod","server_url":"https://prod.example.test"},"dev":{"name":"dev","server_url":"https://dev.example.test"}}}`), 0o600))

	t.Run("use switches the context in the given file", func(t *testing.T) {
		require.NoError(t, runContextCommand(t, "--contexts-file", contextsFile, "context", "use", "dev"))
		assert.Equal(t, "dev", readContexts(t, contextsFile)["current_context"])
	})

	t.Run("rename renames the context in the given file", func(t *testing.T) {
		require.NoError(t, runContextCommand(t, "--contexts-file", contextsFile, "context", "rename", "dev", "staging"))
		contexts := readContexts(t, contextsFile)
		assert.Equal(t, "staging", contexts["current_context"])
		assert.Contains(t, contexts["contexts"], "staging")
	})

	t.Run("default contexts file stays untouched", func(t *testing.T) {
		assert.NoFileExists(t, filepath.Join(configDir, "woodpecker", "contexts.json"))
	})
}
