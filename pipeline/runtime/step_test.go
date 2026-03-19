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
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	backend_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

const testWorkflowID = "WID_test"

// newDummyRuntime creates a Runtime backed by the dummy engine with a pre-setup
// workflow so individual step methods can be tested in isolation.
func newDummyRuntime(t *testing.T, tracer *tracer_mocks.MockTracer) *Runtime {
	t.Helper()
	engine := dummy.New()
	r := New(
		&backend.Config{},
		WithBackend(engine),
		WithTracer(tracer),
		WithTaskUUID(testWorkflowID),
	)
	require.NoError(t, engine.SetupWorkflow(t.Context(), nil, testWorkflowID))
	return r
}

// newTestTracer creates a MockTracer that accepts any number of Trace calls.
func newTestTracer(t *testing.T) *tracer_mocks.MockTracer {
	t.Helper()
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Return(nil).Maybe()
	return tracer
}

func dummyStep(name string) *backend.Step {
	return &backend.Step{
		Name:        name,
		UUID:        name + "-uuid",
		Type:        backend.StepTypeCommands,
		OnSuccess:   true,
		OnFailure:   false,
		Environment: map[string]string{},
		Commands:    []string{"echo hello"},
	}
}

// ---------------------------------------------------------------------------
// shouldSkipStep
// ---------------------------------------------------------------------------

func TestShouldSkipStep(t *testing.T) {
	t.Parallel()

	t.Run("NoErrorOnSuccessTrue", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := &backend.Step{Name: "s", OnSuccess: true, OnFailure: false}

		assert.False(t, r.shouldSkipStep(step))
	})

	t.Run("NoErrorOnSuccessFalse", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := &backend.Step{Name: "s", OnSuccess: false, OnFailure: true}

		assert.True(t, r.shouldSkipStep(step))
	})

	t.Run("ErrorOnFailureTrue", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		r.err.Set(errors.New("previous failure"))
		step := &backend.Step{Name: "s", OnSuccess: false, OnFailure: true}

		assert.False(t, r.shouldSkipStep(step))
	})

	t.Run("ErrorOnFailureFalse", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		r.err.Set(errors.New("previous failure"))
		step := &backend.Step{Name: "s", OnSuccess: true, OnFailure: false}

		assert.True(t, r.shouldSkipStep(step))
	})
}

// ---------------------------------------------------------------------------
// traceStep
// ---------------------------------------------------------------------------

func TestTraceStep(t *testing.T) {
	t.Parallel()

	t.Run("StepStarted", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		r.started = 1000
		step := dummyStep("s1")

		err := r.traceStep(nil, nil, step)

		assert.NoError(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.Equal(t, int64(1000), calls[0].Pipeline.Started)
		assert.Equal(t, step, calls[0].Pipeline.Step)
		assert.False(t, calls[0].Process.Exited)
	})

	t.Run("StepFailedToStart", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("s1")
		startErr := errors.New("image pull failed")

		err := r.traceStep(nil, startErr, step)

		assert.ErrorIs(t, err, startErr)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.True(t, calls[0].Process.Exited)
		assert.Equal(t, startErr, calls[0].Process.Error)
	})

	t.Run("StepFinished", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("s1")
		ps := &backend.State{Exited: true, ExitCode: 0, Started: 42}

		err := r.traceStep(ps, nil, step)

		assert.NoError(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.True(t, calls[0].Process.Exited)
		assert.Equal(t, 0, calls[0].Process.ExitCode)
		assert.Equal(t, int64(42), calls[0].Process.Started)
	})

	t.Run("StepSkipped", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("s1")
		ps := &backend.State{Exited: true, Skipped: true}

		err := r.traceStep(ps, nil, step)

		assert.NoError(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.True(t, calls[0].Process.Skipped)
		assert.True(t, calls[0].Process.Exited)
	})

	t.Run("TracerError", func(t *testing.T) {
		t.Parallel()
		traceErr := errors.New("tracer unavailable")
		tracer := tracer_mocks.NewMockTracer(t)
		tracer.On("Trace", mock.Anything).Return(traceErr).Maybe()
		r := newDummyRuntime(t, tracer)

		err := r.traceStep(nil, nil, dummyStep("s1"))

		assert.ErrorIs(t, err, traceErr)
	})

	t.Run("PipelineErrorPropagated", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		r.err.Set(errors.New("earlier failure"))

		_ = r.traceStep(nil, nil, dummyStep("s1"))

		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.EqualError(t, calls[0].Pipeline.Error, "earlier failure")
	})
}

