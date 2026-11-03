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
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

var skipOn78 = map[int]backend_types.StepSignal{78: backend_types.StepSignalSkip}

func signalTestStep(name string, env map[string]string) *backend_types.Step {
	if env == nil {
		env = map[string]string{}
	}
	return &backend_types.Step{
		Name: name, UUID: name + "-uuid",
		Type: backend_types.StepTypeCommands, OnSuccess: true,
		Environment: env, Commands: []string{"echo " + name},
	}
}

// finalStepStates returns the last traced state per step name.
func finalStepStates(tracer *tracer_mocks.MockTracer) map[string]backend_types.State {
	result := make(map[string]backend_types.State)
	for _, s := range getTracerStates(tracer) {
		if s.CurrStep != nil {
			result[s.CurrStep.Name] = s.CurrStepState
		}
	}
	return result
}

func hasTraceForStep(traces []state.State, name string) bool {
	return indexOfTrace(traces, func(s state.State) bool {
		return s.CurrStep != nil && s.CurrStep.Name == name
	}) >= 0
}

func TestSkipSignalSkipsRemainingStages(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	check.ExitCodeSignals = skipOn78
	build := signalTestStep("build", nil)
	cleanup := signalTestStep("cleanup", nil)
	cleanup.OnFailure = true // runs on success and failure, still skipped

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{check}},
			{Steps: []*backend_types.Step{build}},
			{Steps: []*backend_types.Step{cleanup}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))
	assert.NoError(t, r.Err(), "workflow must stay successful")

	states := finalStepStates(tracer)
	assert.Equal(t, 0, states["check"].ExitCode)
	assert.False(t, states["check"].Skipped)
	assert.True(t, states["build"].Skipped)
	assert.True(t, states["cleanup"].Skipped)
}

func TestSkipSignalWithoutMappingFails(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	build := signalTestStep("build", nil)

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{check}},
			{Steps: []*backend_types.Step{build}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))
	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(r.Err(), &exitErr))
	assert.Equal(t, 78, exitErr.Code)
}

func TestSkipSignalRunningSiblingFailureStillFails(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	// check sends skip right away while its sibling is still running and then fails.
	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	check.ExitCodeSignals = skipOn78
	sibling := signalTestStep("sibling", map[string]string{
		dummy.EnvKeyStepSleep:    "50ms",
		dummy.EnvKeyStepExitCode: "1",
	})
	next := signalTestStep("next", nil)

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{check, sibling}},
			{Steps: []*backend_types.Step{next}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))

	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(r.Err(), &exitErr), "failure of a running sibling must fail the workflow, got err=%v states=%+v", r.Err(), finalStepStates(tracer))
	assert.Equal(t, 1, exitErr.Code)

	states := finalStepStates(tracer)
	assert.False(t, states["sibling"].Skipped, "running sibling finishes normally")
	assert.Equal(t, 1, states["sibling"].ExitCode)
	assert.True(t, states["next"].Skipped)
}

func TestSkipSignalKeepsEarlierError(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	fail := signalTestStep("fail", map[string]string{dummy.EnvKeyStepExitCode: "1"})
	// runs on failure and sends skip, the earlier error must remain.
	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	check.OnSuccess = false
	check.OnFailure = true
	check.ExitCodeSignals = skipOn78
	notify := signalTestStep("notify", nil)
	notify.OnFailure = true

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{fail}},
			{Steps: []*backend_types.Step{check}},
			{Steps: []*backend_types.Step{notify}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))

	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(r.Err(), &exitErr), "skip must not clear an earlier failure")
	assert.Equal(t, 1, exitErr.Code)
	assert.True(t, finalStepStates(tracer)["notify"].Skipped)
}

func TestSkipSignalIsolatedBetweenWorkflows(t *testing.T) {
	t.Parallel()

	// Two runtimes stand for two matrix workflows; skip in one must not affect the other.
	tracerA := newTestTracer(t)
	checkA := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	checkA.ExitCodeSignals = skipOn78
	rA := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{checkA}},
			{Steps: []*backend_types.Step{signalTestStep("build", nil)}},
		}},
		dummy.New(), WithTracer(tracerA), WithLogger(newTestLogger(t)),
	)

	tracerB := newTestTracer(t)
	checkB := signalTestStep("check", nil)
	checkB.ExitCodeSignals = skipOn78
	rB := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{checkB}},
			{Steps: []*backend_types.Step{signalTestStep("build", nil)}},
		}},
		dummy.New(), WithTracer(tracerB), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, rA.Run(t.Context()))
	require.NoError(t, rB.Run(t.Context()))

	assert.True(t, finalStepStates(tracerA)["build"].Skipped)
	statesB := finalStepStates(tracerB)
	assert.False(t, statesB["build"].Skipped)
	assert.True(t, hasTraceForStep(getTracerStates(tracerB), "build"))
	assert.NoError(t, rB.Err())
}

