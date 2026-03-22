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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/dummy"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
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
		if traces[i].Pipeline.Step != nil && traces[i].Pipeline.Step.Name == name {
			return &traces[i]
		}
	}
	return nil
}

func findLastTraceByName(traces []state.State, name string) *state.State {
	for i := len(traces) - 1; i >= 0; i-- {
		if traces[i].Pipeline.Step != nil && traces[i].Pipeline.Step.Name == name {
			return &traces[i]
		}
	}
	return nil
}

func findStartedTrace(traces []state.State, name string) *state.State {
	for i := range traces {
		if traces[i].Pipeline.Step != nil && traces[i].Pipeline.Step.Name == name && !traces[i].Process.Exited {
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
		assert.False(t, traces[i].Process.Exited, "trace %d should be step-started", i)
		assert.True(t, traces[i+1].Process.Exited, "trace %d should be step-completed", i+1)
		assert.Equal(t, 0, traces[i+1].Process.ExitCode)
	}

	for _, name := range []string{"clone", "build", "deploy"} {
		last := findLastTraceByName(traces, name)
		require.NotNil(t, last, "%s should have a final trace", name)
		assert.True(t, last.Process.Exited, "%s last trace should be exited", name)
		assert.Equal(t, 0, last.Process.ExitCode, "%s should exit with code 0", name)
		assert.False(t, last.Process.OOMKilled, "%s should not be OOM killed", name)
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
		assert.EqualValues(t, backend.State{}, traces[0].Process)
		assert.Greater(t, traces[2].Process.Started, int64(0))
		assert.EqualValues(t, backend.State{Started: traces[2].Process.Started, Exited: true}, traces[2].Process)
		assert.EqualValues(t, backend.State{}, traces[3].Process)
		assert.Greater(t, traces[4].Process.Started, int64(0))
		assert.EqualValues(t, backend.State{Started: traces[4].Process.Started, Exited: true}, traces[4].Process)

		assert.Greater(t, traces[4].Pipeline.Started, int64(0))
		assert.EqualValues(t, traces[4], state.State{
			Pipeline: struct {
				Started int64         `json:"time"`
				Step    *backend.Step `json:"step"`
				Error   error         `json:"error"`
			}{
				Started: traces[4].Pipeline.Started,
				Step: &backend.Step{
					Name:      "test",
					UUID:      "test-uuid",
					Type:      "commands",
					OnSuccess: true,
					Environment: map[string]string{
						"DRONE_BUILD_STATUS":             "success",
						"DRONE_REPO_SCM":                 "git",
						"PULLREQUEST_DRONE_PULL_REQUEST": "0",
					},
					Commands: []string{"echo test"},
				},
			},
			Process: backend.State{
				Started: traces[4].Process.Started,
				Exited:  true,
			},
		})
	}
}

func TestWorkflowDetachedStepDoesNotBlockPipeline(t *testing.T) {
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

	// traces := getTracerStates(tracer)

	// TODO: signal failed back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// deployTrace := findFirstTraceByName(calls, "build")
	// require.NotNil(t, deployTrace, "build step should fail")
	// assert.EqualValues(t, 1, deployTrace.Process.ExitCode)

	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// deployTrace := findFirstTraceByName(calls, "deploy")
	// require.NotNil(t, deployTrace, "deploy step should still be traced")
	// assert.True(t, deployTrace.Process.Skipped)
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

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "notify-failure"), "OnFailure step should have started")

	last := findLastTraceByName(getTracerStates(tracer), "notify-failure")
	require.NotNil(t, last)
	assert.True(t, last.Process.Exited, "notify-failure should have exited")
	assert.Equal(t, 0, last.Process.ExitCode, "notify-failure step itself should succeed")
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

	assert.NoError(t, err)
	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// cleanupTrace := findFirstTraceByName(getTracerStates(tracer), "cleanup-on-fail")
	// require.NotNil(t, cleanupTrace, "cleanup step should be traced even when skipped")
	// assert.True(t, cleanupTrace.Process.Skipped)
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
	assert.True(t, last.Process.Exited)
	assert.Equal(t, 0, last.Process.ExitCode)
}

func TestWorkflowFailureIgnoreDoesNotSetPipelineError(t *testing.T) {
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
	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// traces := getTracerStates(tracer)
	// for _, c := range traces {
	// 	if c.Pipeline.Step != nil && c.Pipeline.Step.Name == "deploy" {
	// 		assert.False(t, c.Process.Skipped, "deploy should not be skipped after failure=ignore step")
	// 	}
	// }
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
	assert.True(t, last.Process.Exited)
	assert.True(t, last.Process.OOMKilled)
	assert.Equal(t, 137, last.Process.ExitCode)
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
	assert.True(t, lastFast.Process.Exited)
	assert.Equal(t, 0, lastFast.Process.ExitCode, "test-fast should succeed")

	lastSlow := findLastTraceByName(getTracerStates(tracer), "test-slow")
	require.NotNil(t, lastSlow)
	assert.True(t, lastSlow.Process.Exited)
	assert.Equal(t, 1, lastSlow.Process.ExitCode, "test-slow should fail with code 1")
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
	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// assert.True(t, deployTrace.Process.Skipped)
}

// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
/*
func TestWorkflowContextCancelDuringExecution(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.Process.Exited && !s.Process.Skipped {
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
}.
*/

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

	deployTrace := findLastTraceByName(traces, "notify")
	require.NotNil(t, deployTrace)
	assert.True(t, deployTrace.Process.Exited, "notify should exited")
	assert.EqualValues(t, 0, deployTrace.Process.ExitCode, "notify should be successful")

	lastBuild := findLastTraceByName(traces, "lint")
	require.NotNil(t, lastBuild)
	assert.True(t, lastBuild.Process.Exited)
	assert.Equal(t, 1, lastBuild.Process.ExitCode, "lint should have failed")

	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// deployTrace := findFirstTraceByName(traces, "deploy")
	// require.NotNil(t, deployTrace)
	// assert.True(t, deployTrace.Process.Skipped, "deploy should be skipped after lint failure")

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

	notifyTrace := findFirstTraceByName(traces, "build")
	require.NotNil(t, notifyTrace)
	// TODO: signal skipped back (https://github.com/woodpecker-ci/woodpecker/pull/6166)
	// assert.True(t, notifyTrace.Process.Skipped,		"OnFailure step should be skipped when prior failure was ignored")

	assert.NotNil(t, findStartedTrace(traces, "build"),
		"build should run after ignored failure")
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
