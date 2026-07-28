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

package nsjail

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestName(t *testing.T) {
	backend := New()
	assert.Equal(t, "nsjail", backend.Name())
}

func TestIsAvailable(t *testing.T) {
	t.Run("not available on non-linux", func(t *testing.T) {
		if runtime.GOOS == "linux" {
			t.Skip("skipping on linux")
		}
		backend := New()
		available := backend.IsAvailable(context.Background())
		assert.False(t, available)
	})
}

func TestLoad(t *testing.T) {
	backend, _ := New().(*nsjail)

	t.Run("load without cli context", func(t *testing.T) {
		ctx := context.Background()
		info, err := backend.Load(ctx)

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
	})

	t.Run("load with cli context", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := &cli.Command{}
		cmd.Flags = []cli.Flag{
			&cli.StringFlag{
				Name:  "backend-nsjail-temp-dir",
				Value: tmpDir,
			},
			&cli.BoolFlag{
				Name:  "backend-nsjail-isolated-home",
				Value: true,
			},
		}
		ctx := context.WithValue(context.Background(), types.CliCommand, cmd)

		info, err := backend.Load(ctx)

		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, tmpDir, backend.cfg.tempDir)
		assert.True(t, backend.cfg.isolatedHome)
		assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
	})
}

func TestSetupWorkflow(t *testing.T) {
	t.Run("successful setup", func(t *testing.T) {
		backend, _ := New().(*nsjail)
		backend.cfg.tempDir = t.TempDir()

		ctx := context.Background()
		taskUUID := "test-task-uuid-123"
		config := &types.Config{}

		err := backend.SetupWorkflow(ctx, config, taskUUID)
		require.NoError(t, err)

		// Verify state was saved
		state, err := backend.getWorkflowState(taskUUID)
		require.NoError(t, err)
		assert.NotNil(t, state)
		assert.NotEmpty(t, state.baseDir)
		assert.NotEmpty(t, state.workspaceDir)
		assert.NotEmpty(t, state.homeDir)

		// Verify directories were created
		assert.DirExists(t, state.baseDir)
		assert.DirExists(t, state.workspaceDir)
		assert.DirExists(t, state.homeDir)

		// Verify directory structure
		assert.Equal(t, filepath.Join(state.baseDir, "workspace"), state.workspaceDir)
		assert.Equal(t, filepath.Join(state.baseDir, "home"), state.homeDir)

		// Cleanup
		assert.NoError(t, os.RemoveAll(state.baseDir))
	})

	t.Run("MkdirTemp fails", func(t *testing.T) {
		backend, _ := New().(*nsjail)
		// Use a regular file path as tempDir so MkdirTemp fails
		tmpFile := filepath.Join(t.TempDir(), "not-a-dir")
		require.NoError(t, os.WriteFile(tmpFile, []byte("x"), 0o644))
		backend.cfg.tempDir = tmpFile

		err := backend.SetupWorkflow(context.Background(), &types.Config{}, "fail-uuid")
		require.Error(t, err)

		_, err = backend.getWorkflowState("fail-uuid")
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	})
}

func TestDestroyWorkflow(t *testing.T) {
	backend, _ := New().(*nsjail)
	backend.cfg.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-destroy-task"
	config := &types.Config{}

	// Setup workflow first
	err := backend.SetupWorkflow(ctx, config, taskUUID)
	require.NoError(t, err)

	state, err := backend.getWorkflowState(taskUUID)
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
	_, err = backend.getWorkflowState(taskUUID)
	assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
}

func TestStateManagement(t *testing.T) {
	backend, _ := New().(*nsjail)

	t.Run("save and get state", func(t *testing.T) {
		taskUUID := "test-state-uuid"
		state := &workflowState{
			baseDir:      "/tmp/test",
			homeDir:      "/tmp/test/home",
			workspaceDir: "/tmp/test/workspace",
		}

		backend.workflows.Store(taskUUID, state)

		retrieved, err := backend.getWorkflowState(taskUUID)
		require.NoError(t, err)
		assert.Equal(t, state.baseDir, retrieved.baseDir)
		assert.Equal(t, state.homeDir, retrieved.homeDir)
		assert.Equal(t, state.workspaceDir, retrieved.workspaceDir)
	})

	t.Run("get nonexistent state", func(t *testing.T) {
		_, err := backend.getWorkflowState("nonexistent-uuid")
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	})

	t.Run("delete state", func(t *testing.T) {
		taskUUID := "test-delete-uuid"
		state := &workflowState{}

		backend.workflows.Store(taskUUID, state)

		// Verify state exists
		_, err := backend.getWorkflowState(taskUUID)
		require.NoError(t, err)

		// Delete state
		backend.workflows.Delete(taskUUID)

		// Verify state is gone
		_, err = backend.getWorkflowState(taskUUID)
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	})
}
