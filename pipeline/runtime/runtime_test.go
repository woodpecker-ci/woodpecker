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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	tracer_mocks "go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing/mocks"
)

//
// Step builder helpers.
//

func cmdStep(name string, opts ...func(*backend_types.Step)) *backend_types.Step {
	s := &backend_types.Step{
		Name:        name,
		UUID:        name + "-uuid",
		Type:        backend_types.StepTypeCommands,
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

func withExitCode(code int) func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Environment[dummy.EnvKeyStepExitCode] = fmt.Sprintf("%d", code)
	}
}

func withIgnoreFailure() func(*backend_types.Step) {
	return func(s *backend_types.Step) { s.Failure = string(metadata.FailureIgnore) }
}

func withOnFailure() func(*backend_types.Step) {
	return func(s *backend_types.Step) { s.OnSuccess = false; s.OnFailure = true }
}

func withDetached() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "100ms"
	}
}

// withUnboundedDetached models a detached step that runs until the workflow tears it down.
func withUnboundedDetached() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Type = backend_types.StepTypeService
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "3m"
	}
}

func withService() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Type = backend_types.StepTypeService
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "100ms"
	}
}

// withUnboundedService models a real-world service that runs until the workflow tears it down.
func withUnboundedService() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Type = backend_types.StepTypeService
		s.Detached = true
		s.Environment[dummy.EnvKeyStepSleep] = "3m"
	}
}

func withPlugin() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Type = backend_types.StepTypePlugin
		s.Environment[dummy.EnvKeyStepType] = "plugin"
	}
}

func withOOM() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Environment[dummy.EnvKeyStepOOMKilled] = "true"
		s.Environment[dummy.EnvKeyStepExitCode] = "137"
	}
}

func withStartFail() func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Environment[dummy.EnvKeyStepStartFail] = "true"
	}
}

func withSleep(d string) func(*backend_types.Step) {
	return func(s *backend_types.Step) {
		s.Environment[dummy.EnvKeyStepSleep] = d
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("clone")}},
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))

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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("db", withService()),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("test", withSleep("250ms"))}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))
	traces := getTracerStates(tracer)

	// Each step should emit exactly one "started" and one "exited" trace:
	// db (service/detached), build, test — 3 * 2 = 6 traces total.
	require.Len(t, traces, 6)

	// Per-step invariants: started trace is the zero state, exited trace is
	// Exited=true with a monotonic Started timestamp.
	for _, name := range []string{"db", "build", "test"} {
		started := findFirstTraceByName(traces, name)
		require.NotNil(t, started, "%s should have a started trace", name)
		assert.EqualValues(t, backend_types.State{}, started.CurrStepState,
			"%s started trace should be zero-valued", name)

		last := findLastTraceByName(traces, name)
		require.NotNil(t, last, "%s should have an exited trace", name)
		assert.True(t, last.CurrStepState.Exited, "%s should be exited", name)
		assert.Equal(t, 0, last.CurrStepState.ExitCode, "%s should exit 0", name)
		assert.Greater(t, last.CurrStepState.Started, int64(0),
			"%s should have a non-zero Started timestamp", name)
	}

	// Per-step ordering: started trace precedes exited trace for the same step.
	for _, name := range []string{"db", "build", "test"} {
		startedIdx := indexOfTrace(traces, func(s state.State) bool {
			return s.CurrStep != nil && s.CurrStep.Name == name && !s.CurrStepState.Exited
		})
		exitedIdx := indexOfTrace(traces, func(s state.State) bool {
			return s.CurrStep != nil && s.CurrStep.Name == name && s.CurrStepState.Exited
		})
		assert.Less(t, startedIdx, exitedIdx, "%s started must precede %s exited", name, name)
	}

	// The contract of a service/detached step: it does not block the next
	// stage. Verify that stage 2's `test` step started before db (in stage 1)
	// reported its exit — i.e. test was running in parallel with db, not
	// queued behind it.
	dbExitIdx := indexOfTrace(traces, func(s state.State) bool {
		return s.CurrStep != nil && s.CurrStep.Name == "db" && s.CurrStepState.Exited
	})
	testStartedIdx := indexOfTrace(traces, func(s state.State) bool {
		return s.CurrStep != nil && s.CurrStep.Name == "test" && !s.CurrStepState.Exited
	})
	assert.Less(t, testStartedIdx, dbExitIdx,
		"test (next stage) must start before db (service) exits — otherwise db blocked stage 2")

	// Runtime-injected env vars should be present on the test step's exit trace.
	testExit := findLastTraceByName(traces, "test")
	require.NotNil(t, testExit)
	assert.NotEmpty(t, testExit.CurrStep.Environment["CI_PIPELINE_STARTED"])
	assert.NotEmpty(t, testExit.CurrStep.Environment["CI_STEP_STARTED"])
	assert.Greater(t, testExit.Workflow.Started, int64(0))

	// Strip runtime-injected env for a structural comparison of the step itself.
	delete(testExit.CurrStep.Environment, "CI_PIPELINE_STARTED")
	delete(testExit.CurrStep.Environment, "CI_STEP_STARTED")
	delete(testExit.CurrStep.Environment, dummy.EnvKeyStepSleep)
	assert.EqualValues(t, state.State{
		Workflow: state.Workflow{Started: testExit.Workflow.Started},
		CurrStep: &backend_types.Step{
			Name:        "test",
			UUID:        "test-uuid",
			Type:        "commands",
			OnSuccess:   true,
			Environment: map[string]string{},
			Commands:    []string{"echo test"},
		},
		CurrStepState: backend_types.State{
			Started: testExit.CurrStepState.Started,
			Exited:  true,
		},
	}, *testExit)
}