// ---------------------------------------------------------------------------
// startStep — dummy for success + start/tail failures; mockery mock for logger test
// ---------------------------------------------------------------------------

func TestStartStep(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := dummyStep("s1")

		waitForLogs, startTime, err := r.startStep(step)

		assert.NoError(t, err)
		assert.NotNil(t, waitForLogs)
		assert.Greater(t, startTime, int64(0))
		waitForLogs()
	})

	t.Run("StartStepError", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := dummyStep("fail")
		step.Environment[dummy.EnvKeyStepStartFail] = "true"

		_, _, err := r.startStep(step)

		assert.Error(t, err)
	})

	t.Run("TailStepError", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := dummyStep("tail-fail")
		step.Environment[dummy.EnvKeyStepTailFail] = "true"
		r.logger = logging.Logger(func(_ *backend.Step, _ io.ReadCloser) error { return nil })

		_, _, err := r.startStep(step)

		assert.Error(t, err)
	})

	t.Run("WithLogger", func(t *testing.T) {
		t.Parallel()
		var logCalled int32
		engine := backend_mocks.NewMockBackend(t)
		engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		engine.On("TailStep", mock.Anything, mock.Anything, mock.Anything).
			Return(io.NopCloser(strings.NewReader("log line")), nil)

		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)),
			WithLogger(logging.Logger(func(_ *backend.Step, rc io.ReadCloser) error {
				atomic.AddInt32(&logCalled, 1)
				_, _ = io.ReadAll(rc)
				return nil
			})))
		step := dummyStep("s1")

		waitForLogs, _, err := r.startStep(step)
		require.NoError(t, err)

		waitForLogs()
		assert.Equal(t, int32(1), atomic.LoadInt32(&logCalled))
	})
}

// ---------------------------------------------------------------------------
// completeStep — uses mockery mock for fine-grained control over
// WaitStep/DestroyStep return values that dummy cannot provide.
// ---------------------------------------------------------------------------

func TestCompleteStep(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 0}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.NoError(t, err)
		assert.True(t, ws.Exited)
		assert.Equal(t, 0, ws.ExitCode)
	})

	t.Run("NonZeroExitCode", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 1}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		var exitErr *pipeline_errors.ExitError
		assert.True(t, errors.As(err, &exitErr))
		assert.Equal(t, 1, exitErr.Code)
		assert.Equal(t, 1, ws.ExitCode)
	})

	t.Run("OOMKilled", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, OOMKilled: true, ExitCode: 137}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		var oomErr *pipeline_errors.OomError
		assert.True(t, errors.As(err, &oomErr))
		assert.True(t, ws.OOMKilled)
	})

	t.Run("ContextCanceledNilState", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, context.Canceled)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.NoError(t, err)
		require.NotNil(t, ws, "nil guard must allocate a new State")
		assert.Equal(t, pipeline_errors.ErrCancel, ws.Error)
	})

	t.Run("ContextCanceledWithState", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 0}, context.Canceled)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.NoError(t, err)
		assert.Equal(t, pipeline_errors.ErrCancel, ws.Error)
	})

	t.Run("WaitStepNonCancelError", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("engine exploded"))
		// DestroyStep should NOT be called — early return.
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.EqualError(t, err, "engine exploded")
		assert.Nil(t, ws)
	})

	t.Run("DestroyStepError", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 0}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("cleanup failed"))
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.EqualError(t, err, "cleanup failed")
		assert.Nil(t, ws)
	})

	t.Run("SetsStartTime", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 0}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(newTestTracer(t)))

		ws, err := r.completeStep(context.Background(), dummyStep("s1"), func() {}, 9999)

		assert.NoError(t, err)
		assert.Equal(t, int64(9999), ws.Started)
	})
}