func TestSuccessSignalContinuesWorkflow(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	lint := signalTestStep("lint", map[string]string{dummy.EnvKeyStepExitCode: "1"})
	lint.ExitCodeSignals = map[int]backend_types.StepSignal{1: backend_types.StepSignalSuccess}
	build := signalTestStep("build", nil)

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{lint}},
			{Steps: []*backend_types.Step{build}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))
	assert.NoError(t, r.Err())

	states := finalStepStates(tracer)
	assert.Equal(t, 0, states["lint"].ExitCode)
	assert.False(t, states["build"].Skipped, "success must not skip the next steps")
	assert.True(t, states["build"].Exited)
	assert.False(t, r.skipRequested.Load())
}

func TestSignalsMixedInOneStep(t *testing.T) {
	t.Parallel()

	signals := map[int]backend_types.StepSignal{
		2:  backend_types.StepSignalSuccess,
		78: backend_types.StepSignalSkip,
	}
	for code, wantBuildSkipped := range map[string]bool{"2": false, "78": true} {
		t.Run("exit "+code, func(t *testing.T) {
			t.Parallel()
			tracer := newTestTracer(t)
			check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: code})
			check.ExitCodeSignals = signals
			r := New(
				&backend_types.Config{Stages: []*backend_types.Stage{
					{Steps: []*backend_types.Step{check}},
					{Steps: []*backend_types.Step{signalTestStep("build", nil)}},
				}},
				dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
			)
			require.NoError(t, r.Run(t.Context()))
			assert.NoError(t, r.Err())
			assert.Equal(t, wantBuildSkipped, finalStepStates(tracer)["build"].Skipped)
		})
	}
}

// restartBackend returns a mock backend whose WaitStep returns the given exit codes one after another.
func restartBackend(t *testing.T, exitCodes ...int) *mocks.MockBackend {
	t.Helper()
	engine := mocks.NewMockBackend(t)
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	engine.On("TailStep", mock.Anything, mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ *backend_types.Step, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		})
	for _, code := range exitCodes {
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(&backend_types.State{Exited: true, ExitCode: code}, nil).Once()
	}
	engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return engine
}

// recordingLogger collects everything written to the step logs.
func recordingLogger() (func(*backend_types.Step, io.ReadCloser) error, func() string) {
	var mu sync.Mutex
	var sb strings.Builder
	return func(_ *backend_types.Step, rc io.ReadCloser) error {
			b, _ := io.ReadAll(rc)
			mu.Lock()
			sb.Write(b)
			mu.Unlock()
			return rc.Close()
		}, func() string {
			mu.Lock()
			defer mu.Unlock()
			return sb.String()
		}
}

func TestRestartSignalSucceedsOnLaterAttempt(t *testing.T) {
	t.Parallel()
	engine := restartBackend(t, 75, 0)
	logFn, logs := recordingLogger()
	r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(logFn))
	step := dummyStep("flaky")
	step.ExitCodeSignals = map[int]backend_types.StepSignal{75: backend_types.StepSignalRestart}

	err := r.runBlockingStep(t.Context(), step)

	assert.NoError(t, err)
	engine.AssertNumberOfCalls(t, "StartStep", 2)
	assert.Contains(t, logs(), "starting attempt 2 of 3")
	assert.False(t, r.skipRequested.Load())
}

func TestRestartSignalFailsAfterMaxAttempts(t *testing.T) {
	t.Parallel()
	engine := restartBackend(t, 75, 75, 75)
	logFn, logs := recordingLogger()
	r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(logFn))
	step := dummyStep("flaky")
	step.ExitCodeSignals = map[int]backend_types.StepSignal{75: backend_types.StepSignalRestart}

	err := r.runBlockingStep(t.Context(), step)

	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 75, exitErr.Code)
	engine.AssertNumberOfCalls(t, "StartStep", backend_types.StepSignalRestartDefaultMax+1)
	assert.Contains(t, logs(), "starting attempt 3 of 3")
	assert.NotContains(t, logs(), "starting attempt 4")
}

