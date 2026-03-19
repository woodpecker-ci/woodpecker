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

//go:build test

package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	backend_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

func TestRunNilTracer(t *testing.T) {
	t.Parallel()
	r := New(&backend.Config{}, WithBackend(dummy.New()), WithLogger(newTestLogger(t)))

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tracer must not be nil")
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{{
				Steps: []*backend.Step{{
					Name: "build", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{}, Commands: []string{"echo hello"},
				}},
			}},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	calls := getTracerStates(tracer)
	require.Len(t, calls, 2)
}

func TestRunMultipleStages(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
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
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	calls := getTracerStates(tracer)
	require.Len(t, calls, 4)
}

func TestRunStepError(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
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
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	var exitErr *pipeline_errors.ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 1, exitErr.Code)
}

func TestRunContextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(nil)

	r := New(
		&backend.Config{
			Stages: []*backend.Stage{{
				Steps: []*backend.Step{{
					Name: "s1", UUID: "u1",
					Type: backend.StepTypeCommands, OnSuccess: true,
					Environment: map[string]string{}, Commands: []string{"echo hello"},
				}},
			}},
		},
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithContext(ctx),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
}

func TestRunSetupWorkflowError(t *testing.T) {
	t.Parallel()
	r := New(
		&backend.Config{},
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithTaskUUID(dummy.WorkflowSetupFailUUID),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
}

func TestRunSetupWorkflowInvalidSetupError(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	step := &backend.Step{Name: "clone", UUID: "clone-uuid"}
	setupErr := &pipeline_errors.ErrInvalidWorkflowSetup{
		Err:  errors.New("bad image"),
		Step: step,
	}
	engine := backend_mocks.NewMockBackend(t)
	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(setupErr)
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	r := New(&backend.Config{}, WithBackend(engine), WithTracer(tracer), WithLogger(newTestLogger(t)))

	err := r.Run(t.Context())

	assert.Error(t, err)
	calls := getTracerStates(tracer)
	require.Len(t, calls, 1)
	assert.Equal(t, step, calls[0].Workflow.Step)
	assert.True(t, calls[0].CurrentStep.Exited)
	assert.Equal(t, 1, calls[0].CurrentStep.ExitCode)
}

func TestRunDestroyWorkflowAlwaysCalled(t *testing.T) {
	t.Parallel()
	var destroyed int32
	engine := backend_mocks.NewMockBackend(t)
	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { atomic.AddInt32(&destroyed, 1) }).Return(nil)

	r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))

	_ = r.Run(t.Context())

	assert.Equal(t, int32(1), atomic.LoadInt32(&destroyed))
}

func TestRunDestroyWorkflowCalledOnSetupError(t *testing.T) {
	t.Parallel()
	var destroyed int32
	engine := backend_mocks.NewMockBackend(t)
	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("setup boom"))
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { atomic.AddInt32(&destroyed, 1) }).Return(nil)

	r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))

	_ = r.Run(t.Context())

	assert.Equal(t, int32(1), atomic.LoadInt32(&destroyed))
}

func TestTraceWorkflowSetupError(t *testing.T) {
	t.Parallel()

	t.Run("MatchingError", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer), WithLogger(newTestLogger(t)))
		step := &backend.Step{Name: "setup", UUID: "su"}
		err := &pipeline_errors.ErrInvalidWorkflowSetup{Err: errors.New("bad"), Step: step}

		r.traceWorkflowSetupError(err)

		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.Equal(t, step, calls[0].Workflow.Step)
		assert.True(t, calls[0].CurrentStep.Exited)
		assert.Equal(t, 1, calls[0].CurrentStep.ExitCode)
	})

	t.Run("NonMatchingError", func(t *testing.T) {
		t.Parallel()
		tracer := tracer_mocks.NewMockTracer(t)
		// Trace should NOT be called — no .On() setup means test panics if called.
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer), WithLogger(newTestLogger(t)))

		r.traceWorkflowSetupError(errors.New("generic error"))
	})

	t.Run("TracerFailure", func(t *testing.T) {
		t.Parallel()
		tracer := tracer_mocks.NewMockTracer(t)
		tracer.On("Trace", mock.Anything).Return(errors.New("trace failed"))
		r := New(&backend.Config{}, WithBackend(dummy.New()), WithTracer(tracer), WithLogger(newTestLogger(t)))
		step := &backend.Step{Name: "setup", UUID: "su"}

		// Should not panic — the error is logged, not returned.
		r.traceWorkflowSetupError(&pipeline_errors.ErrInvalidWorkflowSetup{
			Err: errors.New("bad"), Step: step,
		})
	})
}

