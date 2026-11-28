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

package datastore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// addPipeline creates a pipeline with its steps.
func addPipeline(t *testing.T, s *storage, number int64, status model.StatusValue, finished int64, steps ...*model.Step) {
	t.Helper()
	p := &model.Pipeline{RepoID: 1, Status: status, Finished: finished, Number: number}
	_, err := s.engine.Insert(p)
	require.NoError(t, err)
	for i, step := range steps {
		step.PipelineID = p.ID
		step.PID = i + 1
		step.UUID = step.Name
	}
	sess := s.engine.NewSession()
	require.NoError(t, s.stepCreate(sess, steps))
	require.NoError(t, sess.Commit())
}

func stepNames(steps []*model.Step) []string {
	names := make([]string, 0, len(steps))
	for _, s := range steps {
		names = append(names, s.Name)
	}
	return names
}

func TestStepListByLogState(t *testing.T) {
	s, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	const (
		before = int64(1000)
		old    = int64(500)
		recent = int64(2000)
	)

	// oldest pipeline first
	addPipeline(t, s, 1, model.StatusSuccess, old,
		&model.Step{Name: "due", State: model.StatusSuccess, Started: 10, Finished: old},
		// listed too, the finalizer marks it without calling the log store
		&model.Step{Name: "never-started", State: model.StatusSkipped},
		&model.Step{Name: "done", State: model.StatusSuccess, Started: 10, Finished: old, LogState: model.LogStateFinalized},
		&model.Step{Name: "claimed", State: model.StatusSuccess, Started: 10, Finished: old, LogState: model.LogStateFinalizing},
		&model.Step{Name: "step-too-recent", State: model.StatusSuccess, Started: 10, Finished: recent},
		&model.Step{Name: "step-at-cutoff", State: model.StatusSuccess, Started: 10, Finished: before},
	)
	for i, status := range []model.StatusValue{model.StatusRunning, model.StatusPending, model.StatusBlocked, model.StatusCreated} {
		addPipeline(t, s, int64(10+i), status, old, &model.Step{Name: "pipeline-" + string(status), State: model.StatusSuccess, Started: 10, Finished: old})
	}
	addPipeline(t, s, 20, model.StatusFailure, recent, &model.Step{Name: "pipeline-too-recent", State: model.StatusFailure, Started: 10, Finished: old})
	addPipeline(t, s, 21, model.StatusFailure, before, &model.Step{Name: "pipeline-at-cutoff", State: model.StatusFailure, Started: 10, Finished: old})
	// a canceled pipeline is killed at once, its running steps until the agent reports back:
	// the agent may still send lines (see allowAppendingLogs)
	addPipeline(t, s, 30, model.StatusKilled, old, &model.Step{Name: "step-still-running", State: model.StatusRunning, Started: 10})
	addPipeline(t, s, 31, model.StatusFailure, before-1, &model.Step{Name: "second-pipeline", State: model.StatusFailure, Started: 10, Finished: before - 1})

	// due selects the open logs that are complete, as the log finalizer does
	due := func(limit int) *model.StepLogFilter {
		return &model.StepLogFilter{States: []model.LogState{model.LogStateOpen}, CompleteBefore: before, Limit: limit}
	}

	t.Run("selects only complete logs", func(t *testing.T) {
		got, err := s.StepListByLogState(t.Context(), due(0))
		require.NoError(t, err)
		assert.Equal(t, []string{"due", "never-started", "second-pipeline"}, stepNames(got))
	})

	t.Run("limit counts steps, oldest pipeline first", func(t *testing.T) {
		got, err := s.StepListByLogState(t.Context(), due(1))
		require.NoError(t, err)
		assert.Equal(t, []string{"due"}, stepNames(got))

		got, err = s.StepListByLogState(t.Context(), due(2))
		require.NoError(t, err)
		assert.Equal(t, []string{"due", "never-started"}, stepNames(got))
	})

	t.Run("count ignores the limit", func(t *testing.T) {
		n, err := s.StepCountByLogState(t.Context(), due(1))
		require.NoError(t, err)
		assert.Equal(t, int64(3), n)
	})

	t.Run("states without cutoff", func(t *testing.T) {
		got, err := s.StepListByLogState(t.Context(), &model.StepLogFilter{
			States: []model.LogState{model.LogStateFinalizing, model.LogStateFinalized},
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"done", "claimed"}, stepNames(got))
	})

	t.Run("no states selects any", func(t *testing.T) {
		n, err := s.StepCountByLogState(t.Context(), &model.StepLogFilter{})
		require.NoError(t, err)
		assert.Equal(t, int64(14), n, "every step")
	})
}

