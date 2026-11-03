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

package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func waitAsync(r *PauseRegistry, ctx context.Context, workflowID int64, stepUUID string) <-chan [2]any {
	out := make(chan [2]any, 1)
	go func() {
		resume, err := r.Wait(ctx, workflowID, stepUUID)
		out <- [2]any{resume, err}
	}()
	return out
}

func TestPauseRegistry(t *testing.T) {
	t.Parallel()

	t.Run("resume wakes up waiting agent", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		res := waitAsync(r, t.Context(), 1, "step-a")
		time.Sleep(10 * time.Millisecond)

		assert.True(t, r.Decide(1, true))
		got := <-res
		assert.Equal(t, true, got[0])
		assert.Nil(t, got[1])
	})

	t.Run("decline wakes up waiting agent", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		res := waitAsync(r, t.Context(), 1, "step-a")
		time.Sleep(10 * time.Millisecond)

		assert.True(t, r.Decide(1, false))
		got := <-res
		assert.Equal(t, false, got[0])
		assert.Nil(t, got[1])
	})

	t.Run("decision while agent reconnects is kept", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		ctx, cancel := context.WithCancel(t.Context())
		res := waitAsync(r, ctx, 1, "step-a")
		time.Sleep(10 * time.Millisecond)
		cancel() // connection lost
		got := <-res
		assert.ErrorIs(t, got[1].(error), context.Canceled)

		assert.True(t, r.Decide(1, true))

		resume, err := r.Wait(t.Context(), 1, "step-a") // agent reconnected
		assert.NoError(t, err)
		assert.True(t, resume)
	})

	t.Run("decision without waiting agent (server restart) is kept", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		assert.True(t, r.Decide(7, false))

		resume, err := r.Wait(t.Context(), 7, "any-step")
		assert.NoError(t, err)
		assert.False(t, resume)
	})

	t.Run("second decision is rejected", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		assert.True(t, r.Decide(1, true))
		assert.False(t, r.Decide(1, false))
	})

	t.Run("decision is consumed once", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		require.True(t, r.Decide(1, true))
		resume, err := r.Wait(t.Context(), 1, "step-a")
		require.NoError(t, err)
		require.True(t, resume)

		// a later pause of the same workflow waits for a new decision.
		ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
		defer cancel()
		_, err = r.Wait(ctx, 1, "step-b")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("workflows are isolated", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		resA := waitAsync(r, t.Context(), 1, "step-a")
		resB := waitAsync(r, t.Context(), 2, "step-a")
		time.Sleep(10 * time.Millisecond)

		r.Decide(2, true)
		gotB := <-resB
		assert.Equal(t, true, gotB[0])
		select {
		case <-resA:
			t.Fatal("workflow 1 must still be paused")
		case <-time.After(20 * time.Millisecond):
		}
		r.Decide(1, false)
		gotA := <-resA
		assert.Equal(t, false, gotA[0])
	})

	t.Run("take returns an existing decision once", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		_, ok := r.Take(1, "step-a")
		assert.False(t, ok, "no decision yet")
		r.Decide(1, true)
		resume, ok := r.Take(1, "step-a")
		assert.True(t, ok)
		assert.True(t, resume)
		_, ok = r.Take(1, "step-a")
		assert.False(t, ok, "decision is consumed")
	})

	t.Run("forget drops a pending decision", func(t *testing.T) {
		t.Parallel()
		r := NewPauseRegistry()
		r.Decide(1, true)
		r.Forget(1)
		assert.True(t, r.Decide(1, false), "a new decision is possible after forget")
	})
}

func TestDecidePause(t *testing.T) {
	t.Run("not paused", func(t *testing.T) {
		_, err := DecidePause(store_mocks.NewMockStore(t), &model.Workflow{ID: 101, State: model.StatusRunning}, true)
		assert.ErrorIs(t, err, ErrWorkflowNotPaused)
	})

	t.Run("resume sets running", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("WorkflowUpdate", mock.MatchedBy(func(w *model.Workflow) bool {
			return w.ID == 102 && w.State == model.StatusRunning
		})).Return(nil).Once()
		t.Cleanup(func() { Pauses.Forget(102) })

		w, err := DecidePause(s, &model.Workflow{ID: 102, State: model.StatusPaused}, true)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusRunning, w.State)
	})

	t.Run("decline keeps paused until agent is done", func(t *testing.T) {
		t.Cleanup(func() { Pauses.Forget(103) })
		w, err := DecidePause(store_mocks.NewMockStore(t), &model.Workflow{ID: 103, State: model.StatusPaused}, false)
		assert.NoError(t, err)
		assert.Equal(t, model.StatusPaused, w.State)
	})

	t.Run("decided twice", func(t *testing.T) {
		t.Cleanup(func() { Pauses.Forget(104) })
		_, err := DecidePause(store_mocks.NewMockStore(t), &model.Workflow{ID: 104, State: model.StatusPaused}, false)
		require.NoError(t, err)
		_, err = DecidePause(store_mocks.NewMockStore(t), &model.Workflow{ID: 104, State: model.StatusPaused}, false)
		assert.ErrorIs(t, err, ErrWorkflowNotPaused)
	})
}

func TestPausedStatus(t *testing.T) {
	assert.NoError(t, model.StatusPaused.Validate())
	assert.True(t, (&model.Workflow{State: model.StatusPaused}).Running(), "a paused workflow keeps the pipeline running")
	assert.Equal(t, model.StatusPaused, MergeStatusValues(model.StatusSuccess, model.StatusPaused))
	assert.Equal(t, model.StatusPaused, MergeStatusValues(model.StatusPaused, model.StatusRunning))
	assert.Equal(t, model.StatusKilled, MergeStatusValues(model.StatusPaused, model.StatusKilled))
}