func TestRunStage(t *testing.T) {
	t.Parallel()

	t.Run("ParallelExecution", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)

		steps := []*backend.Step{
			{Name: "a", UUID: "ua", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo a"}},
			{Name: "b", UUID: "ub", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo b"}},
			{Name: "c", UUID: "uc", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo c"}},
		}

		err := <-r.runStage(t.Context(), steps)

		assert.NoError(t, err)
		assert.Len(t, getTracerStates(tracer), 6)
	})

	t.Run("OneStepFails", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)

		steps := []*backend.Step{
			{Name: "good", UUID: "ug", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{}, Commands: []string{"echo ok"}},
			{Name: "bad", UUID: "ub", Type: backend.StepTypeCommands, OnSuccess: true, Environment: map[string]string{dummy.EnvKeyStepExitCode: "1"}, Commands: []string{"exit 1"}},
		}

		err := <-r.runStage(t.Context(), steps)

		assert.Error(t, err)
	})
}

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
	tracer := newTestTracer(t)
	ctx := context.Background()
	desc := map[string]string{"repo": "test"}

	r := New(&backend.Config{},
		WithBackend(engine),
		WithTracer(tracer),
		WithContext(ctx),
		WithDescription(desc),
		WithTaskUUID("custom-uuid"),
		WithLogger(newTestLogger(t)),
	)

	assert.Equal(t, engine, r.engine)
	assert.Equal(t, tracer, r.tracer)
	assert.Equal(t, ctx, r.ctx)
	assert.Equal(t, "custom-uuid", r.taskUUID)
	assert.Equal(t, "test", r.description["repo"])
}

func TestMakeLoggerWithDescription(t *testing.T) {
	t.Parallel()
	r := New(&backend.Config{},
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
		WithDescription(map[string]string{"repo": "woodpecker", "branch": "main"}),
	)
	r.logStages()
}

func TestGetShutdownCtx(t *testing.T) {
	ctx := GetShutdownCtx()
	assert.NotNil(t, ctx)

	ctx2 := GetShutdownCtx()
	assert.Equal(t, ctx, ctx2)
}

// Gap A: logger == nil guard.
func TestRunNilLogger(t *testing.T) {
	t.Parallel()
	r := New(&backend.Config{},
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		// WithLogger intentionally omitted
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger must not be nil")
}

// Gap B: runnerCtx is already done inside the defer → GetShutdownCtx() fallback.
func TestRunDestroyWorkflowFallsBackToShutdownCtx(t *testing.T) {
	t.Parallel()
	engine := backend_mocks.NewMockBackend(t)
	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	var destroyCtx context.Context
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			destroyCtx, _ = args.Get(0).(context.Context)
		}).Return(nil)

	// Pass a pre-canceled runnerCtx so ctx.Err() != nil in the defer.
	runnerCtx, cancel := context.WithCancelCause(context.Background())
	cancel(nil)

	r := New(&backend.Config{},
		WithBackend(engine),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	_ = r.Run(runnerCtx)

	require.NotNil(t, destroyCtx)
	// The shutdown context is not the canceled runnerCtx — it must still be valid
	// (or at least not the same canceled one).
	assert.NotEqual(t, runnerCtx, destroyCtx,
		"DestroyWorkflow should receive the shutdown fallback context, not the canceled runnerCtx")
}