func TestRestartThenOtherSignal(t *testing.T) {
	t.Parallel()
	// first attempt asks for restart, second one sends skip.
	engine := restartBackend(t, 75, 78)
	r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))
	step := dummyStep("check")
	step.ExitCodeSignals = map[int]backend_types.StepSignal{
		75: backend_types.StepSignalRestart,
		78: backend_types.StepSignalSkip,
	}

	err := r.runBlockingStep(t.Context(), step)

	assert.NoError(t, err)
	engine.AssertNumberOfCalls(t, "StartStep", 2)
	assert.True(t, r.skipRequested.Load())
}

func TestRestartSignalWithFailureIgnore(t *testing.T) {
	t.Parallel()
	engine := restartBackend(t, 75, 75, 75)
	r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))
	step := dummyStep("flaky")
	step.Failure = "ignore"
	step.ExitCodeSignals = map[int]backend_types.StepSignal{75: backend_types.StepSignalRestart}

	assert.NoError(t, r.runBlockingStep(t.Context(), step))
}

func TestCancelSignalCancelsWorkflow(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "70"})
	check.ExitCodeSignals = map[int]backend_types.StepSignal{70: backend_types.StepSignalCancel}
	// runs in parallel with check and gets stopped by the cancel signal.
	sibling := signalTestStep("sibling", map[string]string{dummy.EnvKeyStepSleep: "5s"})
	next := signalTestStep("next", nil)

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{check, sibling}},
			{Steps: []*backend_types.Step{next}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	// the agent reports ErrCancel as a canceled (killed) workflow.
	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)

	states := finalStepStates(tracer)
	assert.Equal(t, 0, states["check"].ExitCode)
	assert.NoError(t, states["check"].Error, "the cancelling step itself is successful")
	assert.ErrorIs(t, states["sibling"].Error, pipeline_errors.ErrCancel, "running sibling is stopped")
	_, nextTraced := states["next"]
	assert.False(t, nextTraced, "steps of later stages never start")
}

func TestCancelSignalIsolatedBetweenWorkflows(t *testing.T) {
	t.Parallel()

	checkA := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "70"})
	checkA.ExitCodeSignals = map[int]backend_types.StepSignal{70: backend_types.StepSignalCancel}
	rA := New(&backend_types.Config{Stages: []*backend_types.Stage{{Steps: []*backend_types.Step{checkA}}}},
		dummy.New(), WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))

	tracerB := newTestTracer(t)
	rB := New(&backend_types.Config{Stages: []*backend_types.Stage{
		{Steps: []*backend_types.Step{signalTestStep("check", nil)}},
		{Steps: []*backend_types.Step{signalTestStep("build", nil)}},
	}}, dummy.New(), WithTracer(tracerB), WithLogger(newTestLogger(t)))

	assert.ErrorIs(t, rA.Run(t.Context()), pipeline_errors.ErrCancel)
	assert.NoError(t, rB.Run(t.Context()))
	assert.True(t, finalStepStates(tracerB)["build"].Exited)
}

var pauseOn71 = map[int]backend_types.StepSignal{71: backend_types.StepSignalPause}

func pauseTestRuntime(t *testing.T, tracer *tracer_mocks.MockTracer, f PauseFunc, stages ...[]*backend_types.Step) *Runtime {
	t.Helper()
	cfg := &backend_types.Config{}
	for _, steps := range stages {
		cfg.Stages = append(cfg.Stages, &backend_types.Stage{Steps: steps})
	}
	opts := []Option{WithTracer(tracer), WithLogger(newTestLogger(t))}
	if f != nil {
		opts = append(opts, WithPauseFunc(f))
	}
	return New(cfg, dummy.New(), opts...)
}

func pausingStep() *backend_types.Step {
	s := signalTestStep("approve", map[string]string{dummy.EnvKeyStepExitCode: "71"})
	s.ExitCodeSignals = pauseOn71
	return s
}

