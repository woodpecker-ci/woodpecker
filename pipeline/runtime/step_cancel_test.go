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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

// transportCancelErr mimics what a backend using an HTTP client returns when
// its request context is canceled: net/http surfaces context.Cause(ctx), which
// is the cancel cause the agent set, not context.Canceled.
func transportCancelErr(cause error) error {
	return fmt.Errorf(`error during connect: Post "http://.../containers/x/wait": %w`, cause)
}

func canceledRuntime(t *testing.T, engine backend_types.Backend) *Runtime {
	t.Helper()
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(pipeline_errors.ErrCancel)
	return New(&backend_types.Config{}, engine,
		WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)), WithContext(ctx))
}

// Backend errors raised because the workflow was canceled describe the aborted
// call, not a step failure. They must be reported as a plain cancellation.
func TestCompleteStepCancelFallout(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		waitErr error
	}{
		{"TransportErrorWrappingCancelCause", transportCancelErr(pipeline_errors.ErrCancel)},
		{"TransportErrorWrappingContextCanceled", transportCancelErr(context.Canceled)},
		{"BareContextCanceled", context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			engine := mocks.NewMockBackend(t)
			engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).Return(nil, tc.waitErr)
			engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			r := canceledRuntime(t, engine)

			ws, err := r.completeStep(t.Context(), dummyStep("s1"), func() {}, time.Now().Unix())

			assert.NoError(t, err)
			require.NotNil(t, ws)
			assert.Equal(t, pipeline_errors.ErrCancel, ws.Error,
				"a backend error caused by the cancellation must not leak into the step state")
		})
	}

	t.Run("UnrelatedErrorStillSurfaces", func(t *testing.T) {
		t.Parallel()
		engine := mocks.NewMockBackend(t)
		engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("engine exploded"))
		r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))

		ws, err := r.completeStep(t.Context(), dummyStep("s1"), func() {}, time.Now().Unix())

		assert.EqualError(t, err, "engine exploded")
		assert.Nil(t, ws)
	})
}

// A step that could not even be started because the workflow was canceled must
// be reported as canceled, not with the backend's transport error.
func TestRunBlockingStepStartFailureAfterCancel(t *testing.T) {
	t.Parallel()

	engine := mocks.NewMockBackend(t)
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).
		Return(transportCancelErr(pipeline_errors.ErrCancel))
	r := canceledRuntime(t, engine)

	err := r.runBlockingStep(t.Context(), dummyStep("s1"))

	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
	assert.Equal(t, pipeline_errors.ErrCancel, err,
		"the transport error must be replaced, not just wrapped")
}

// A start failure that has nothing to do with cancellation must still surface.
func TestRunBlockingStepStartFailureWithoutCancel(t *testing.T) {
	t.Parallel()

	engine := mocks.NewMockBackend(t)
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("no such image"))
	r := New(&backend_types.Config{}, engine, WithTracer(newTestTracer(t)), WithLogger(newTestLogger(t)))

	err := r.runBlockingStep(t.Context(), dummyStep("s1"))

	assert.EqualError(t, err, "no such image")
}

// The error a canceled step ends with must not reach the stage as a step
// failure, whatever shape the backend gave it.
func TestRunBlockingStepCompleteFalloutAfterCancel(t *testing.T) {
	t.Parallel()

	engine := mocks.NewMockBackend(t)
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	engine.On("TailStep", mock.Anything, mock.Anything, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)
	engine.On("WaitStep", mock.Anything, mock.Anything, mock.Anything).
		Return(&backend_types.State{Exited: true, ExitCode: 130}, nil)
	engine.On("DestroyStep", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	r := canceledRuntime(t, engine)

	err := r.runBlockingStep(t.Context(), dummyStep("s1"))

	assert.ErrorIs(t, err, pipeline_errors.ErrCancel)
	assert.False(t, pipeline_errors.IsStepFailure(err),
		"a step aborted by the cancellation is not a step failure")
}
