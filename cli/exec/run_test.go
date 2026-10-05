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

//go:build test

package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// execWorkflows runs the exec command with the dummy backend on the given
// workflow files and returns what it printed to stderr.
func execWorkflows(ctx context.Context, t *testing.T, workflows map[string]string) (string, error) {
	t.Helper()

	repoDir := t.TempDir()
	configDir := filepath.Join(repoDir, ".woodpecker")
	require.NoError(t, os.Mkdir(configDir, 0o700))
	for name, content := range workflows {
		require.NoError(t, os.WriteFile(filepath.Join(configDir, name+".yaml"), []byte(content), 0o600))
	}

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = oldStderr })

	output := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output <- buf.String()
	}()

	clearEnv(t)

	runErr := Command.Run(ctx, []string{
		"woodpecker-cli",
		"--backend-engine", "dummy",
		"--repo-path", repoDir,
		configDir,
	})

	w.Close()
	return <-output, runErr
}

func TestExecFollowsWorkflowDependencies(t *testing.T) {
	out, err := execWorkflows(t.Context(), t, map[string]string{
		"build": `
when:
  - event: manual
steps:
  - name: compile
    image: alpine
    commands: make
    environment:
      STEP_EXIT_CODE: 2
  - name: package
    image: alpine
    commands: make package
`,
		"test": `
when:
  - event: manual
depends_on: [build]
steps:
  - name: unit
    image: alpine
    commands: make test
`,
		"notify": `
when:
  - event: manual
    status: [failure]
depends_on: [build]
steps:
  - name: mail
    image: alpine
    commands: mail
`,
	})

	require.EqualError(t, err, "pipeline finished with status failure")

	assert.Contains(t, out, "# build\n")
	assert.Contains(t, out, "[build/compile:L0:0s] StepName: compile\n")
	assert.Contains(t, out, "# build: failure\nstep compile: exit code 2\n")
	assert.NotContains(t, out, "[build/package:", "a step after a failed one must not run")

	assert.Contains(t, out, "# test: skipped\n")
	assert.NotContains(t, out, "[test/unit:", "a workflow must not run if its dependency failed")

	assert.Contains(t, out, "[notify/mail:L0:0s] StepName: mail\n", "a workflow running on failure must run")
}

func TestExecIgnoresFailureOfStepSetToIgnore(t *testing.T) {
	out, err := execWorkflows(t.Context(), t, map[string]string{
		"build": `
when:
  - event: manual
steps:
  - name: flaky
    image: alpine
    commands: make
    failure: ignore
    environment:
      STEP_EXIT_CODE: 1
  - name: package
    image: alpine
    commands: make package
`,
	})

	require.NoError(t, err)
	assert.Contains(t, out, "[package:L0:0s] StepName: package\n")
}

func TestExecIgnoresFailureOfService(t *testing.T) {
	_, err := execWorkflows(t.Context(), t, map[string]string{
		"test": `
when:
  - event: manual
steps:
  - name: unit
    image: alpine
    commands: make test
    environment:
      SLEEP: 500ms
services:
  - name: database
    image: alpine
    environment:
      SLEEP: 1ms
      STEP_EXIT_CODE: 1
`,
	})

	require.NoError(t, err, "a server ignores it by default as well")
}

func TestExecCancelsPipelineIfStepSetToCancelFails(t *testing.T) {
	start := time.Now()
	out, err := execWorkflows(t.Context(), t, map[string]string{
		"build": `
when:
  - event: manual
steps:
  - name: wait
    image: alpine
    commands: sleep
    environment:
      SLEEP: 500ms
  - name: compile
    image: alpine
    commands: make
    failure: cancel
    environment:
      STEP_EXIT_CODE: 1
`,
		"long": `
when:
  - event: manual
steps:
  - name: sleep
    image: alpine
    commands: sleep
    environment:
      SLEEP: 1m
`,
		"deploy": `
when:
  - event: manual
depends_on: [build]
steps:
  - name: upload
    image: alpine
    commands: upload
`,
	})

	require.EqualError(t, err, "pipeline finished with status killed")
	assert.Less(t, time.Since(start), 30*time.Second, "running workflows must be stopped")

	assert.Contains(t, out, "# long: killed\n")
	assert.Contains(t, out, "# deploy: skipped\n")
	assert.NotContains(t, out, "[deploy/upload:")
}

func TestExecStopsOnCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	time.AfterFunc(500*time.Millisecond, func() { cancel(nil) })

	start := time.Now()
	out, err := execWorkflows(ctx, t, map[string]string{
		"build": `
when:
  - event: manual
steps:
  - name: compile
    image: alpine
    commands: make
    environment:
      SLEEP: 1m
`,
		"deploy": `
when:
  - event: manual
depends_on: [build]
steps:
  - name: upload
    image: alpine
    commands: upload
`,
	})

	require.EqualError(t, err, "pipeline finished with status killed")
	assert.Less(t, time.Since(start), 30*time.Second)
	assert.Contains(t, out, "# build: killed\n")
	assert.Contains(t, out, "# deploy: skipped\n")
}
