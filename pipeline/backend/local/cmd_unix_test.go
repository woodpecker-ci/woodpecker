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

//go:build !windows

package local

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// startCanceledStep stores a step state whose command was created but never
// started. The local backend stores the step state before it calls cmd.Start(),
// and Start() returns early without setting cmd.Process if the context is
// already canceled. This is what happens when a workflow gets canceled (e.g. by
// a new push) after the agent picked it up but before the step was started.
func startCanceledStep(t *testing.T, backend *local, taskUUID string) *types.Step {
	t.Helper()

	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(nil) // pre-cancel

	step := &types.Step{
		UUID:     "step-canceled-before-start",
		Name:     "canceled-before-start",
		Type:     types.StepTypeCommands,
		Image:    "sh",
		Commands: []string{"echo never executed"},
	}

	require.Error(t, backend.StartStep(ctx, step, taskUUID), "step must not start on a canceled context")

	state, err := backend.getStepState(taskUUID, step.UUID)
	require.NoError(t, err)
	require.NotNil(t, state.cmd, "step command must be stored")
	require.Nil(t, state.cmd.Process, "step command must not have been started")

	return step
}

// TestDestroyWorkflowOfNotStartedStep ensures the workflow cleanup does not
// panic if a step command was created but never started. DestroyWorkflow calls
// the cancel hook of every stored step command, and the hook set by newCmd
// dereferences cmd.Process, which is nil for a command that never started.
//
// Regression test for: agent panic on canceled workflow using the local backend.
func TestDestroyWorkflowOfNotStartedStep(t *testing.T) {
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-destroy-workflow-not-started"
	require.NoError(t, backend.SetupWorkflow(ctx, &types.Config{}, taskUUID))

	startCanceledStep(t, backend, taskUUID)

	state, err := backend.getWorkflowState(taskUUID)
	require.NoError(t, err)

	assert.NoError(t, backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID))
	assert.NoDirExists(t, state.baseDir, "workflow directory must be cleaned up")
}

// TestDestroyStepOfNotStartedStep ensures the step cleanup does not panic if the
// step command was created but never started.
//
// Regression test for: agent panic on canceled workflow using the local backend.
func TestDestroyStepOfNotStartedStep(t *testing.T) {
	backend, _ := New().(*local)
	backend.tempDir = t.TempDir()

	ctx := context.Background()
	taskUUID := "test-destroy-step-not-started"
	require.NoError(t, backend.SetupWorkflow(ctx, &types.Config{}, taskUUID))
	t.Cleanup(func() {
		_ = backend.DestroyWorkflow(ctx, &types.Config{}, taskUUID)
	})

	step := startCanceledStep(t, backend, taskUUID)

	assert.NoError(t, backend.DestroyStep(ctx, step, taskUUID))
}