func TestWorkflowDetachedStepDoesNotBlockWorkflow(t *testing.T) {
	t.Parallel()
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("background-worker", withDetached()),
					cmdStep("main-build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("clone")}},
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
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
	assert.True(t, buildTrace.CurrStepState.Exited, "build should have started")

	buildTrace = findLastTraceByName(traces, "build")
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(2))}},
				{Steps: []*backend_types.Step{cmdStep("notify-failure", withOnFailure())}},
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
	assert.EqualValues(t, backend_types.State{Started: last.CurrStepState.Started, Exited: true}, last.CurrStepState)
}

func TestWorkflowOnFailureStepSkippedOnSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("cleanup-on-fail", withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	require.NoError(t, r.Run(t.Context()))

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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("lint", withExitCode(1), withIgnoreFailure()),
				}},
				{Steps: []*backend_types.Step{cmdStep("build")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()), "pipeline should succeed when failing step has failure=ignore")

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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("flaky-test", withExitCode(1), withIgnoreFailure()),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))

	traces := getTracerStates(tracer)
	firstDeployTrace := findFirstTraceByName(traces, "deploy")
	lastDeployTrace := findLastTraceByName(traces, "deploy")
	assert.NotEqualValues(t, firstDeployTrace, lastDeployTrace, "we expect two traces")
	assert.False(t, lastDeployTrace.CurrStepState.Skipped, "deploy should not be skipped after failure=ignore step")
}

func TestWorkflowPluginStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("clone")}},
				{Steps: []*backend_types.Step{cmdStep("publish", withPlugin())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))

	lastPluginTrace := findLastTraceByName(getTracerStates(tracer), "publish")
	if assert.NotNil(t, lastPluginTrace) {
		delete(lastPluginTrace.CurrStep.Environment, "CI_PIPELINE_STARTED")
		delete(lastPluginTrace.CurrStep.Environment, "CI_STEP_STARTED")

		assert.EqualValues(t, map[string]string{
			"DRONE_BUILD_STATUS":             "success",
			"DRONE_REPO_SCM":                 "git",
			"EXPECT_TYPE":                    "plugin",
			"PULLREQUEST_DRONE_PULL_REQUEST": "0",
		}, lastPluginTrace.CurrStep.Environment)
	}
}

func TestWorkflowOOMKilledStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withOOM())}},
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("clone")}},
				{Steps: []*backend_types.Step{
					cmdStep("test-unit"),
					cmdStep("test-integration"),
					cmdStep("test-e2e"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))

	assert.Len(t, getTracerStates(tracer), 10)
}

func TestWorkflowParallelStepOneFailsOthersComplete(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("test-fast"),
					cmdStep("test-slow", withExitCode(1)),
				}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.Error(t, r.Run(t.Context()))

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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("deploy", withStartFail())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.Error(t, r.Run(t.Context()))

	deployTrace := findFirstTraceByName(getTracerStates(tracer), "build")
	require.NotNil(t, deployTrace)
	assert.EqualValues(t, backend_types.State{}, deployTrace.CurrStepState)
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
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
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("redis", withService()),
					cmdStep("clone"),
				}},
				{Steps: []*backend_types.Step{
					cmdStep("build"),
					cmdStep("lint", withExitCode(1)),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
				{Steps: []*backend_types.Step{cmdStep("notify", withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.Error(t, r.Run(t.Context()))

	traces := getTracerStates(tracer)

	assert.NotNil(t, findStartedTrace(traces, "notify"), "notify (OnFailure) should have started")
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
}

func TestWorkflowIgnoredFailureFollowedByOnFailureStep(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("lint", withExitCode(1), withIgnoreFailure()),
				}},
				{Steps: []*backend_types.Step{cmdStep("error-notify", withOnFailure())}},
				{Steps: []*backend_types.Step{cmdStep("build")}},
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
		&backend_types.Config{Stages: []*backend_types.Stage{}},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	assert.Empty(t, getTracerStates(tracer))
}

//
// outcome: failure
//

func TestPluginStepFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("publish", withPlugin(), withExitCode(1))}},
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

	last := findLastTraceByName(getTracerStates(tracer), "publish")
	require.NotNil(t, last)
	assert.True(t, last.CurrStepState.Exited)
	assert.Equal(t, 1, last.CurrStepState.ExitCode)
}

