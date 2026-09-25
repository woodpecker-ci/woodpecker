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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
)

func TestExecDummy(t *testing.T) {
	repoDir := t.TempDir()
	workflowPath := filepath.Join(repoDir, "workflow.yaml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(`
when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo hello
`), 0o600))

	// LineWriter writes to os.Stderr directly, so redirect the fd
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	t.Cleanup(func() {
		os.Stderr = oldStderr
	})

	// This is important, else it will work on your system but if run in woodpecker,
	// the exec will use the metadata the current test is running in.
	clearEnv(t)

	err = Command.Run(t.Context(), []string{
		"woodpecker-cli",
		"--backend-engine", "dummy",
		"--repo-path", repoDir,
		workflowPath,
	})
	require.NoError(t, err)

	// close write end so Read below doesn't block
	w.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	r.Close()
	stdout := buf.String()

	assert.Contains(
		t, stdout,
		`[build:L0:0s] StepName: build
[build:L1:0s] StepType: commands
[build:L2:0s] StepUUID: `,
	)
	assert.Contains(
		t, stdout,
		`[build:L3:0s] StepCommands:
[build:L4:0s] ------------------
[build:L5:0s] echo hello
[build:L6:0s] ------------------`,
	)

	require.NoError(t, err)
}

func clearEnv(t *testing.T) {
	t.Helper()
	osEnv := os.Environ()
	t.Cleanup(func() {
		for _, env := range osEnv {
			k, v, _ := strings.Cut(env, "=")
			_ = os.Setenv(k, v) //nolint:usetesting
		}
	})
	os.Clearenv()
}

func TestRepoRootFromFile(t *testing.T) {
	tmp := t.TempDir()

	// pipeline file inside a `.woodpecker` config folder => repo root is the parent
	wpDir := filepath.Join(tmp, "myrepo", ".woodpecker")
	require.NoError(t, os.MkdirAll(wpDir, 0o755))
	wpFile := filepath.Join(wpDir, "securityscan.yaml")
	require.NoError(t, os.WriteFile(wpFile, []byte("steps: {}\n"), 0o644))
	assert.Equal(t, filepath.Join(tmp, "myrepo"), repoRootFromFile(wpFile))

	// legacy single config file at repo root => repo root is the file's directory
	rootDir := filepath.Join(tmp, "legacy")
	require.NoError(t, os.MkdirAll(rootDir, 0o755))
	rootFile := filepath.Join(rootDir, ".woodpecker.yml")
	require.NoError(t, os.WriteFile(rootFile, []byte("steps: {}\n"), 0o644))
	assert.Equal(t, rootDir, repoRootFromFile(rootFile))

	// arbitrary file not in a `.woodpecker` folder => its own directory
	otherFile := filepath.Join(rootDir, "ci.yaml")
	require.NoError(t, os.WriteFile(otherFile, []byte("steps: {}\n"), 0o644))
	assert.Equal(t, rootDir, repoRootFromFile(otherFile))
}

// TestExecStepVolumesReferenceWorkflowVolume ensures that the workspace volume
// mounted into every step is the same volume the backend creates from
// config.Volume, and that the local repository path is still mounted for
// backends that run on the host.
func TestExecStepVolumesReferenceWorkflowVolume(t *testing.T) {
	const (
		workspaceBase = "/woodpecker"
		workspacePath = "src"
	)

	repoDir := t.TempDir()
	workflowPath := filepath.Join(repoDir, "workflow.yaml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(`when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo hello
`), 0o600))

	engine := mocks.NewMockBackend(t)
	engine.On("Name").Return("dummy")
	engine.On("Load", mock.Anything).Return(&backend_types.BackendInfo{}, nil)

	var workflowVolume string
	var stepVolumes [][]string

	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			workflowVolume = args.Get(1).(*backend_types.Config).Volume
		}).
		Return(nil)
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stepVolumes = append(stepVolumes, args.Get(1).(*backend_types.Step).Volumes)
		}).
		Return(nil)
	engine.On("TailStep", mock.Anything, mock.Anything, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)
	engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
		Return(&backend_types.State{Exited: true}, nil)
	engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	originalBackends := backends
	backends = []backend_types.Backend{engine}
	t.Cleanup(func() { backends = originalBackends })

	// This is important, else the metadata of the running test leaks
	// into the exec command.
	clearEnv(t)

	require.NoError(t, Command.Run(t.Context(), []string{
		"woodpecker-cli",
		"--backend-engine", "dummy",
		"--repo-path", repoDir,
		workflowPath,
	}))

	require.NotEmpty(t, workflowVolume, "backend should create a workflow volume")
	require.NotEmpty(t, stepVolumes, "backend should start at least one step")

	workspaceMount := workflowVolume + ":" + workspaceBase
	repoMount := repoDir + ":" + workspaceBase + "/" + workspacePath

	for _, volumes := range stepVolumes {
		assert.Contains(t, volumes, workspaceMount,
			"steps must mount the workflow volume the backend creates from config.Volume")
		assert.Contains(t, volumes, repoMount,
			"steps must still mount the local repository for backends running on the host")
	}
}