// ---------------------------------------------------------------------------
// executeStep — uses dummy for the full step lifecycle
// ---------------------------------------------------------------------------

func TestExecuteStep(t *testing.T) {
	t.Parallel()

	t.Run("SkippedStepTraced", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := &backend.Step{
			Name: "skip-me", UUID: "skip-uuid",
			Type: backend.StepTypeCommands, Environment: map[string]string{},
			OnSuccess: false, OnFailure: true,
		}

		err := r.executeStep(t.Context(), step)

		assert.NoError(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.True(t, calls[0].Process.Skipped)
		assert.True(t, calls[0].Process.Exited)
	})

	t.Run("BlockingStepSuccess", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("build")

		err := r.executeStep(t.Context(), step)

		assert.NoError(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 2)
		assert.False(t, calls[0].Process.Exited, "first trace should be step-started")
		assert.True(t, calls[1].Process.Exited, "second trace should be step-completed")
	})

	t.Run("BlockingStepFailure", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("fail")
		step.Environment[dummy.EnvKeyStepExitCode] = "1"

		err := r.executeStep(t.Context(), step)

		assert.Error(t, err)
		var exitErr *pipeline_errors.ExitError
		assert.True(t, errors.As(err, &exitErr))
		assert.Equal(t, 1, exitErr.Code)
	})

	t.Run("DetachedStep", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("svc")
		step.Detached = true
		step.Type = backend.StepTypeService
		step.Environment[dummy.EnvKeyStepSleep] = "1ms"

		err := r.executeStep(t.Context(), step)

		assert.NoError(t, err)
		assert.Eventually(t, func() bool {
			return len(getTracerStates(tracer)) >= 2
		}, time.Second, 10*time.Millisecond)
	})
}

// ---------------------------------------------------------------------------
// runBlockingStep
// ---------------------------------------------------------------------------

func TestRunBlockingStep(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))

		err := r.runBlockingStep(t.Context(), dummyStep("s1"))

		assert.NoError(t, err)
	})

	t.Run("FailureIgnore", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := dummyStep("s1")
		step.Failure = metadata.FailureIgnore
		step.Environment[dummy.EnvKeyStepExitCode] = "1"

		err := r.runBlockingStep(t.Context(), step)

		assert.NoError(t, err, "error should be suppressed when Failure==FailureIgnore")
	})

	t.Run("StartFailure", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("s1")
		step.Environment[dummy.EnvKeyStepStartFail] = "true"

		err := r.runBlockingStep(t.Context(), step)

		assert.Error(t, err)
		calls := getTracerStates(tracer)
		require.Len(t, calls, 1)
		assert.True(t, calls[0].Process.Exited)
	})

	t.Run("DestroyStepErrorMappedToErrCancel", func(t *testing.T) {
		t.Parallel()
		engine := backend_mocks.NewMockBackend(t)
		engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend.State{Exited: true, ExitCode: 0}, nil)
		engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).
			Return(context.Canceled)

		tracer := newTestTracer(t)
		r := New(&backend.Config{}, WithBackend(engine), WithTracer(tracer))

		err := r.runBlockingStep(context.Background(), dummyStep("s1"))

		assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
	})
}

// ---------------------------------------------------------------------------
// runDetachedStep
// ---------------------------------------------------------------------------

func TestRunDetachedStep(t *testing.T) {
	t.Parallel()

	t.Run("ReturnsImmediately", func(t *testing.T) {
		t.Parallel()
		tracer := newTestTracer(t)
		r := newDummyRuntime(t, tracer)
		step := dummyStep("svc")
		step.Environment[dummy.EnvKeyStepSleep] = "1ms"

		err := r.runDetachedStep(t.Context(), step)

		assert.NoError(t, err)
		assert.Eventually(t, func() bool {
			return len(getTracerStates(tracer)) >= 1
		}, time.Second, 10*time.Millisecond)
	})

	t.Run("StartFailure", func(t *testing.T) {
		t.Parallel()
		r := newDummyRuntime(t, newTestTracer(t))
		step := dummyStep("svc")
		step.Environment[dummy.EnvKeyStepStartFail] = "true"

		err := r.runDetachedStep(t.Context(), step)

		assert.Error(t, err)
	})
}