func TestPauseSignalResumed(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	var pausedBy string
	var calls int
	r := pauseTestRuntime(t, tracer, func(_ context.Context, step *backend_types.Step) (bool, bool, error) {
		calls++
		pausedBy = step.Name
		return true, false, nil
	}, []*backend_types.Step{pausingStep()}, []*backend_types.Step{signalTestStep("deploy", nil)})

	require.NoError(t, r.Run(t.Context()))
	assert.NoError(t, r.Err())
	assert.Equal(t, 1, calls)
	assert.Equal(t, "approve", pausedBy)
	states := finalStepStates(tracer)
	assert.Equal(t, 0, states["approve"].ExitCode)
	assert.True(t, states["deploy"].Exited, "deploy runs after resume")
}

func TestPauseSignalDeclined(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := pauseTestRuntime(t, tracer, func(context.Context, *backend_types.Step) (bool, bool, error) {
		return false, false, nil
	}, []*backend_types.Step{pausingStep()}, []*backend_types.Step{signalTestStep("deploy", nil)})

	assert.ErrorIs(t, r.Run(t.Context()), pipeline_errors.ErrCancel)
	_, deployTraced := finalStepStates(tracer)["deploy"]
	assert.False(t, deployTraced, "deploy never starts after decline")
}

func TestPauseSignalCanceledWhilePaused(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	tracer := newTestTracer(t)
	cfg := &backend_types.Config{Stages: []*backend_types.Stage{
		{Steps: []*backend_types.Step{pausingStep()}},
		{Steps: []*backend_types.Step{signalTestStep("deploy", nil)}},
	}}
	r := New(cfg, dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)), WithContext(ctx),
		WithPauseFunc(func(pctx context.Context, _ *backend_types.Step) (bool, bool, error) {
			cancel() // e.g. the pipeline got canceled in the UI while paused
			<-pctx.Done()
			return false, false, pctx.Err()
		}))

	assert.ErrorIs(t, r.Run(t.Context()), pipeline_errors.ErrCancel)
	_, deployTraced := finalStepStates(tracer)["deploy"]
	assert.False(t, deployTraced)
}

func TestPauseSignalPauseFuncError(t *testing.T) {
	t.Parallel()
	r := pauseTestRuntime(t, newTestTracer(t), func(context.Context, *backend_types.Step) (bool, bool, error) {
		return false, false, errors.New("server unreachable")
	}, []*backend_types.Step{pausingStep()}, []*backend_types.Step{signalTestStep("deploy", nil)})

	assert.ErrorContains(t, r.Run(t.Context()), "server unreachable")
}

func TestPauseSignalWithoutPauseFuncContinues(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := pauseTestRuntime(t, tracer, nil, []*backend_types.Step{pausingStep()}, []*backend_types.Step{signalTestStep("deploy", nil)})

	require.NoError(t, r.Run(t.Context()))
	assert.True(t, finalStepStates(tracer)["deploy"].Exited)
}

func TestPauseSignalNotCalled(t *testing.T) {
	t.Parallel()
	notCalled := func(context.Context, *backend_types.Step) (bool, bool, error) {
		t.Error("pause func must not be called")
		return true, false, nil
	}

	t.Run("last stage", func(t *testing.T) {
		t.Parallel()
		r := pauseTestRuntime(t, newTestTracer(t), notCalled, []*backend_types.Step{pausingStep()})
		require.NoError(t, r.Run(t.Context()))
	})

	t.Run("skip in same stage", func(t *testing.T) {
		t.Parallel()
		skipper := signalTestStep("skipper", map[string]string{dummy.EnvKeyStepExitCode: "78"})
		skipper.ExitCodeSignals = skipOn78
		r := pauseTestRuntime(t, newTestTracer(t), notCalled,
			[]*backend_types.Step{pausingStep(), skipper},
			[]*backend_types.Step{signalTestStep("deploy", nil)})
		require.NoError(t, r.Run(t.Context()))
	})
}

func TestPauseSignalWaitsForParallelSiblings(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	sibling := signalTestStep("sibling", map[string]string{dummy.EnvKeyStepSleep: "50ms"})
	r := pauseTestRuntime(t, tracer, func(context.Context, *backend_types.Step) (bool, bool, error) {
		// the pause starts only after the whole stage finished.
		assert.True(t, finalStepStates(tracer)["sibling"].Exited, "sibling finished before pause")
		return true, false, nil
	}, []*backend_types.Step{pausingStep(), sibling}, []*backend_types.Step{signalTestStep("deploy", nil)})

	require.NoError(t, r.Run(t.Context()))
	assert.True(t, finalStepStates(tracer)["deploy"].Exited)
}

