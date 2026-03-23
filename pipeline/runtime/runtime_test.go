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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

//
// Step builder helpers.
//

func cmdStep(name string, opts ...func(*backend.Step)) *backend.Step {
	s := &backend.Step{
		Name:        name,
		UUID:        name + "-uuid",
		Type:        backend.StepTypeCommands,
		OnSuccess:   true,
		OnFailure:   false,
		Environment: map[string]string{},
		Commands:    []string{"echo " + name},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func withExitCode(code int) func(*backend.Step) {
	return func(s *backend.Step) {
		s.Environment[dummy.EnvKeyStepExitCode] = fmt.Sprintf("%d", code)
	}
}

func withFailure(mode string) func(*backend.Step) {
	return func(s *backend.Step) { s.Failure = mode }
}

func withOnFailure() func(*backend.Step) {
	return func(s *backend.Step) { s.OnSuccess = false; s.OnFailure = true }
}

func withDetached() func(*backend.Step) {
	return func(s *backend.Step) {
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "100ms"
	}
}

func withService() func(*backend.Step) {
	return func(s *backend.Step) {
		s.Type = backend.StepTypeService
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "100ms"
	}
}

func withPlugin() func(*backend.Step) {
	return func(s *backend.Step) {
		s.Type = backend.StepTypePlugin
		s.Environment[dummy.EnvKeyStepType] = "plugin"
	}
}

func withOOM() func(*backend.Step) {
	return func(s *backend.Step) {
		s.Environment[dummy.EnvKeyStepOOMKilled] = "true"
		s.Environment[dummy.EnvKeyStepExitCode] = "137"
	}
}

func withStartFail() func(*backend.Step) {
	return func(s *backend.Step) {
		s.Environment[dummy.EnvKeyStepStartFail] = "true"
	}
}

//
// Trace assertion helpers.
//

func findFirstTraceByName(traces []state.State, name string) *state.State {
	for i := range traces {
		if traces[i].CurrStep != nil && traces[i].CurrStep.Name == name {
			return &traces[i]
		}
	}
	return nil
}

func findLastTraceByName(traces []state.State, name string) *state.State {
	for i := len(traces) - 1; i >= 0; i-- {
		if traces[i].CurrStep != nil && traces[i].CurrStep.Name == name {
			return &traces[i]
		}
	}
	return nil
}

func findStartedTrace(traces []state.State, name string) *state.State {
	for i := range traces {
		if traces[i].CurrStep != nil && traces[i].CurrStep.Name == name && !traces[i].CurrStepState.Exited {
			return &traces[i]
		}
	}
	return nil
}

//
// Realistic workflow simulations.
//

func TestWorkflowCloneBuildDeploy(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("clone")}},
				{Steps: []*backend.Step{cmdStep("build")}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	traces := getTracerStates(tracer)
	assert.Len(t, traces, 6)
	for i := 0; i < 6; i += 2 {
		assert.False(t, traces[i].CurrStepState.Exited, "trace %d should be step-started", i)
		assert.True(t, traces[i+1].CurrStepState.Exited, "trace %d should be step-completed", i+1)
		assert.Equal(t, 0, traces[i+1].CurrStepState.ExitCode)
	}

	for _, name := range []string{"clone", "build", "deploy"} {
		last := findLastTraceByName(traces, name)
		require.NotNil(t, last, "%s should have a final trace", name)
		assert.True(t, last.CurrStepState.Exited, "%s last trace should be exited", name)
		assert.Equal(t, 0, last.CurrStepState.ExitCode, "%s should exit with code 0", name)
		assert.False(t, last.CurrStepState.OOMKilled, "%s should not be OOM killed", name)
	}
}

func TestWorkflowWithServiceStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("db", withService()),
					cmdStep("build"),
				}},
				{Steps: []*backend.Step{cmdStep("test")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
	traces := getTracerStates(tracer)
	if assert.Len(t, traces, 5) {
		assert.EqualValues(t, backend.State{}, traces[0].CurrStepState)
		assert.Greater(t, traces[2].CurrStepState.Started, int64(0))
		assert.EqualValues(t, backend.State{Started: traces[2].CurrStepState.Started, Exited: true}, traces[2].CurrStepState)
		assert.EqualValues(t, backend.State{}, traces[3].CurrStepState)
		assert.Greater(t, traces[4].CurrStepState.Started, int64(0))
		assert.EqualValues(t, backend.State{Started: traces[4].CurrStepState.Started, Exited: true}, traces[4].CurrStepState)

		assert.Greater(t, traces[4].Workflow.Started, int64(0))
		assert.EqualValues(t, state.State{
			Workflow: struct {
				Started int64 `json:"time"`
				Error   error `json:"error"`
			}{
				Started: traces[4].Workflow.Started,
			},
			CurrStep: &backend.Step{
				Name:        "test",
				UUID:        "test-uuid",
				Type:        "commands",
				OnSuccess:   true,
				Environment: map[string]string{},
				Commands:    []string{"echo test"},
			},
			CurrStepState: backend.State{
				Started: traces[4].CurrStepState.Started,
				Exited:  true,
			},
		}, traces[4])
	}
}

func TestWorkflowDetachedStepDoesNotBlockWorkflow(t *testing.T) {
	t.Parallel()
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("background-worker", withDetached()),
					cmdStep("main-build"),
				}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
}

