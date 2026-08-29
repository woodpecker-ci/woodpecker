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

package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestLintDirPrefersDefaultWoodpeckerDirectory(t *testing.T) {
	repoRoot := t.TempDir()

	writeFile(t, filepath.Join(repoRoot, ".woodpecker", "pipeline.yaml"), `steps:
  test:
    image: alpine
    commands:
      - echo ok
`)
	writeFile(t, filepath.Join(repoRoot, "compose.yaml"), `services:
  database:
    image: postgres
`)

	runLintCommand(t, func(c *cli.Command) {
		require.NoError(t, lintDir(t.Context(), c, repoRoot))
	})
}

func runLintCommand(t *testing.T, fn func(c *cli.Command)) {
	t.Helper()

	command := &cli.Command{
		Flags: Command.Flags,
		Action: func(_ context.Context, c *cli.Command) error {
			fn(c)
			return nil
		},
	}
	require.NoError(t, command.Run(t.Context(), []string{"woodpecker-cli"}))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
