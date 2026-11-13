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
		env = make(map[string]string)
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
	assert.NoError(t, states["check"].Error, "the canceling step itself is successful")
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

// The skip takes precedence over the status flags: when a running sibling fails,
// the workflow fails, but a following failure step is still skipped.
func TestSkipSignalSkipsFailureStepAfterSiblingFailed(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)

	check := signalTestStep("check", map[string]string{dummy.EnvKeyStepExitCode: "78"})
	check.ExitCodeSignals = skipOn78
	sibling := signalTestStep("sibling", map[string]string{dummy.EnvKeyStepExitCode: "1"})
	notify := signalTestStep("notify", nil)
	notify.OnSuccess = false
	notify.OnFailure = true

	r := New(
		&backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{check, sibling}},
			{Steps: []*backend_types.Step{notify}},
		}},
		dummy.New(), WithTracer(tracer), WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))

	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(r.Err(), &exitErr), "the failed sibling fails the workflow")
	assert.Equal(t, 1, exitErr.Code)
	assert.True(t, finalStepStates(tracer)["notify"].Skipped, "skip wins over status: failure")
}
