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

//go:build linux

package local

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

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

	_, _ = backend.WaitStep(context.Background(), step, taskUUID)

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
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "State:") {
			// e.g. "State:\tZ (zombie)"
			return !strings.Contains(line, "Z")
		}
	}
	return false
}
