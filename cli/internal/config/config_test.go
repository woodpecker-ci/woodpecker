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

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
	"github.com/zalando/go-keyring"
)

const testCLIName = "woodpecker-cli"

// runLoad runs Load for a "repo" subcommand with the given global flags and
// returns the resulting command so the loaded flag values can be inspected.
func runLoad(t *testing.T, args ...string) (*cli.Command, error) {
	t.Helper()

	var loadErr error
	cmd := &cli.Command{
		Name: testCLIName,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server"},
			&cli.StringFlag{Name: "token"},
			&cli.StringFlag{Name: "log-level"},
			&cli.StringFlag{Name: "contexts-file"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			loadErr = Load(ctx, c)
			return nil
		},
	}

	args = append(append([]string{testCLIName}, args...), "repo")
	require.NoError(t, cmd.Run(t.Context(), args))
	return cmd, loadErr
}

func TestLoad(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	xdg.Reload()
	keyring.MockInit()

	t.Run("not configured without context or flags", func(t *testing.T) {
		_, err := runLoad(t)
		assert.ErrorContains(t, err, "woodpecker-cli is not configured")
	})

	t.Run("server and token flags work without context", func(t *testing.T) {
		cmd, err := runLoad(t, "--server", "https://flags.example.test", "--token", "flag-token")
		require.NoError(t, err)
		assert.Equal(t, "https://flags.example.test", cmd.String("server"))
		assert.Equal(t, "flag-token", cmd.String("token"))
	})

	t.Run("legacy config file is ignored", func(t *testing.T) {
		legacyFile := filepath.Join(configDir, "woodpecker", "config.json")
		require.NoError(t, os.MkdirAll(filepath.Dir(legacyFile), 0o755))
		require.NoError(t, os.WriteFile(legacyFile, []byte(`{"server_url":"https://legacy.example.test"}`), 0o600))
		require.NoError(t, keyring.Set(testCLIName, "https://legacy.example.test", "legacy-token"))

		_, err := runLoad(t)
		assert.ErrorContains(t, err, "woodpecker-cli is not configured")
	})

	require.NoError(t, AddOrUpdateContext(&cli.Command{Name: testCLIName}, "prod", "https://ctx.example.test", "ctx-token", "debug", true))

	t.Run("current context fills server, token and log level", func(t *testing.T) {
		cmd, err := runLoad(t)
		require.NoError(t, err)
		assert.Equal(t, "https://ctx.example.test", cmd.String("server"))
		assert.Equal(t, "ctx-token", cmd.String("token"))
		assert.Equal(t, "debug", cmd.String("log-level"))
	})

	t.Run("flags take precedence over current context", func(t *testing.T) {
		cmd, err := runLoad(t, "--server", "https://flags.example.test", "--token", "flag-token", "--log-level", "warn")
		require.NoError(t, err)
		assert.Equal(t, "https://flags.example.test", cmd.String("server"))
		assert.Equal(t, "flag-token", cmd.String("token"))
		assert.Equal(t, "warn", cmd.String("log-level"))
	})

	t.Run("contexts file flag replaces the default contexts file", func(t *testing.T) {
		contextsFile := filepath.Join(t.TempDir(), "custom.json")
		require.NoError(t, os.WriteFile(contextsFile, []byte(`{"current_context":"custom","contexts":{"custom":{"name":"custom","server_url":"https://custom.example.test"}}}`), 0o600))
		require.NoError(t, keyring.Set(testCLIName, "https://custom.example.test", "custom-token"))

		cmd, err := runLoad(t, "--contexts-file", contextsFile)
		require.NoError(t, err)
		assert.Equal(t, "https://custom.example.test", cmd.String("server"))
		assert.Equal(t, "custom-token", cmd.String("token"))
	})
}