func TestWorkflowBuildFailSkipsSubsequentStages(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("clone")}},
				{Steps: []*backend.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 1, exitErr.Code)

	traces := getTracerStates(tracer)

	buildTrace := findLastTraceByName(traces, "build")
	require.NotNil(t, buildTrace, "build step should fail")
	assert.EqualValues(t, 1, buildTrace.CurrStepState.ExitCode)

	deployTrace := findLastTraceByName(traces, "deploy")
	require.NotNil(t, deployTrace, "deploy step should still be traced")
	assert.True(t, deployTrace.CurrStepState.Skipped)
}

func TestWorkflowOnFailureStepRuns(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build", withExitCode(2))}},
				{Steps: []*backend.Step{cmdStep("notify-failure", withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())
	traces := getTracerStates(tracer)

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(traces, "notify-failure"), "OnFailure step should have started")

	last := findLastTraceByName(traces, "notify-failure")
	require.NotNil(t, last)
	assert.Greater(t, last.CurrStepState.Started, int64(0), "step should have started")
	assert.EqualValues(t, backend.State{Started: last.CurrStepState.Started, Exited: true}, last.CurrStepState)
}

func TestWorkflowOnFailureStepSkippedOnSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build")}},
				{Steps: []*backend.Step{cmdStep("cleanup-on-fail", withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())
	require.NoError(t, err)
	traces := getTracerStates(tracer)

	firstCleanupTrace := findFirstTraceByName(traces, "cleanup-on-fail")
	lastCleanupTrace := findLastTraceByName(traces, "cleanup-on-fail")
	assert.Equal(t, firstCleanupTrace, lastCleanupTrace, "we expect on skipped steps to only have one trace")
	assert.True(t, lastCleanupTrace.CurrStepState.Skipped, "cleanup-on-fail should be skipped after no failure happened")
}

func TestWorkflowFailureIgnore(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("lint", withExitCode(1), withFailure(metadata.FailureIgnore)),
				}},
				{Steps: []*backend.Step{cmdStep("build")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err, "pipeline should succeed when failing step has failure=ignore")
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "build"), "build step should run after ignored failure")

	last := findLastTraceByName(getTracerStates(tracer), "build")
	require.NotNil(t, last)
	assert.True(t, last.CurrStepState.Exited)
	assert.Equal(t, 0, last.CurrStepState.ExitCode)
}

func TestWorkflowFailureIgnoreDoesNotSetWorkflowError(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("flaky-test", withExitCode(1), withFailure(metadata.FailureIgnore)),
				}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	traces := getTracerStates(tracer)
	firstDeployTrace := findFirstTraceByName(traces, "deploy")
	lastDeployTrace := findLastTraceByName(traces, "deploy")
	assert.NotEqualValues(t, firstDeployTrace, lastDeployTrace, "we expect two traces")
	assert.False(t, lastDeployTrace.CurrStepState.Skipped, "deploy should not be skipped after failure=ignore step")
}

func TestWorkflowPluginStep(t *testing.T) {
	t.Parallel()
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("clone")}},
				{Steps: []*backend.Step{cmdStep("publish", withPlugin())}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
}

func TestWorkflowOOMKilledStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build", withOOM())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	var oomErr *pipeline_errors.OomError
	assert.True(t, errors.As(err, &oomErr))

	last := findLastTraceByName(getTracerStates(tracer), "build")
	require.NotNil(t, last)
	assert.True(t, last.CurrStepState.Exited)
	assert.True(t, last.CurrStepState.OOMKilled)
	assert.Equal(t, 137, last.CurrStepState.ExitCode)
}

func TestWorkflowParallelStepsInStage(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("clone")}},
				{Steps: []*backend.Step{
					cmdStep("test-unit"),
					cmdStep("test-integration"),
					cmdStep("test-e2e"),
				}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	assert.Len(t, getTracerStates(tracer), 10)
}

func TestWorkflowParallelStepOneFailsOthersComplete(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("test-fast"),
					cmdStep("test-slow", withExitCode(1)),
				}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Len(t, getTracerStates(tracer), 4, "both parallel steps should complete and be traced")

	lastFast := findLastTraceByName(getTracerStates(tracer), "test-fast")
	require.NotNil(t, lastFast)
	assert.True(t, lastFast.CurrStepState.Exited)
	assert.Equal(t, 0, lastFast.CurrStepState.ExitCode, "test-fast should succeed")

	lastSlow := findLastTraceByName(getTracerStates(tracer), "test-slow")
	require.NotNil(t, lastSlow)
	assert.True(t, lastSlow.CurrStepState.Exited)
	assert.Equal(t, 1, lastSlow.CurrStepState.ExitCode, "test-slow should fail with code 1")
}

func TestWorkflowStepStartFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build")}},
				{Steps: []*backend.Step{cmdStep("deploy", withStartFail())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	deployTrace := findFirstTraceByName(getTracerStates(tracer), "build")
	require.NotNil(t, deployTrace)
	assert.EqualValues(t, backend.State{}, deployTrace.CurrStepState)
}

func TestWorkflowContextCancelDuringExecution(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.CurrStepState.Exited && !s.CurrStepState.Skipped {
			stageCount++
			if stageCount >= 1 {
				cancel(nil)
			}
		}
	}).Return(nil).Maybe()

	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build")}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithContext(ctx),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
}

func TestWorkflowSetupFailure(t *testing.T) {
	t.Parallel()
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build")}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithTaskUUID(dummy.WorkflowSetupFailUUID),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected fail to setup workflow")
}

func TestWorkflowServiceWithParallelBuildAndOnFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("redis", withService()),
					cmdStep("clone"),
				}},
				{Steps: []*backend.Step{
					cmdStep("build"),
					cmdStep("lint", withExitCode(1)),
				}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
				{Steps: []*backend.Step{cmdStep("notify", withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	traces := getTracerStates(tracer)

	notifyTrace := findLastTraceByName(traces, "notify")
	require.NotNil(t, notifyTrace)
	assert.True(t, notifyTrace.CurrStepState.Exited, "notify should exited")
	assert.EqualValues(t, 0, notifyTrace.CurrStepState.ExitCode, "notify should be successful")

	lastBuild := findLastTraceByName(traces, "lint")
	require.NotNil(t, lastBuild)
	assert.True(t, lastBuild.CurrStepState.Exited)
	assert.Equal(t, 1, lastBuild.CurrStepState.ExitCode, "lint should have failed")

	deployTrace := findFirstTraceByName(traces, "deploy")
	require.NotNil(t, deployTrace)
	assert.True(t, deployTrace.CurrStepState.Skipped, "deploy should be skipped after lint failure")

	assert.NotNil(t, findStartedTrace(traces, "notify"),
		"notify (OnFailure) should have started")
}

func TestWorkflowIgnoredFailureFollowedByOnFailureStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{
					cmdStep("lint", withExitCode(1), withFailure(metadata.FailureIgnore)),
				}},
				{Steps: []*backend.Step{cmdStep("error-notify", withOnFailure())}},
				{Steps: []*backend.Step{cmdStep("build")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	traces := getTracerStates(tracer)

	notifyTrace := findFirstTraceByName(traces, "error-notify")
	require.NotNil(t, notifyTrace)
	assert.True(t, notifyTrace.CurrStepState.Skipped, "OnFailure step should be skipped when prior failure was ignored")

	assert.NotNil(t, findStartedTrace(traces, "build"), "build should run after ignored failure")
}

func TestWorkflowEmptyStages(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{Stages: []*backend.Stage{}},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	assert.Empty(t, getTracerStates(tracer))
}
