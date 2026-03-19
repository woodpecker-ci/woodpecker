//go:build test

// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
)

// ---------------------------------------------------------------------------
// Run — integration tests using dummy backend
// ---------------------------------------------------------------------------

func TestRunNilTracer(t *testing.T) {
	t.Parallel()
	r := New(&backend.Config{}, WithBackend(dummy.New()))
	// tracer is NOT set

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tracer must not be nil")
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{{
				Steps: []*backend.Step{{
					Name: "build", UUID: "u1",
					Type:        backend.StepTypeCommands,
					OnSuccess:   true,
					Environment: map[string]string{},
					Commands:    []string{"echo hello"},
				}},
			}},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	// step-started + step-completed
	calls := tracer.getCalls()
	require.Len(t, calls, 2)
}

func TestRunMultipleStages(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{{
					Name: "stage1", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{}, Commands: []string{"echo 1"},
				}}},
				{Steps: []*backend.Step{{
					Name: "stage2", UUID: "u2",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{}, Commands: []string{"echo 2"},
				}}},
			},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	// 2 stages × (started + completed) = 4 traces
	calls := tracer.getCalls()
	require.Len(t, calls, 4)
}

func TestRunStepError(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{{
				Steps: []*backend.Step{{
					Name: "fail", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{dummy.EnvKeyStepExitCode: "1"},
					Commands:    []string{"exit 1"},
				}},
			}},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	var exitErr *pipeline_errors.ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 1, exitErr.Code)
}

func TestRunStepErrorSkipsSubsequentStages(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{{
					Name: "fail", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{dummy.EnvKeyStepExitCode: "1"},
					Commands:    []string{"exit 1"},
				}}},
				{Steps: []*backend.Step{{
					Name: "after", UUID: "u2",
					Type: backend.StepTypeCommands, OnSuccess: true, OnFailure: false,
					Environment: map[string]string{},
					Commands:    []string{"echo should not run"},
				}}},
			},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	// "after" should be traced as skipped, not started by the engine.
	calls := tracer.getCalls()
	// fail: started + completed = 2, after: skipped = 1 → total 3
	require.Len(t, calls, 3)
	lastCall := calls[2]
	assert.True(t, lastCall.Process.Skipped)
	assert.True(t, lastCall.Process.Exited)
}

func TestRunContextCanceled(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // pre-cancel

	r := New(
		&backend.Config{
			Stages: []*backend.Stage{{
				Steps: []*backend.Step{{
					Name: "s1", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{},
					Commands:    []string{"echo hello"},
				}},
			}},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithContext(ctx),
	)

	err := r.Run(t.Context())

	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
}

func TestRunSetupWorkflowError(t *testing.T) {
	t.Parallel()
	tracer := &mockTracer{}
	r := New(
		&backend.Config{},
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithTaskUUID(dummy.WorkflowSetupFailUUID),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
}

func TestRunSetupWorkflowInvalidSetupError(t *testing.T) {
	t.Parallel()
	// Uses mockEngine: dummy cannot return ErrInvalidWorkflowSetup.
	tracer := &mockTracer{}
	step := &backend.Step{Name: "clone", UUID: "clone-uuid"}
	setupErr := &pipeline_errors.ErrInvalidWorkflowSetup{
		Err:  errors.New("bad image"),
		Step: step,
	}
	engine := &mockEngine{
		setupWorkflowFn: func(_ context.Context, _ *backend.Config, _ string) error {
			return setupErr
		},
	}
	r := New(
		&backend.Config{},
		WithBackend(engine),
		WithTracer(tracer),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	calls := tracer.getCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, step, calls[0].Pipeline.Step)
	assert.True(t, calls[0].Process.Exited)
	assert.Equal(t, 1, calls[0].Process.ExitCode)
}

func TestRunDestroyWorkflowAlwaysCalled(t *testing.T) {
	t.Parallel()
	// Uses mockEngine: need to track DestroyWorkflow calls.
	var destroyed int32
	tracer := &mockTracer{}
	engine := &mockEngine{
		destroyWorkflowFn: func(_ context.Context, _ *backend.Config, _ string) error {
			atomic.AddInt32(&destroyed, 1)
			return nil
		},
	}
	r := New(&backend.Config{}, WithBackend(engine), WithTracer(tracer))

	_ = r.Run(t.Context())

	assert.Equal(t, int32(1), atomic.LoadInt32(&destroyed))
}

func TestRunDestroyWorkflowCalledOnSetupError(t *testing.T) {
	t.Parallel()
	// Uses mockEngine: need to track DestroyWorkflow even when setup fails.
	var destroyed int32
	tracer := &mockTracer{}
	engine := &mockEngine{
		setupWorkflowFn: func(_ context.Context, _ *backend.Config, _ string) error {
			return errors.New("setup boom")
		},
		destroyWorkflowFn: func(_ context.Context, _ *backend.Config, _ string) error {
			atomic.AddInt32(&destroyed, 1)
			return nil
		},
	}
	r := New(&backend.Config{}, WithBackend(engine), WithTracer(tracer))

	_ = r.Run(t.Context())

	assert.Equal(t, int32(1), atomic.LoadInt32(&destroyed))
}

// ---------------------------------------------------------------------------
// traceWorkflowSetupError
// ---------------------------------------------------------------------------

func TestTraceWorkflowSetupError(t *testing.T) {
	t.Parallel()

	t.Run("MatchingError", func(t *testing.T) {
		t.Parallel()
		tracer := &mockTracer{}
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer))
		step := &backend.Step{Name: "setup", UUID: "su"}
		err := &pipeline_errors.ErrInvalidWorkflowSetup{Err: errors.New("bad"), Step: step}

		r.traceWorkflowSetupError(err)

		calls := tracer.getCalls()
		require.Len(t, calls, 1)
		assert.Equal(t, step, calls[0].Pipeline.Step)
		assert.True(t, calls[0].Process.Exited)
		assert.Equal(t, 1, calls[0].Process.ExitCode)
	})

	t.Run("NonMatchingError", func(t *testing.T) {
		t.Parallel()
		tracer := &mockTracer{}
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer))

		r.traceWorkflowSetupError(errors.New("generic error"))

		assert.Empty(t, tracer.getCalls())
	})

	t.Run("TracerFailure", func(t *testing.T) {
		t.Parallel()
		tracer := &mockTracer{fn: func(_ *state.State) error {
			return errors.New("trace failed")
		}}
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer))
		step := &backend.Step{Name: "setup", UUID: "su"}
		err := &pipeline_errors.ErrInvalidWorkflowSetup{Err: errors.New("bad"), Step: step}

		// Should not panic.
		r.traceWorkflowSetupError(err)
	})
}