func TestRestartSignalMaxRestarts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		maxRestarts  int
		wantAttempts int
	}{
		{maxRestarts: 1, wantAttempts: 2},
		{maxRestarts: 4, wantAttempts: 5},
	} {
		t.Run(fmt.Sprintf("max_restarts=%d", tc.maxRestarts), func(t *testing.T) {
			t.Parallel()
			codes := make([]int, tc.wantAttempts)
			for i := range codes {
				codes[i] = 75
			}
			engine := restartBackend(t, codes...)
			logFn, logs := recordingLogger()
			r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(logFn))
			step := dummyStep("flaky")
			step.ExitCodeSignals = map[int]backend_types.StepSignal{75: backend_types.StepSignalRestart}
			step.MaxRestarts = tc.maxRestarts

			err := r.runBlockingStep(t.Context(), step)

			var exitErr *pipeline_errors.ExitError
			require.True(t, errors.As(err, &exitErr))
			engine.AssertNumberOfCalls(t, "StartStep", tc.wantAttempts)
			assert.Contains(t, logs(), fmt.Sprintf("starting attempt %d of %d", tc.wantAttempts, tc.wantAttempts))
			assert.NotContains(t, logs(), fmt.Sprintf("starting attempt %d", tc.wantAttempts+1))
		})
	}

	t.Run("succeeds before the limit", func(t *testing.T) {
		t.Parallel()
		engine := restartBackend(t, 75, 75, 75, 0)
		r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))
		step := dummyStep("flaky")
		step.ExitCodeSignals = map[int]backend_types.StepSignal{75: backend_types.StepSignalRestart}
		step.MaxRestarts = 5

		assert.NoError(t, r.runBlockingStep(t.Context(), step))
		engine.AssertNumberOfCalls(t, "StartStep", 4)
	})
}

func TestRestartMaxAttemptsDefault(t *testing.T) {
	assert.Equal(t, 3, (&backend_types.Step{}).RestartMaxAttempts())
	assert.Equal(t, 6, (&backend_types.Step{MaxRestarts: 5}).RestartMaxAttempts())
}

func TestPauseSignalTimeoutLogs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		resumeOnTO bool
		resume     bool
		timedOut   bool
		wantErr    bool
		wantLines  []string
	}{
		{
			name: "declined by timeout", resume: false, timedOut: true, wantErr: true,
			wantLines: []string{"waiting up to 10m0s for resume or decline, then: decline", "no decision within 10m0s, workflow declined"},
		},
		{
			name: "resumed by timeout", resumeOnTO: true, resume: true, timedOut: true,
			wantLines: []string{"waiting up to 10m0s for resume or decline, then: resume", "no decision within 10m0s, workflow resumed"},
		},
		{
			name: "resumed by user before timeout", resume: true,
			wantLines: []string{"waiting up to 10m0s", "--- woodpecker: workflow resumed ---"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tracer := newTestTracer(t)
			logFn, logs := recordingLogger()
			step := pausingStep()
			step.PauseTimeout = 10 * time.Minute
			step.PauseTimeoutResume = tc.resumeOnTO
			var gotStep *backend_types.Step
			r := New(&backend_types.Config{Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{step}},
				{Steps: []*backend_types.Step{signalTestStep("deploy", nil)}},
			}}, dummy.New(), WithTracer(tracer), WithLogger(logFn),
				WithPauseFunc(func(_ context.Context, s *backend_types.Step) (bool, bool, error) {
					gotStep = s
					return tc.resume, tc.timedOut, nil
				}))

			err := r.Run(t.Context())
			if tc.wantErr {
				assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
			} else {
				assert.NoError(t, err)
				assert.True(t, finalStepStates(tracer)["deploy"].Exited)
			}
			require.NotNil(t, gotStep)
			assert.Equal(t, 10*time.Minute, gotStep.PauseTimeout, "pause func gets the timeout of the step")
			for _, line := range tc.wantLines {
				assert.Contains(t, logs(), line)
			}
		})
	}
}
