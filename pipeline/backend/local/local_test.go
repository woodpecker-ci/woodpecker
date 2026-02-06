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

package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestIsAvailable(t *testing.T) {
	t.Run("not available in container", func(t *testing.T) {
		backend := New()

		t.Setenv("WOODPECKER_IN_CONTAINER", "true")

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
	backend, _ := New().(*local)

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
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()

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
}

func TestDestroyWorkflow(t *testing.T) {
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()

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

func prepairEnv(t *testing.T) {
	prevEnv := os.Environ()
	os.Clearenv()
	t.Cleanup(func() {
		for i := range prevEnv {
			env := strings.SplitN(prevEnv[i], "=", 2)
			//nolint:usetesting // reason: the suggested t.Setenv will be undone on t.Run() end witch we explizite dont want here
			_ = os.Setenv(env[0], env[1])
		}
	})
}

func TestRunStep(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping on non linux due to shell availability and symlink capability")
	}

	// we lookup shell tools we use first and create the PATH var based on that
	shBinary, err := exec.LookPath("sh")
	require.NoError(t, err)
	path := []string{filepath.Dir(shBinary)}
	echoBinary, err := exec.LookPath("echo")
	require.NoError(t, err)
	if echoPath := filepath.Dir(echoBinary); !slices.Contains(path, echoPath) {
		path = append(path, echoPath)
	}
	// we make a symlinc to have a posix but non default shell
	altShellDir := t.TempDir()
	altShellPath := filepath.Join(altShellDir, "altsh")
	require.NoError(t, os.Symlink(shBinary, altShellPath))
	path = append(path, altShellDir)

	prepairEnv(t)
	//nolint:usetesting // reason: we use prepairEnv()
	os.Setenv("PATH", strings.Join(path, ":"))

	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()
	ctx := t.Context()
	taskUUID := "test-run-tasks"

	// Setup workflow
	require.NoError(t, backend.SetupWorkflow(ctx, &types.Config{}, taskUUID))

	t.Run("type commands", func(t *testing.T) {
		step := &types.Step{
			UUID:     "step-1",
			Name:     "test-step",
			Type:     types.StepTypeCommands,
			Image:    "sh",
			Commands: []string{"echo hello", "env"},
			Environment: map[string]string{
				"TEST_VAR": "test_value",
			},
		}

		t.Run("start successful", func(t *testing.T) {
			err = backend.StartStep(ctx, step, taskUUID)
			require.NoError(t, err)

			// Verify command was started
			state, err := backend.getWorkflowState(taskUUID)
			require.NoError(t, err)
			stepStateWraped, contains := state.stepState.Load(step.UUID)
			assert.True(t, contains)
			stepState, _ := stepStateWraped.(*stepState)
			assert.NotNil(t, stepState.cmd)

			var outputData []byte
			outputDataMutex := sync.Mutex{}
			go t.Run("TailStep", func(t *testing.T) {
				outputDataMutex.Lock()
				go outputDataMutex.Unlock()
				output, err := backend.TailStep(ctx, step, taskUUID)
				require.NoError(t, err)
				assert.NotNil(t, output)

				// Read output
				outputData, err = io.ReadAll(output)
				require.NoError(t, err)
			})

			// Wait for step to finish
			t.Run("TestWaitStep", func(t *testing.T) {
				time.Sleep(time.Second / 5) // needed to prevent race condition on outputData
				state, err := backend.WaitStep(ctx, step, taskUUID)
				require.NoError(t, err)
				assert.True(t, state.Exited)
				assert.Equal(t, 0, state.ExitCode)
			})

			// Verify output
			outputDataMutex.Lock()
			go outputDataMutex.Unlock()
			outputLines := strings.Split(strings.TrimSpace(string(outputData)), "\n")
			require.Truef(t, len(outputLines) > 3, "output of lines must be bigger than 3 at least but we got: %#v", outputLines)
			// we first test output without environments
			wantBeforeEnvs := []string{
				"+ echo hello",
				"hello",
				"+ env",
			}
			gotBeforeEnvs := outputLines[:len(wantBeforeEnvs)]
			assert.Equal(t, wantBeforeEnvs, gotBeforeEnvs)
			// we filter out nixos specific stuff catched up in env output
			gotEnvs := slices.DeleteFunc(outputLines[len(wantBeforeEnvs):], func(s string) bool {
				return strings.HasPrefix(s, "_=") || strings.HasPrefix(s, "SHLVL=")
			})
			assert.ElementsMatch(t, []string{
				"PWD=" + state.baseDir + "/workspace",
				"USERPROFILE=" + state.baseDir + "/home",
				"TEST_VAR=test_value",
				"HOME=" + state.baseDir + "/home",
				"CI_WORKSPACE=" + state.baseDir + "/workspace",
				"PATH=" + strings.Join(path, ":"),
			}, gotEnvs)

			t.Run("TestDestroyStep", func(t *testing.T) {
				err := backend.DestroyStep(ctx, step, taskUUID)
				require.NoError(t, err)
			})
		})
	})

	t.Run("run command in alternate unix shell", func(t *testing.T) {
		step := &types.Step{
			UUID:     "step-altshell",
			Name:     "altshell",
			Type:     types.StepTypeCommands,
			Image:    "altsh",
			Commands: []string{"echo success"},
		}

		err = backend.StartStep(ctx, step, taskUUID)
		require.NoError(t, err)

		state, err := backend.WaitStep(ctx, step, taskUUID)
		require.NoError(t, err)
		assert.True(t, state.Exited)
		assert.Equal(t, 0, state.ExitCode)
	})

	t.Run("command should fail", func(t *testing.T) {
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

	t.Run("WaitStep", func(t *testing.T) {
		t.Run("step not found", func(t *testing.T) {
			step := &types.Step{
				UUID: "nonexistent-step",
				Name: "missing",
			}

			_, err = backend.WaitStep(ctx, step, taskUUID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		})
	})

	t.Run("type plugin", func(t *testing.T) {
		step := &types.Step{
			UUID:        "step-plugin-1",
			Name:        "test-plugin",
			Type:        types.StepTypePlugin,
			Image:       "echo", // Use a binary that exists
			Environment: map[string]string{},
		}

		t.Run("start", func(t *testing.T) {
			err = backend.StartStep(ctx, step, taskUUID)
			require.NoError(t, err)

			// Verify command was started
			state, err := backend.getStepState(taskUUID, step.UUID)
			require.NoError(t, err)
			assert.NotEqualf(t, 0, state.cmd.Process.Pid, "expect an pid of the process")
		})
	})

	t.Run("type unsupported", func(t *testing.T) {
		step := &types.Step{
			UUID: "step-unsupported",
			Name: "test-unsupported",
			Type: "unsupported-type",
		}

		t.Run("start", func(t *testing.T) {
			err = backend.StartStep(ctx, step, taskUUID)
			assert.ErrorIs(t, err, ErrUnsupportedStepType)
		})
	})

	// Cleanup
	assert.NoError(t, backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID))
}

func TestStateManagement(t *testing.T) {
	backend, _ := New().(*local)

	t.Run("save and get state", func(t *testing.T) {
		taskUUID := "test-state-uuid"
		state := &workflowState{
			baseDir:      "/tmp/test",
			homeDir:      "/tmp/test/2home",
			workspaceDir: "/tmp/test/2workspace",
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

func TestConcurrentWorkflows(t *testing.T) {
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()

	// Create multiple workflows concurrently
	taskUUIDs := []string{"task-1", "task-2", "task-3"}

	for _, uuid := range taskUUIDs {
		err := backend.SetupWorkflow(ctx, &types.Config{}, uuid)
		require.NoError(t, err)
	}

	counter := atomic.Int32{}
	counter.Store(0)
	for _, uuid := range taskUUIDs {
		go t.Run("start step in "+uuid, func(t *testing.T) {
			for i := 0; i < 3; i++ {
				counter.Store(counter.Load() + 1)
				step := &types.Step{
					UUID:        fmt.Sprintf("step-%s-%d", uuid, i),
					Name:        fmt.Sprintf("step-name-%s-%d", uuid, i),
					Type:        types.StepTypePlugin,
					Image:       "sh",
					Commands:    []string{fmt.Sprintf("echo %s %d", uuid, i)},
					Environment: map[string]string{},
				}
				require.NoError(t, backend.StartStep(ctx, step, uuid))
				_, err := backend.WaitStep(ctx, step, uuid)
				require.NoError(t, err)
				counter.Store(counter.Load() - 1)
			}
		})
	}

	// Verify all states exist
	for _, uuid := range taskUUIDs {
		state, err := backend.getWorkflowState(uuid)
		require.NoError(t, err)
		assert.NotNil(t, state)
	}

	failSave := 0
loop:
	for {
		if failSave == 1000 { // wait max 1s
			t.Log("failSave was hit")
			t.FailNow()
		}
		failSave++
		select {
		case <-time.After(time.Millisecond):
			if count := counter.Load(); count == 0 {
				break loop
			} else {
				t.Logf("count at: %d", count)
			}
		case <-ctx.Done():
			return
		}
	}

	// Cleanup all workflows
	for _, uuid := range taskUUIDs {
		// Cleanup all steps
		for i := 0; i < 3; i++ {
			stepUUID := fmt.Sprintf("step-%s-%d", uuid, i)
			assert.NoError(t, backend.DestroyStep(ctx, &types.Step{UUID: stepUUID}, uuid))
		}

		// finish with workflow cleanup
		err := backend.DestroyWorkflow(ctx, &types.Config{}, uuid)
		require.NoError(t, err)
	}

	// Verify all states are deleted
	for _, uuid := range taskUUIDs {
		_, err := backend.getWorkflowState(uuid)
		assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
	}
}