// ---------------------------------------------------------------------------
// runStage
// ---------------------------------------------------------------------------

func TestRunStage(t *testing.T) {
	t.Parallel()

	t.Run("ParallelExecution", func(t *testing.T) {
		t.Parallel()
		tracer := &mockTracer{}
		r := newDummyRuntime(t, tracer)

		steps := []*backend.Step{
			{Name: "a", UUID: "ua", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo a"}},
			{Name: "b", UUID: "ub", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo b"}},
			{Name: "c", UUID: "uc", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo c"}},
		}

		err := <-r.runStage(t.Context(), steps)

		assert.NoError(t, err)
		// 3 steps × (started + completed) = 6 traces
		assert.Len(t, tracer.getCalls(), 6)
	})

	t.Run("OneStepFails", func(t *testing.T) {
		t.Parallel()
		tracer := &mockTracer{}
		r := newDummyRuntime(t, tracer)

		steps := []*backend.Step{
			{Name: "good", UUID: "ug", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo ok"}},
			{Name: "bad", UUID: "ub", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{dummy.EnvKeyStepExitCode: "1"}, Commands: []string{"exit 1"}},
		}

		err := <-r.runStage(t.Context(), steps)

		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// New + Options
// ---------------------------------------------------------------------------

func TestNewDefaults(t *testing.T) {
	t.Parallel()
	spec := &backend.Config{}

	r := New(spec)

	assert.Equal(t, spec, r.spec)
	assert.NotEmpty(t, r.taskUUID)
	assert.NotNil(t, r.ctx)
	assert.Nil(t, r.tracer)
	assert.Nil(t, r.engine)
	assert.NoError(t, r.err.Get())
}

func TestWithOptions(t *testing.T) {
	t.Parallel()
	engine := dummy.New()
	tracer := &mockTracer{}
	ctx := context.Background()
	desc := map[string]string{"repo": "test"}

	r := New(&backend.Config{},
		WithBackend(engine),
		WithTracer(tracer),
		WithContext(ctx),
		WithDescription(desc),
		WithTaskUUID("custom-uuid"),
	)

	assert.Equal(t, engine, r.engine)
	assert.Equal(t, tracer, r.tracer)
	assert.Equal(t, ctx, r.ctx)
	assert.Equal(t, "custom-uuid", r.taskUUID)
	assert.Equal(t, "test", r.description["repo"])
}

// ---------------------------------------------------------------------------
// GetShutdownCtx
// ---------------------------------------------------------------------------

func TestGetShutdownCtx(t *testing.T) {
	ctx := GetShutdownCtx()
	assert.NotNil(t, ctx)

	ctx2 := GetShutdownCtx()
	assert.Equal(t, ctx, ctx2)
}
