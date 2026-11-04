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

package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// fakeContainerAPI is an in-memory stand-in for the container calls used by
// StartStep/WaitStep/DestroyStep. It needs no Docker daemon.
type fakeContainerAPI struct {
	client.APIClient // unused methods panic on purpose

	removed  bool
	exitCode int
	inspects int
}

func (f *fakeContainerAPI) notFound() error {
	// Podman's lowercase wording, which the string checks do not match.
	return errdefs.ErrNotFound.WithMessage(`no container with name or ID "postgres" found: no such container`)
}

func (f *fakeContainerAPI) ContainerCreate(context.Context, client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	return client.ContainerCreateResult{}, nil
}

func (f *fakeContainerAPI) ContainerStart(context.Context, string, client.ContainerStartOptions) (client.ContainerStartResult, error) {
	return client.ContainerStartResult{}, nil
}

func (f *fakeContainerAPI) ContainerStop(context.Context, string, client.ContainerStopOptions) (client.ContainerStopResult, error) {
	if f.removed {
		return client.ContainerStopResult{}, f.notFound()
	}
	return client.ContainerStopResult{}, nil
}

func (f *fakeContainerAPI) ContainerKill(context.Context, string, client.ContainerKillOptions) (client.ContainerKillResult, error) {
	if f.removed {
		return client.ContainerKillResult{}, f.notFound()
	}
	return client.ContainerKillResult{}, errors.New("is not running")
}

func (f *fakeContainerAPI) ContainerRemove(context.Context, string, client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	if f.removed {
		return client.ContainerRemoveResult{}, f.notFound()
	}
	f.removed = true
	return client.ContainerRemoveResult{}, nil
}

func (f *fakeContainerAPI) ContainerInspect(context.Context, string, client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	f.inspects++
	if f.removed {
		return client.ContainerInspectResult{}, f.notFound()
	}
	return client.ContainerInspectResult{Container: container.InspectResponse{
		State: &container.State{ExitCode: f.exitCode},
	}}, nil
}

func (f *fakeContainerAPI) ContainerWait(context.Context, string, client.ContainerWaitOptions) client.ContainerWaitResult {
	resultC := make(chan container.WaitResponse, 1)
	errC := make(chan error, 1)
	if f.removed {
		errC <- f.notFound()
	} else {
		resultC <- container.WaitResponse{StatusCode: int64(f.exitCode)}
	}
	return client.ContainerWaitResult{Result: resultC, Error: errC}
}

func newWaitTestStep() *backend_types.Step {
	return &backend_types.Step{Name: "postgres", UUID: "uuid-1", Type: backend_types.StepTypeService, Detached: true}
}

// A detached service is stopped and removed by the workflow teardown before its
// background goroutine calls WaitStep. It must still be reported as exited.
func TestWaitStepAfterDestroyStep(t *testing.T) {
	ctx := t.Context()
	step := newWaitTestStep()
	api := &fakeContainerAPI{}
	e := &docker{client: api}

	require.NoError(t, e.StartStep(ctx, step, "task"))
	require.NoError(t, e.DestroyStep(ctx, step, "task"))
	require.True(t, api.removed)

	state, err := e.WaitStep(ctx, step, "task")
	require.NoError(t, err)
	assert.True(t, state.Exited)
	assert.Equal(t, 0, state.ExitCode)

	// nothing is kept once WaitStep has returned
	assert.Empty(t, e.stepStates)
}

// completeStep destroys a step again after WaitStep returned. The container is
// already gone, and the engine answers with a not-found error, which must be
// ignored.
func TestDestroyStepTwiceAroundWaitStep(t *testing.T) {
	ctx := t.Context()
	step := newWaitTestStep()
	api := &fakeContainerAPI{}
	e := &docker{client: api}

	require.NoError(t, e.StartStep(ctx, step, "task"))
	require.NoError(t, e.DestroyStep(ctx, step, "task"))

	state, err := e.WaitStep(ctx, step, "task")
	require.NoError(t, err)
	assert.True(t, state.Exited)

	require.NoError(t, e.DestroyStep(ctx, step, "task"))
}

func TestWaitStepAfterDestroyStepKeepsExitCode(t *testing.T) {
	ctx := t.Context()
	step := newWaitTestStep()
	api := &fakeContainerAPI{exitCode: 3}
	e := &docker{client: api}

	require.NoError(t, e.StartStep(ctx, step, "task"))
	require.NoError(t, e.DestroyStep(ctx, step, "task"))

	state, err := e.WaitStep(ctx, step, "task")
	require.NoError(t, err)
	assert.Equal(t, 3, state.ExitCode)
}

// A step that was waited on before it is destroyed (the normal case) must not
// cost an extra inspect call or leave anything behind.
func TestDestroyStepAfterWaitStepRecordsNothing(t *testing.T) {
	ctx := t.Context()
	step := newWaitTestStep()
	api := &fakeContainerAPI{}
	e := &docker{client: api}

	require.NoError(t, e.StartStep(ctx, step, "task"))
	_, err := e.WaitStep(ctx, step, "task")
	require.NoError(t, err)
	inspectsAfterWait := api.inspects

	require.NoError(t, e.DestroyStep(ctx, step, "task"))
	assert.Equal(t, inspectsAfterWait, api.inspects)
	assert.Empty(t, e.stepStates)
}

// A container that disappeared without DestroyStep recording it is still an error.
func TestWaitStepNotFoundWithoutRecordedState(t *testing.T) {
	step := newWaitTestStep()
	e := &docker{client: &fakeContainerAPI{removed: true}}

	_, err := e.WaitStep(t.Context(), step, "task")
	require.Error(t, err)
	assert.True(t, errdefs.IsNotFound(err))
}
