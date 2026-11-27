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

package local

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestCopyWorkspace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need extra rights on windows")
	}
	src := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub", "deep"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(src, "file"), []byte("content"), 0o640))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "deep", "run.sh"), []byte("#!/bin/sh"), 0o750))
	require.NoError(t, os.Symlink("file", filepath.Join(src, "link")))

	dst := filepath.Join(t.TempDir(), "copy")
	require.NoError(t, copyWorkspace(src, dst))

	data, err := os.ReadFile(filepath.Join(dst, "file"))
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
	info, err := os.Stat(filepath.Join(dst, "sub", "deep", "run.sh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o750), info.Mode().Perm())
	target, err := os.Readlink(filepath.Join(dst, "link"))
	require.NoError(t, err)
	assert.Equal(t, "file", target)

	// changes in the copy do not reach the source
	require.NoError(t, os.WriteFile(filepath.Join(dst, "file"), []byte("changed"), 0o640))
	data, err = os.ReadFile(filepath.Join(src, "file"))
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}

func TestStartStepWorkspaceReadOnly(t *testing.T) {
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()
	taskUUID := "test-workspace-readonly"
	require.NoError(t, backend.SetupWorkflow(t.Context(), &types.Config{}, taskUUID))

	err := backend.StartStep(t.Context(), &types.Step{
		UUID:          "step-ro",
		Name:          "lint",
		Type:          types.StepTypeCommands,
		Image:         "sh",
		Commands:      []string{"echo hello"},
		WorkspaceMode: types.WorkspaceModeReadOnly,
	}, taskUUID)
	assert.ErrorContains(t, err, "can not mount the workspace read-only")
}

func TestStartStepWorkspaceOverlay(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping on non linux due to shell availability")
	}
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()
	taskUUID := "test-workspace-overlay"
	require.NoError(t, backend.SetupWorkflow(t.Context(), &types.Config{}, taskUUID))
	state, err := backend.getWorkflowState(taskUUID)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(state.workspaceDir, "file"), []byte("original"), 0o640))

	step := &types.Step{
		UUID:          "step-overlay",
		Name:          "overlay",
		Type:          types.StepTypeCommands,
		Image:         "sh",
		Commands:      []string{`echo changed > "$CI_WORKSPACE/file"`, `touch new`},
		WorkspaceMode: types.WorkspaceModeOverlay,
	}
	require.NoError(t, backend.StartStep(t.Context(), step, taskUUID))
	stepState, err := backend.WaitStep(t.Context(), step, taskUUID)
	require.NoError(t, err)
	assert.Equal(t, 0, stepState.ExitCode)

	overlayDir := filepath.Join(state.baseDir, "overlay", step.UUID)
	data, err := os.ReadFile(filepath.Join(overlayDir, "file"))
	require.NoError(t, err)
	assert.Equal(t, "changed\n", string(data))
	assert.FileExists(t, filepath.Join(overlayDir, "new"))

	// the workspace keeps its content
	data, err = os.ReadFile(filepath.Join(state.workspaceDir, "file"))
	require.NoError(t, err)
	assert.Equal(t, "original", string(data))
	assert.NoFileExists(t, filepath.Join(state.workspaceDir, "new"))

	require.NoError(t, backend.DestroyStep(t.Context(), step, taskUUID))
	assert.NoDirExists(t, overlayDir)
}
