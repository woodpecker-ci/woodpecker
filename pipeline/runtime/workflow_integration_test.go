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

func findTraceByName(traces []state.State, name string) *state.State {
	for i := range traces {
		if traces[i].Workflow.Step != nil && traces[i].Workflow.Step.Name == name {
			return &traces[i]
		}
	}
	return nil
}

func findStartedTrace(traces []state.State, name string) *state.State {
	for i := range traces {
		if traces[i].Workflow.Step != nil && traces[i].Workflow.Step.Name == name && !traces[i].CurrentStep.Exited {
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	calls := getTracerStates(tracer)
	assert.Len(t, calls, 6)
	for i := 0; i < 6; i += 2 {
		assert.False(t, calls[i].CurrentStep.Exited, "trace %d should be step-started", i)
		assert.True(t, calls[i+1].CurrentStep.Exited, "trace %d should be step-completed", i+1)
		assert.Equal(t, 0, calls[i+1].CurrentStep.ExitCode)
	}
}

func TestWorkflowWithServiceStep(t *testing.T) {
	t.Parallel()
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
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
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
		WithBackend(dummy.New()),
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	var exitErr *pipeline_errors.ExitError
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, 1, exitErr.Code)

	deployTrace := findTraceByName(getTracerStates(tracer), "deploy")
	require.NotNil(t, deployTrace, "deploy step should still be traced")
	assert.True(t, deployTrace.CurrentStep.Skipped)
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "notify-failure"),
		"OnFailure step should have started")
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	cleanupTrace := findTraceByName(getTracerStates(tracer), "cleanup-on-fail")
	require.NotNil(t, cleanupTrace, "cleanup step should be traced even when skipped")
	assert.True(t, cleanupTrace.CurrentStep.Skipped)
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err, "pipeline should succeed when failing step has failure=ignore")
	assert.NotNil(t, findStartedTrace(getTracerStates(tracer), "build"),
		"build step should run after ignored failure")
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	traces := getTracerStates(tracer)
	for _, c := range traces {
		if c.Workflow.Step != nil && c.Workflow.Step.Name == "deploy" {
			assert.False(t, c.CurrentStep.Skipped, "deploy should not be skipped after failure=ignore step")
		}
	}
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
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	assert.NoError(t, r.Run(t.Context()))
}

func TestWorkflowOOMKilledStep(t *testing.T) {
	t.Parallel()
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build", withOOM())}},
			},
		},
		WithBackend(dummy.New()),
		WithTracer(newTestTracer(t)),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	var oomErr *pipeline_errors.OomError
	assert.True(t, errors.As(err, &oomErr))
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
		WithBackend(dummy.New()),
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	assert.Len(t, getTracerStates(tracer), 4, "both parallel steps should complete and be traced")
}

func TestWorkflowStepStartFailure(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{
			Stages: []*backend.Stage{
				{Steps: []*backend.Step{cmdStep("build", withStartFail())}},
				{Steps: []*backend.Step{cmdStep("deploy")}},
			},
		},
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	deployTrace := findTraceByName(getTracerStates(tracer), "deploy")
	require.NotNil(t, deployTrace)
	assert.True(t, deployTrace.CurrentStep.Skipped)
}

func TestWorkflowContextCancelDuringExecution(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())

	var stageCount int
	tracer := tracer_mocks.NewMockTracer(t)
	tracer.On("Trace", mock.Anything).Run(func(args mock.Arguments) {
		s, _ := args.Get(0).(*state.State)
		if s.CurrentStep.Exited && !s.CurrentStep.Skipped {
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
		WithBackend(dummy.New()),
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
		WithBackend(dummy.New()),
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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.Error(t, err)
	traces := getTracerStates(tracer)

	deployTrace := findTraceByName(traces, "deploy")
	require.NotNil(t, deployTrace)
	assert.True(t, deployTrace.CurrentStep.Skipped, "deploy should be skipped after lint failure")

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
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	traces := getTracerStates(tracer)

	notifyTrace := findTraceByName(traces, "error-notify")
	require.NotNil(t, notifyTrace)
	assert.True(t, notifyTrace.CurrentStep.Skipped,
		"OnFailure step should be skipped when prior failure was ignored")

	assert.NotNil(t, findStartedTrace(traces, "build"),
		"build should run after ignored failure")
}

func TestWorkflowEmptyStages(t *testing.T) {
	t.Parallel()
	tracer := newTestTracer(t)
	r := New(
		&backend.Config{Stages: []*backend.Stage{}},
		WithBackend(dummy.New()),
		WithTracer(tracer),
		WithLogger(newTestLogger(t)),
	)

	err := r.Run(t.Context())

	assert.NoError(t, err)
	assert.Empty(t, getTracerStates(tracer))
}