func TestDetachedStepFailure(t *testing.T) {
	t.Parallel()
	// A detached step that exits non-zero; since it is detached the runtime
	// only waits for setup, so the pipeline itself should still succeed.
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("background", withDetached(), withExitCode(1)),
					cmdStep("build"),
				}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	// Detached step errors are not propagated to the pipeline result.
	assert.NoError(t, r.Run(t.Context()))
}

func TestServiceStepFailure(t *testing.T) {
	t.Parallel()
	// A service that exits non-zero; same semantics as detached — the pipeline
	// should still complete because services are fire-and-forget from the
	// runtime's perspective.
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("db", withService(), withExitCode(1)),
					cmdStep("test"),
				}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
}

//
// outcome: start failure
//

func TestPluginStepStartFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("publish", withPlugin(), withStartFail())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
}

func TestDetachedStepStartFailure(t *testing.T) {
	t.Parallel()
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("background", withDetached(), withStartFail()),
					cmdStep("build"),
				}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	// A detached step that fails to start should surface the error, since the
	// runtime waits for setup to complete before continuing.
	err := r.Run(t.Context())

	assert.Error(t, err)
}

func TestServiceStepStartFailure(t *testing.T) {
	t.Parallel()
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("db", withService(), withStartFail()),
					cmdStep("test"),
				}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
}

//
// Run condition: OnFailure for plugin / detached / service.
//

func TestPluginOnFailureStepRuns(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend_types.Step{cmdStep("notify", withPlugin(), withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "notify"),
		"plugin OnFailure step should have started")

	last := findLastTraceByName(getTracerStates(tracer), "notify")
	require.NotNil(t, last)
	assert.True(t, last.CurrStepState.Exited)
	assert.Equal(t, 0, last.CurrStepState.ExitCode)
}

func TestPluginOnFailureStepSkippedOnSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("notify", withPlugin(), withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	trace := findLastTraceByName(getTracerStates(tracer), "notify")
	trace.CurrStepState.Started = 0
	assert.EqualValues(t, backend_types.State{Skipped: true}, trace.CurrStepState,
		"plugin OnFailure step should not run when pipeline succeeds")
}

func TestDetachedOnFailureStepRuns(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend_types.Step{cmdStep("cleanup", withDetached(), withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "cleanup"),
		"detached OnFailure step should have started")
}

func TestDetachedOnFailureStepSkippedOnSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("cleanup", withDetached(), withOnFailure())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	trace := findLastTraceByName(getTracerStates(tracer), "cleanup")
	trace.CurrStepState.Started = 0
	assert.EqualValues(t, backend_types.State{Skipped: true}, trace.CurrStepState,
		"detached OnFailure step should not run when pipeline succeeds")
}

//
// Run condition: OnSuccess=true + OnFailure=true (always-run).
//

func withAlwaysRun() func(*backend_types.Step) {
	return func(s *backend_types.Step) { s.OnSuccess = true; s.OnFailure = true }
}

func TestAlwaysRunStepRunsOnSuccess(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("report", withAlwaysRun())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	last := findLastTraceByName(getTracerStates(tracer), "report")
	require.NotNil(t, last, "always-run step should be traced")
	assert.True(t, last.CurrStepState.Exited)
	assert.Equal(t, 0, last.CurrStepState.ExitCode)
}

func TestAlwaysRunStepRunsOnFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend_types.Step{cmdStep("report", withAlwaysRun())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "report"),
		"always-run step should start even when pipeline is failing")

	last := findLastTraceByName(getTracerStates(tracer), "report")
	require.NotNil(t, last)
	assert.True(t, last.CurrStepState.Exited)
	assert.Equal(t, 0, last.CurrStepState.ExitCode)
}

func TestAlwaysRunPluginRunsOnFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build", withExitCode(1))}},
				{Steps: []*backend_types.Step{cmdStep("report", withPlugin(), withAlwaysRun())}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "report"),
		"always-run plugin step should start even when pipeline is failing")
}

//
// Failure handling: failure=ignore for plugin.
//

func TestPluginFailureIgnore(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("lint", withPlugin(), withExitCode(1), withIgnoreFailure()),
				}},
				{Steps: []*backend_types.Step{cmdStep("build")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err, "pipeline should succeed when failing plugin has failure=ignore")
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "build"),
		"build step should run after ignored plugin failure")
}

func TestDetachedFailureIgnore(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("watcher", withDetached(), withExitCode(1), withIgnoreFailure()),
					cmdStep("build"),
				}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
}

//
// Cancellation.
//

func TestWorkflowContextCancelWithPluginStep(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.CurrStepState.Exited {
			stageCount++
			if stageCount >= 1 {
				cancel(nil)
			}
		}
	}).Return(nil).Maybe()

	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{cmdStep("build")}},
				{Steps: []*backend_types.Step{cmdStep("publish", withPlugin())}},
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

func TestWorkflowContextCancelWithDetachedStep(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.CurrStepState.Exited {
			stageCount++
			if stageCount >= 1 {
				cancel(nil)
			}
		}
	}).Return(nil).Maybe()

	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("background", withDetached()),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
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

func TestWorkflowContextCancelWithServiceStep(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.CurrStepState.Exited {
			stageCount++
			if stageCount >= 1 {
				cancel(nil)
			}
		}
	}).Return(nil).Maybe()

	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("db", withService()),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
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

// TestWorkflowCancelDuringStepSleep verifies that canceling the workflow context
// while a step is sleeping (via SLEEP env) causes the runtime to return ErrCancel
// promptly — without waiting the full sleep duration — and that subsequent stages
// are never executed.
//
// The tracer callback cancels the context the moment the first stage ("prepare")
// completes. The "slow" step uses a short sleep so that even if WaitStep enters
// the sleep select, the context cancellation unblocks it quickly.
//
// Note: we do not assert on the slow step's exit code here because Run() may
// return (via ctx.Done()) before the stage goroutine's WaitStep completes,
// causing DestroyWorkflow to clean up state that WaitStep still needs. The
// exit-code-130 behavior of a canceled sleep is verified at the backend unit
// level in TestWaitStepCanceledBySleep.
func TestWorkflowCancelDuringStepSleep(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(t.Context())

	var prepareExited int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s == nil || s.CurrStep == nil {
			return
		}
		// Cancel as soon as the first stage ("prepare") finishes.
		if s.CurrStep.Name == "prepare" && s.CurrStepState.Exited {
			prepareExited++
			if prepareExited >= 1 {
				cancel(nil)
			}
		}
	}).Return(nil).Maybe()

	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("prepare"),
				}},
				{Steps: []*backend_types.Step{
					// Short sleep so the test doesn't hang if WaitStep enters the timer.
					cmdStep("slow", func(s *backend_types.Step) {
						s.Environment[dummy.EnvKeyStepSleep] = "100ms"
					}),
				}},
				{Steps: []*backend_types.Step{
					cmdStep("never-reached"),
				}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithContext(ctx),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())
	assert.ErrorIs(t, err, pipeline_errors.ErrCancel, "canceled workflow must return ErrCancel")

	// Give the orphaned stage goroutine a moment to finish tracing (best effort).
	time.Sleep(200 * time.Millisecond)

	assert.Nil(t, findFirstTraceByName(getTracerStates(tracer), "never-reached"),
		"never-reached must not have been traced")
}

