// Copyright 2022 Woodpecker Authors
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

//go:build linux
// +build linux

package local

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestIsAvailable(t *testing.T) {
	t.Run("not available in container", func(t *testing.T) {
		backend := New()

		os.Setenv("WOODPECKER_IN_CONTAINER", "true")
		defer os.Unsetenv("WOODPECKER_IN_CONTAINER")

		available := backend.IsAvailable(context.Background())
		assert.False(t, available)
	})

	t.Run("available without container env and no cli context", func(t *testing.T) {
		backend := New()

		os.Unsetenv("WOODPECKER_IN_CONTAINER")
		available := backend.IsAvailable(context.Background())
		assert.True(t, available)
	})
}

func TestLoad(t *testing.T) {
	backend := New().(*local)

	t.Run("load without cli context", func(t *testing.T) {
		ctx := context.Background()
		info, err := backend.Load(ctx)

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
	})

	t.Run("load with cli context and temp dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := &cli.Command{}
		cmd.Flags = []cli.Flag{
			&cli.StringFlag{
				Name:  "backend-local-temp-dir",
				Value: tmpDir,
			},
		}
		ctx := context.WithValue(context.Background(), types.CliCommand, cmd)

		info, err := backend.Load(ctx)

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, tmpDir, backend.tempDir)
		assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
	})
}

func TestSetupWorkflow(t *testing.T) {
	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-task-uuid-123"
	config := &types.Config{}

	err := backend.SetupWorkflow(ctx, config, taskUUID)
	require.NoError(t, err)

	// Verify state was saved
	state, err := backend.getState(taskUUID)
	require.NoError(t, err)
	assert.NotNil(t, state)
	assert.NotEmpty(t, state.baseDir)
	assert.NotEmpty(t, state.workspaceDir)
	assert.NotEmpty(t, state.homeDir)
	assert.NotNil(t, state.stepCMDs)

	// Verify directories were created
	assert.DirExists(t, state.baseDir)
	assert.DirExists(t, state.workspaceDir)
	assert.DirExists(t, state.homeDir)

	// Verify directory structure
	assert.Equal(t, filepath.Join(state.baseDir, "workspace"), state.workspaceDir)
	assert.Equal(t, filepath.Join(state.baseDir, "home"), state.homeDir)

	// Cleanup
	os.RemoveAll(state.baseDir)
}

func TestDestroyWorkflow(t *testing.T) {
	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-destroy-task"
	config := &types.Config{}

	// Setup workflow first
	err := backend.SetupWorkflow(ctx, config, taskUUID)
	require.NoError(t, err)

	state, err := backend.getState(taskUUID)
	require.NoError(t, err)
	baseDir := state.baseDir

	// Verify directory exists
	assert.DirExists(t, baseDir)

	// Destroy workflow
	err = backend.DestroyWorkflow(ctx, config, taskUUID)
	require.NoError(t, err)

	// Verify directory was removed
	assert.NoDirExists(t, baseDir)

	// Verify state was deleted
	_, err = backend.getState(taskUUID)
	assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
}

func TestStartStepCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows due to shell availability")
	}

	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-commands-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	step := &types.Step{
		UUID:     "step-1",
		Name:     "test-step",
		Type:     types.StepTypeCommands,
		Image:    "sh",
		Commands: []string{"echo hello", "pwd"},
		Environment: map[string]string{
			"TEST_VAR": "test_value",
		},
	}

	err = backend.StartStep(ctx, step, taskUUID)
	require.NoError(t, err)

	// Verify command was started
	state, err := backend.getState(taskUUID)
	require.NoError(t, err)
	assert.Contains(t, state.stepCMDs, step.UUID)

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestStartStepPlugin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-plugin-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	step := &types.Step{
		UUID:        "step-plugin-1",
		Name:        "test-plugin",
		Type:        types.StepTypePlugin,
		Image:       "echo", // Use a binary that exists
		Environment: map[string]string{},
	}

	err = backend.StartStep(ctx, step, taskUUID)
	require.NoError(t, err)

	// Verify command was started
	state, err := backend.getState(taskUUID)
	require.NoError(t, err)
	assert.Contains(t, state.stepCMDs, step.UUID)

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestStartStepUnsupportedType(t *testing.T) {
	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-unsupported-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	step := &types.Step{
		UUID: "step-unsupported",
		Name: "test-unsupported",
		Type: "unsupported-type",
	}

	err = backend.StartStep(ctx, step, taskUUID)
	assert.ErrorIs(t, err, ErrUnsupportedStepType)

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestWaitStep(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-wait-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	t.Run("successful step", func(t *testing.T) {
		step := &types.Step{
			UUID:     "step-success",
			Name:     "success-step",
			Type:     types.StepTypeCommands,
			Image:    "sh",
			Commands: []string{"echo success"},
		}

		err = backend.StartStep(ctx, step, taskUUID)
		require.NoError(t, err)

		state, err := backend.WaitStep(ctx, step, taskUUID)
		require.NoError(t, err)
		assert.True(t, state.Exited)
		assert.Equal(t, 0, state.ExitCode)
	})

	t.Run("failed step", func(t *testing.T) {
		step := &types.Step{
			UUID:     "step-fail",
			Name:     "fail-step",
			Type:     types.StepTypeCommands,
			Image:    "sh",
			Commands: []string{"exit 1"},
		}

		err = backend.StartStep(ctx, step, taskUUID)
		require.NoError(t, err)

		state, err := backend.WaitStep(ctx, step, taskUUID)
		require.NoError(t, err)
		assert.True(t, state.Exited)
		assert.Equal(t, 1, state.ExitCode)
	})

	t.Run("step not found", func(t *testing.T) {
		step := &types.Step{
			UUID: "nonexistent-step",
			Name: "missing",
		}

		_, err = backend.WaitStep(ctx, step, taskUUID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestTailStep(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-tail-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	step := &types.Step{
		UUID:     "step-tail",
		Name:     "tail-step",
		Type:     types.StepTypeCommands,
		Image:    "sh",
		Commands: []string{"echo 'test output'"},
	}

	err = backend.StartStep(ctx, step, taskUUID)
	require.NoError(t, err)

	output, err := backend.TailStep(ctx, step, taskUUID)
	require.NoError(t, err)
	assert.NotNil(t, output)

	// Read output
	data, err := io.ReadAll(output)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test output")

	// Wait for step to complete
	backend.WaitStep(ctx, step, taskUUID)

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestDestroyStep(t *testing.T) {
	backend := New().(*local)

	ctx := context.Background()
	step := &types.Step{UUID: "test-step"}

	// DestroyStep should not return error (it's a no-op)
	err := backend.DestroyStep(ctx, step, "task-uuid")
	assert.NoError(t, err)
}

func TestStateManagement(t *testing.T) {
	backend := New().(*local)

	t.Run("save and get state", func(t *testing.T) {
		taskUUID := "test-state-uuid"
		state := &workflowState{
			stepCMDs:     make(map[string]*exec.Cmd),
			baseDir:      "/tmp/test",
			homeDir:      "/tmp/test/home",
			workspaceDir: "/tmp/test/workspace",
		}

		backend.saveState(taskUUID, state)

		retrieved, err := backend.getState(taskUUID)
		require.NoError(t, err)
		assert.Equal(t, state.baseDir, retrieved.baseDir)
		assert.Equal(t, state.homeDir, retrieved.homeDir)
		assert.Equal(t, state.workspaceDir, retrieved.workspaceDir)
	})

	t.Run("get nonexistent state", func(t *testing.T) {
		_, err := backend.getState("nonexistent-uuid")
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	})

	t.Run("delete state", func(t *testing.T) {
		taskUUID := "test-delete-uuid"
		state := &workflowState{
			stepCMDs: make(map[string]*exec.Cmd),
		}

		backend.saveState(taskUUID, state)

		// Verify state exists
		_, err := backend.getState(taskUUID)
		require.NoError(t, err)

		// Delete state
		backend.deleteState(taskUUID)

		// Verify state is gone
		_, err = backend.getState(taskUUID)
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	})
}

func TestEnvironmentVariables(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-env-task"

	// Setup workflow
	err := backend.SetupWorkflow(ctx, &types.Config{}, taskUUID)
	require.NoError(t, err)

	state, _ := backend.getState(taskUUID)

	step := &types.Step{
		UUID:     "step-env",
		Name:     "env-step",
		Type:     types.StepTypeCommands,
		Image:    "sh",
		Commands: []string{"env | grep -E '(HOME|CI_WORKSPACE|CUSTOM_VAR)'"},
		Environment: map[string]string{
			"CUSTOM_VAR": "custom_value",
		},
	}

	err = backend.StartStep(ctx, step, taskUUID)
	require.NoError(t, err)

	// Wait and check output
	output, _ := backend.TailStep(ctx, step, taskUUID)
	data, _ := io.ReadAll(output)
	outputStr := string(data)

	backend.WaitStep(ctx, step, taskUUID)

	// Verify HOME and CI_WORKSPACE are set
	assert.Contains(t, outputStr, "HOME="+state.homeDir)
	assert.Contains(t, outputStr, "CI_WORKSPACE="+state.workspaceDir)
	assert.Contains(t, outputStr, "CUSTOM_VAR=custom_value")

	// Cleanup
	backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
}

func TestConcurrentWorkflows(t *testing.T) {
	backend := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()

	// Create multiple workflows concurrently
	taskUUIDs := []string{"task-1", "task-2", "task-3"}

	for _, uuid := range taskUUIDs {
		err := backend.SetupWorkflow(ctx, &types.Config{}, uuid)
		require.NoError(t, err)
	}

	// Verify all states exist
	for _, uuid := range taskUUIDs {
		state, err := backend.getState(uuid)
		require.NoError(t, err)
		assert.NotNil(t, state)
	}

	// Cleanup all workflows
	for _, uuid := range taskUUIDs {
		err := backend.DestroyWorkflow(ctx, &types.Config{}, uuid)
		require.NoError(t, err)
	}

	// Verify all states are deleted
	for _, uuid := range taskUUIDs {
		_, err := backend.getState(uuid)
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	}
}
