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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestLintPrefersWoodpeckerDirectory(t *testing.T) {
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

	runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.NoError(t, lint(t.Context(), c))
	})
}

func TestLintWithFileConfig(t *testing.T) {
	repoRoot := t.TempDir()

	writeFile(t, filepath.Join(repoRoot, ".woodpecker.yml"), `steps:
  test:
    image: alpine
    commands:
      - echo ok
`)

	runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.NoError(t, lint(t.Context(), c))
	})
}

func TestLintWalksAllYAMLWhenNoConfig(t *testing.T) {
	repoRoot := t.TempDir()

	writeFile(t, filepath.Join(repoRoot, "pipeline-a.yaml"), `steps:
  test:
    image: alpine
    commands:
      - echo a
`)
	writeFile(t, filepath.Join(repoRoot, "pipeline-b.yml"), `steps:
  test:
    image: alpine
    commands:
      - echo b
`)

	runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.NoError(t, lint(t.Context(), c))
	})
}

func TestLintKeepsUnknownVariables(t *testing.T) {
	repoRoot := t.TempDir()

	// values of variables are unknown while linting, so they must not end up empty
	writeFile(t, filepath.Join(repoRoot, ".woodpecker.yaml"), `when:
  - branch:
      - ${CI_REPO_DEFAULT_BRANCH}
      - release/*

steps:
  test:
    image: alpine
    commands:
      - echo ok
`)

	runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.NoError(t, lint(t.Context(), c))
	})
}

func TestLintReportsSubstitutionErrors(t *testing.T) {
	repoRoot := t.TempDir()

	writeFile(t, filepath.Join(repoRoot, ".woodpecker.yaml"), `steps:
  test:
    image: alpine
    commands:
      - echo ${CI_COMMIT_SHA
`)

	out := runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.EqualError(t, lint(t.Context(), c), ".woodpecker.yaml has errors")
	})
	assert.Contains(t, out, "environment variable substitution failed")
}

func TestLintDirNamesBrokenFiles(t *testing.T) {
	repoRoot := t.TempDir()

	writeFile(t, filepath.Join(repoRoot, ".woodpecker", "good.yaml"), `when:
  event: push
steps:
  test:
    image: alpine
    commands:
      - echo ok
`)
	writeFile(t, filepath.Join(repoRoot, ".woodpecker", "broken.yaml"), `steps:
  test:
    image: alpine
    commands:
      - echo ${CI_COMMIT_SHA
`)

	out := runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.EqualError(t, lint(t.Context(), c), "ERRORS: broken.yaml has errors")
	})
	assert.Contains(t, out, "✅ good.yaml is valid")
	assert.Contains(t, out, "🔥 broken.yaml has 1 errors:\n└── ❌ compile time environment variable substitution failed")
}

func TestLintWorkflowDependsOnOnlyForDir(t *testing.T) {
	repoRoot := t.TempDir()

	file := filepath.Join(repoRoot, ".woodpecker", "test.yaml")
	writeFile(t, file, `when:
  event: push
depends_on: [build]
steps:
  test:
    image: alpine
    commands:
      - echo ok
`)

	// a single file does not know the other workflows
	runLintCommand(t, file, func(c *cli.Command) {
		require.NoError(t, lint(t.Context(), c))
	})

	out := runLintCommand(t, repoRoot, func(c *cli.Command) {
		require.EqualError(t, lint(t.Context(), c), "ERRORS: test.yaml has errors")
	})
	assert.Contains(t, out, `🔥 test.yaml has 1 errors:
└── ❌ depends_on  Workflow "build" does not exist`)
}

// runLintCommand runs fn as the action of a lint command for dir and returns its output.
func runLintCommand(t *testing.T, dir string, fn func(c *cli.Command)) string {
	t.Helper()

	var out strings.Builder
	command := &cli.Command{
		Writer: &out,
		Flags:  Command.Flags,
		Action: func(_ context.Context, c *cli.Command) error {
			fn(c)
			return nil
		},
	}
	require.NoError(t, command.Run(t.Context(), []string{"woodpecker-cli", dir}))
	return out.String()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