// TestWorkflowFailingServiceDoesNotFailWorkflow pins down the intentional design:
// a service/detached step that fails in the background has its failure logged
// and traced, but it must NOT propagate to the workflow error. Subsequent
// stages must still run, and Run() must return nil.
//
// This is the explicit contract in runDetachedStep:
// "Any error that occurs after setup is logged but not propagated — it cannot
//
//	influence the pipeline outcome at that point."
func TestWorkflowFailingServiceDoesNotFailWorkflow(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					// Service runs ~100ms (from withService), then exits non-zero.
					cmdStep("db", withService(), withExitCode(1)),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy", withSleep("120ms"))}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	// Contract 1: workflow succeeds even though the service failed.
	assert.NoError(t, r.Run(t.Context()),
		"service failure must not fail the workflow (detached errors are not propagated)")

	traces := getTracerStates(tracer)

	// Contract 2: the service's failure IS visible in traces. This is the
	// observability guarantee — the failure is logged and recorded even though
	// it doesn't kill the workflow.
	dbExit := findLastTraceByName(traces, "db")
	require.NotNil(t, dbExit, "db must have an exit trace")
	assert.True(t, dbExit.CurrStepState.Exited, "db should be marked exited")
	assert.Equal(t, 1, dbExit.CurrStepState.ExitCode, "db exit code must be preserved in trace")

	// Contract 3: deploy must run normally — NOT skipped — because the service
	// failure didn't set r.err.
	deployExit := findLastTraceByName(traces, "deploy")
	require.NotNil(t, deployExit, "deploy must be traced")
	assert.False(t, deployExit.CurrStepState.Skipped, "deploy must run when only a service failed")
	assert.True(t, deployExit.CurrStepState.Exited, "deploy should complete normally")
	assert.Equal(t, 0, deployExit.CurrStepState.ExitCode)

	// Contract 4: uploadWait at the end of Run() guarantees the detached trace
	// has been emitted BEFORE Run() returns. This is non-timing-dependent:
	// if Run() returned, the exit trace for every detached step must exist.
	// This is what the uploadWait plumbing in this PR is actually for.
	assert.NotNil(t, findLastTraceByName(traces, "db"),
		"detached step exit trace must be emitted before Run() returns (uploadWait contract)")
}

// TestWorkflowFailingDetachedStepDoesNotFailWorkflow is the non-service
// counterpart: Detached=true, Type=commands (a background worker). Same
// contract — failures don't propagate.
func TestWorkflowFailingDetachedStepDoesNotFailWorkflow(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					// Detached (non-service) worker, ~100ms (from withDetached), exits code 2.
					cmdStep("background-worker", withDetached(), withExitCode(2)),
					cmdStep("main-build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("deploy")}},
			},
		},
		dummy.New(),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()),
		"detached worker failure must not fail the workflow")

	time.Sleep(100 * time.Millisecond)

	traces := getTracerStates(tracer)

	workerExit := findLastTraceByName(traces, "background-worker")
	require.NotNil(t, workerExit, "background-worker must have an exit trace")
	assert.True(t, workerExit.CurrStepState.Exited)
	assert.Equal(t, 2, workerExit.CurrStepState.ExitCode,
		"exit code from detached step must be preserved in trace")

	deployExit := findLastTraceByName(traces, "deploy")
	require.NotNil(t, deployExit, "deploy must be traced")
	assert.False(t, deployExit.CurrStepState.Skipped,
		"deploy must run when only a detached worker failed")
	assert.True(t, deployExit.CurrStepState.Exited)
	assert.Equal(t, 0, deployExit.CurrStepState.ExitCode)
}

// TestWorkflowUnboundedServiceDoesNotHang asserts that when all normal steps
// have finished, a long-running service does NOT keep the workflow blocked
// forever. The runtime must tear the service down on its own (the whole point
// of declaring a step as a service is that it runs alongside the build, not
// that the build waits for it).
//
// Regression for https://github.com/woodpecker-ci/woodpecker/commit/4dd3be7f96
// which moved the upload waitgroup from per-upload (logger/tracer) to
// per-detached-goroutine. The detached goroutine wraps WaitStep, which on
// services blocks until the workflow context is canceled — so the workflow
// hangs waiting for its own service to exit.
func TestWorkflowUnboundedServiceDoesNotHang(t *testing.T) {
	t.Parallel()
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("db", withUnboundedService()),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("test")}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	// Use a deadline well below the dummy backend's testServiceTimeout (1s) so
	// that if this test "passes" it's because the runtime tore the service down,
	// not because dummy's safety timeout fired.
	done := make(chan error, 1)
	go func() { done <- r.Run(t.Context()) }()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("workflow hung: runtime did not tear down the unbounded service after normal steps finished")
	}
}

// TestWorkflowUnboundedDetachedDoesNotHang is the same as the service test but
// for plain detached steps (Detached=true, Type=commands). The bug is the same
// — a long-running detached step also pins the upload waitgroup.
func TestWorkflowUnboundedDetachedDoesNotHang(t *testing.T) {
	t.Parallel()
	r := New(
		&backend_types.Config{
			Stages: []*backend_types.Stage{
				{Steps: []*backend_types.Step{
					cmdStep("background-worker", withUnboundedDetached()),
					cmdStep("build"),
				}},
				{Steps: []*backend_types.Step{cmdStep("test")}},
			},
		},
		dummy.New(),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	done := make(chan error, 1)
	go func() { done <- r.Run(t.Context()) }()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("workflow hung: runtime did not tear down the unbounded detached step after normal steps finished")
	}
}