func TestStepCountByLogStateEmpty(t *testing.T) {
	s, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	n, err := s.StepCountByLogState(t.Context(), &model.StepLogFilter{States: []model.LogState{model.LogStateOpen}, CompleteBefore: 1000})
	require.NoError(t, err)
	assert.Zero(t, n)
}

// TestStepLogStateSwap ensures a log state change is one atomic claim: it only happens if the
// stored state is still the one of the given step, so only one concurrent call wins.
func TestStepLogStateSwap(t *testing.T) {
	s, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()

	addPipeline(t, s, 1, model.StatusSuccess, 10,
		&model.Step{Name: "a", State: model.StatusFailure, Started: 1, Finished: 10, ExitCode: 3},
		&model.Step{Name: "b", State: model.StatusSuccess, Started: 1, Finished: 10},
	)
	callA, err := s.StepByUUID("a")
	require.NoError(t, err)
	callB, err := s.StepByUUID("a")
	require.NoError(t, err)
	assert.Equal(t, model.LogStateOpen, callA.LogState, "new steps are open")

	claimed, err := s.StepLogStateSwap(t.Context(), callA, model.LogStateFinalizing)
	require.NoError(t, err)
	assert.True(t, claimed)
	assert.Equal(t, model.LogStateFinalizing, callA.LogState, "the step is updated")
	claimed, err = s.StepLogStateSwap(t.Context(), callB, model.LogStateFinalizing)
	require.NoError(t, err)
	assert.False(t, claimed, "the second claim of the same state loses")

	listed, err := s.StepListByLogState(t.Context(), &model.StepLogFilter{
		States: []model.LogState{model.LogStateFinalizing, model.LogStateDeleting},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"a"}, stepNames(listed))

	claimed, err = s.StepLogStateSwap(t.Context(), callA, model.LogStateFinalized)
	require.NoError(t, err)
	require.True(t, claimed)
	stored, err := s.StepByUUID("a")
	require.NoError(t, err)
	assert.Equal(t, model.LogStateFinalized, stored.LogState)
	assert.Equal(t, 3, stored.ExitCode, "only the log state is written")

	list, err := s.StepListByLogState(t.Context(), &model.StepLogFilter{
		States: []model.LogState{model.LogStateOpen}, CompleteBefore: 100, Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"b"}, stepNames(list))
}

// TestStepLogCanceled ensures the log finalizer queries give up once their context is done,
// so a hung database cannot block the finalizer or the server shutdown.
func TestStepLogCanceled(t *testing.T) {
	s, closer := newTestStore(t, new(model.Step), new(model.Pipeline))
	defer closer()
	defer holdConnection(t, s)()

	assertCanceled(t, func(ctx context.Context) error {
		_, err := s.StepCountByLogState(ctx, &model.StepLogFilter{CompleteBefore: 1000})
		return err
	})
	assertCanceled(t, func(ctx context.Context) error {
		_, err := s.StepListByLogState(ctx, &model.StepLogFilter{CompleteBefore: 1000, Limit: 10})
		return err
	})
	assertCanceled(t, func(ctx context.Context) error {
		_, err := s.StepLogStateSwap(ctx, &model.Step{ID: 1}, model.LogStateFinalizing)
		return err
	})
}
