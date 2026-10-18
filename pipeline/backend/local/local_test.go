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
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
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

		available := backend.IsAvailable(t.Context())
		assert.False(t, available)
	})

	t.Run("available without container env and no cli context", func(t *testing.T) {
		backend := New()

		os.Unsetenv("WOODPECKER_IN_CONTAINER")
		available := backend.IsAvailable(t.Context())
		assert.True(t, available)
	})
}

func TestLoad(t *testing.T) {
	backend, _ := New().(*local)

	t.Run("load without cli context", func(t *testing.T) {
		ctx := t.Context()
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
		ctx := context.WithValue(t.Context(), types.CliCommand, cmd)

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

	ctx := t.Context()
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

	ctx := t.Context()
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
	backend.isolatedHome = true
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

			outputDataCh := make(chan []byte)
			go func() {
				output, err := backend.TailStep(ctx, step, taskUUID)
				require.NoError(t, err)
				assert.NotNil(t, output)

				// Read output
				data, err := io.ReadAll(output)
				require.NoError(t, err)
				outputDataCh <- data
			}()

			// Wait for step to finish
			t.Run("TestWaitStep", func(t *testing.T) {
				time.Sleep(time.Second / 5) // needed to prevent race condition on outputData
				state, err := backend.WaitStep(ctx, step, taskUUID)
				require.NoError(t, err)
				assert.True(t, state.Exited)
				assert.Equal(t, 0, state.ExitCode)
			})

			// Verify output
			var outputData []byte
			select {
			case outputData = <-outputDataCh:
			case <-t.Context().Done():
				t.Fail()
				return
			}
			outputLines := strings.Split(strings.TrimSpace(string(outputData)), "\n")
			require.Truef(t, len(outputLines) > 3, "output of lines must be bigger than 3 at least but we got: %#v", outputLines)
			// we first test output without environments
			wantBeforeEnvs := []string{
				"▶  echo hello",
				"hello",
				"▶  env",
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

	ctx := t.Context()

	// Create multiple workflows concurrently
	taskUUIDs := []string{"task-1", "task-2", "task-3"}

	for _, uuid := range taskUUIDs {
		err := backend.SetupWorkflow(ctx, &types.Config{}, uuid)
		require.NoError(t, err)
	}

	counter := atomic.Int32{}
	counter.Store(0)
	t.Parallel()
	for _, uuid := range taskUUIDs {
		t.Run("start step in "+uuid, func(t *testing.T) {
			for i := range 3 {
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
		for i := range 3 {
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

// resolveTestPath looks up every binary in the current PATH and returns a PATH
// value composed of their (unique) directories. Used to rebuild PATH after
// prepairEnv() clears the environment.
func resolveTestPath(t *testing.T, bins ...string) string {
	t.Helper()
	var dirs []string
	for _, bin := range bins {
		p, err := exec.LookPath(bin)
		require.NoErrorf(t, err, "lookup %q", bin)
		d := filepath.Dir(p)
		if !slices.Contains(dirs, d) {
			dirs = append(dirs, d)
		}
	}
	return strings.Join(dirs, ":")
}

// TestStepInOwnProcessGroup ensures a step's shell is spawned in its own
// process group, isolating it from the agent (the test process). Without this
// isolation, signals the step sends to its own process group (e.g. `make -j`
// cleaning up failed parallel jobs) would also reach the agent.
//
// Regression test for: local backend signal propagation to agent.
func TestStepInOwnProcessGroup(t *testing.T) {
	path := resolveTestPath(t, "sh", "sleep")
	prepairEnv(t)
	//nolint:usetesting // see prepairEnv
	os.Setenv("PATH", path)

	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()
	ctx := t.Context()

	taskUUID := "test-pgrp-isolation"
	require.NoError(t, backend.SetupWorkflow(ctx, &types.Config{}, taskUUID))
	t.Cleanup(func() {
		_ = backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
	})

	step := &types.Step{
		UUID:     "step-pgrp",
		Name:     "pgrp",
		Type:     types.StepTypeCommands,
		Image:    "sh",
		Commands: []string{"sleep 5"},
	}

	require.NoError(t, backend.StartStep(ctx, step, taskUUID))

	stepState, err := backend.getStepState(taskUUID, step.UUID)
	require.NoError(t, err)
	require.NotNil(t, stepState.cmd)
	require.NotNil(t, stepState.cmd.Process)

	childPID := stepState.cmd.Process.Pid
	childPgid, err := syscall.Getpgid(childPID)
	require.NoError(t, err)
	agentPgid, err := syscall.Getpgid(os.Getpid())
	require.NoError(t, err)

	// The child must NOT share the agent's process group, otherwise signals
	// the child sends to its own group (e.g. via `make` or `kill -- -$$`) hit
	// the agent too.
	assert.NotEqualf(t, agentPgid, childPgid,
		"step shell shares process group with agent (pgid=%d); signals from the step would reach the agent",
		agentPgid)

	// The child should be the leader of its own group (pgid == pid).
	assert.Equalf(t, childPID, childPgid,
		"step shell is not the leader of its own process group (pid=%d, pgid=%d)",
		childPID, childPgid)

	require.NoError(t, backend.DestroyStep(ctx, step, taskUUID))
}

// TestStepCancelKillsGrandchildren ensures that canceling a step also kills
// processes spawned by the step's shell. Default exec.CommandContext only
// signals the direct child; without a group-aware cancel hook the
// grandchildren (e.g. `make`, `nix`, `cc1`) become orphans and keep running.
//
// Regression test for: orphan grandchildren after step cancel.
func TestStepCancelKillsGrandchildren(t *testing.T) {
	path := resolveTestPath(t, "sh", "sleep")
	prepairEnv(t)
	//nolint:usetesting // see prepairEnv
	os.Setenv("PATH", path)

	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)

	taskUUID := "test-cancel-grandchild"
	require.NoError(t, backend.SetupWorkflow(ctx, &types.Config{}, taskUUID))
	t.Cleanup(func() {
		_ = backend.DestroyWorkflow(context.Background(), &types.Config{}, taskUUID)
	})

	pidFile := filepath.Join(t.TempDir(), "grandchild.pid")

	step := &types.Step{
		UUID:  "step-grandchild",
		Name:  "grandchild",
		Type:  types.StepTypeCommands,
		Image: "sh",
		Commands: []string{
			// Background `sleep` is the "grandchild". Write its PID, then
			// `wait` so the shell stays alive until the context is canceled.
			"sleep 30 & echo $! > " + pidFile + "; wait",
		},
	}

	require.NoError(t, backend.StartStep(ctx, step, taskUUID))

	// Wait for the grandchild to record its PID.
	var grandchildPID int
	require.Eventually(t, func() bool {
		data, err := os.ReadFile(pidFile)
		if err != nil {
			return false
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil || pid <= 0 {
			return false
		}
		grandchildPID = pid
		return true
	}, 3*time.Second, 20*time.Millisecond, "grandchild never wrote its pid")

	// Cancel the context — this should fire the step's cancel hook and kill
	// the entire process group, taking the grandchild with it.
	cancel(nil)

	_, _ = backend.WaitStep(t.Context(), step, taskUUID)

	require.Eventuallyf(t, func() bool {
		return !pidAlive(grandchildPID)
	}, 3*time.Second, 50*time.Millisecond,
		"grandchild pid %d is still alive after step cancel; cancel did not propagate to the process group",
		grandchildPID)
}

// The pidAlive reports whether pid still maps to a non-zombie process,
// kill(pid, 0) succeeds for zombies too, which would give false positives,
// so /proc/<pid>/status is the more reliable signal on Linux.
func pidAlive(pid int) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return false
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "State:") {
			// e.g. "State:\tZ (zombie)"
			return !strings.Contains(line, "Z")
		}
	}
	return false
}
