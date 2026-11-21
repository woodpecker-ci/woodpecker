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

package finalizer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	log_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/log/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

type fixture struct {
	store store.Store
	logs  *log_mocks.MockService
	now   time.Time
	repo  *model.Repo
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	s := datastore.NewTestStore(t)
	repo := &model.Repo{ForgeRemoteID: "1", FullName: "o/r", Owner: "o", Name: "r"}
	require.NoError(t, s.CreateRepo(repo))
	return &fixture{store: s, logs: log_mocks.NewMockService(t), now: time.Now(), repo: repo}
}

// ago returns the unix time d before now.
func (f *fixture) ago(d time.Duration) int64 {
	return f.now.Add(-d).Unix()
}

// pipeline creates a pipeline that finished at the given unix time, with one step per name.
// Steps end in the state of the pipeline; step names are unique per test and double as step UUID.
// A step named "never-started" was skipped.
func (f *fixture) pipeline(t *testing.T, status model.StatusValue, finished int64, names ...string) {
	t.Helper()
	steps := make([]*model.Step, 0, len(names))
	for i, name := range names {
		step := &model.Step{Name: name, UUID: name, PID: i + 1, State: status, Started: 1, Finished: finished}
		switch {
		case name == "never-started":
			step.State, step.Started, step.Finished = model.StatusSkipped, 0, 0
		case status == model.StatusRunning:
			step.Finished = 0
		}
		steps = append(steps, step)
	}
	require.NoError(t, f.store.CreatePipeline(&model.Pipeline{RepoID: f.repo.ID, Status: status, Finished: finished}, steps...))
}

func (f *fixture) finalized(t *testing.T, name string) bool {
	t.Helper()
	step, err := f.store.StepByUUID(name)
	require.NoError(t, err)
	return step.LogState == model.LogStateFinalized
}

// state returns the log state of the step with the given name.
func (f *fixture) state(t *testing.T, name string) model.LogState {
	t.Helper()
	step, err := f.store.StepByUUID(name)
	require.NoError(t, err)
	return step.LogState
}

// recover hands back a claimed step log like RecoverClaims does once its delay is over.
func (f *fixture) recover(t *testing.T, name string) {
	t.Helper()
	step, err := f.store.StepByUUID(name)
	require.NoError(t, err)
	claimed, err := f.store.StepLogStateSwap(t.Context(), step, model.LogStateOpen)
	require.NoError(t, err)
	require.True(t, claimed)
}

func (f *fixture) finalizer() *Finalizer {
	return &Finalizer{Store: f.store, Logs: f.logs, Interval: time.Minute, Batch: 10}
}

// hung behaves like a log store that never answers but gives up once ctx is done.
func hung(ctx context.Context, _ *model.Step) error {
	<-ctx.Done()
	return context.Cause(ctx)
}

func TestTickFinalizesDueSteps(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "due", "never-started")
	// canceled pipelines may lack a finish time, their steps are due anyway
	f.pipeline(t, model.StatusKilled, 0, "killed")
	// inside the window the agent may still send lines
	f.pipeline(t, model.StatusSuccess, f.ago(time.Minute), "recent")
	f.pipeline(t, model.StatusRunning, f.ago(time.Hour), "running")

	var called []string
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		called = append(called, step.Name)
		return nil
	})

	fin := f.finalizer()
	res, err := fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 3}, res)
	// a step that never started has no log, it is only marked
	assert.Equal(t, []string{"due", "killed"}, called)
	for _, name := range []string{"due", "never-started", "killed"} {
		assert.True(t, f.finalized(t, name), name)
	}
	assert.False(t, f.finalized(t, "recent"))
	assert.False(t, f.finalized(t, "running"))

	// a finalized step is not passed again
	res, err = fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{}, res)
	assert.Equal(t, []string{"due", "killed"}, called)
}

// TestTickCutoff pins the cutoff: the late-log window, to the second.
func TestTickCutoff(t *testing.T) {
	f := newFixture(t)
	// an append admitted at the end of the late-log window may take AppendTimeout
	window := service_log.StreamDelayAllowed + service_log.AppendTimeout
	f.pipeline(t, model.StatusSuccess, f.ago(window), "at-cutoff")
	f.pipeline(t, model.StatusSuccess, f.ago(window+time.Second), "past-cutoff")
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(nil).Once()

	res, err := f.finalizer().Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 1}, res)
	assert.False(t, f.finalized(t, "at-cutoff"))
	assert.True(t, f.finalized(t, "past-cutoff"))
}

// TestTickRetriesFailedSteps ensures a step whose log store call failed stays claimed,
// so ticks skip it until RecoverClaims hands it back, and it is retried then.
func TestTickRetriesFailedSteps(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusFailure, f.ago(time.Hour), "ok", "bad")

	fail := true
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		if step.Name == "bad" && fail {
			return errors.New("store down")
		}
		return nil
	})

	fin := f.finalizer()
	res, err := fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 1, Failed: 1}, res)
	assert.True(t, f.finalized(t, "ok"))
	assert.Equal(t, model.LogStateFinalizing, f.state(t, "bad"), "stays claimed")

	fail = false
	res, err = fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{}, res, "a claimed step is skipped")

	f.recover(t, "bad")
	res, err = fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 1}, res)
	assert.True(t, f.finalized(t, "bad"))
}

// TestTickFailingStepsDoNotBlock ensures steps that keep failing cannot hold back newer pipelines:
// a failed step stays claimed, so the next tick takes the next pipeline.
func TestTickFailingStepsDoNotBlock(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusFailure, f.ago(time.Hour), "bad")
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "good")

	var called []string
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		called = append(called, step.Name)
		if step.Name == "bad" {
			return errors.New("log gone")
		}
		return nil
	})

	fin := f.finalizer()
	fin.Batch = 1
	for range 3 {
		_, err := fin.Tick(t.Context(), f.now)
		require.NoError(t, err)
	}
	assert.Equal(t, []string{"bad", "good"}, called)
	assert.True(t, f.finalized(t, "good"))
	assert.False(t, f.finalized(t, "bad"))
}

// TestTickRetriesUnderContinuousArrivals ensures a full batch of newer pipelines due every interval
// neither blocks behind a failed step nor keeps it from being retried once it is handed back.
func TestTickRetriesUnderContinuousArrivals(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusFailure, f.ago(time.Hour), "bad")

	down := true
	badCalls := 0
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		if step.Name != "bad" {
			return nil
		}
		badCalls++
		if down {
			return errors.New("store down")
		}
		return nil
	})

	fin := f.finalizer()
	fin.Batch = 1
	tick := func(name string) {
		t.Helper()
		f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), name)
		_, err := fin.Tick(t.Context(), f.now)
		require.NoError(t, err)
	}
	for i := range 5 {
		tick(fmt.Sprintf("new-%d", i))
		down = false
	}
	assert.Equal(t, 1, badCalls)
	for i := range 4 {
		assert.True(t, f.finalized(t, fmt.Sprintf("new-%d", i)), "new-%d", i)
	}

	// handed back, the oldest pipeline comes first again
	f.recover(t, "bad")
	tick("new-5")
	assert.Equal(t, 2, badCalls)
	assert.True(t, f.finalized(t, "bad"))
}

// TestTickConcurrent ensures concurrent finalizers sharing the database hand every
// step to the log store once: each claims a step before it works on it.
func TestTickConcurrent(t *testing.T) {
	f := newFixture(t)
	for i := range 5 {
		f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), fmt.Sprintf("a%d", i), fmt.Sprintf("b%d", i))
	}
	var mu sync.Mutex
	calls := make(map[string]int)
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		mu.Lock()
		calls[step.Name]++
		mu.Unlock()
		time.Sleep(time.Millisecond) // let the finalizers overlap
		return nil
	})

	results := make(chan Result)
	for range 3 {
		go func() {
			res, _ := f.finalizer().Tick(t.Context(), f.now)
			results <- res
		}()
	}
	done := 0
	for range 3 {
		done += (<-results).Done
	}
	assert.Equal(t, 10, done)
	assert.Len(t, calls, 10)
	for name, n := range calls {
		assert.Equal(t, 1, n, "step %s", name)
	}
}

// TestRecoverClaims ensures claims a crashed server left behind are reopened after the delay,
// but not a claim whose state changed meanwhile.
func TestRecoverClaims(t *testing.T) {
	oldDelay, oldJitter := recoverDelay, recoverJitter
	recoverDelay, recoverJitter = 100*time.Millisecond, 1
	t.Cleanup(func() { recoverDelay, recoverJitter = oldDelay, oldJitter })

	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "left", "taken-again")
	claim := func(name string, to model.LogState) *model.Step {
		step, err := f.store.StepByUUID(name)
		require.NoError(t, err)
		claimed, err := f.store.StepLogStateSwap(t.Context(), step, to)
		require.NoError(t, err)
		require.True(t, claimed)
		return step
	}
	claim("left", model.LogStateFinalizing)
	retaken := claim("taken-again", model.LogStateDeleting)

	ctx, cancel := context.WithCancelCause(t.Context())
	stopped := make(chan struct{})
	go func() {
		RecoverClaims(ctx, f.store)
		close(stopped)
	}()
	defer func() {
		cancel(nil)
		<-stopped
	}()
	time.Sleep(recoverDelay / 2) // the claims are listed
	// the delete finished and a finalizer claimed the log again
	claimed, err := f.store.StepLogStateSwap(t.Context(), retaken, model.LogStateOpen)
	require.NoError(t, err)
	require.True(t, claimed)
	claim("taken-again", model.LogStateFinalizing)

	state := func(name string) model.LogState {
		step, err := f.store.StepByUUID(name)
		require.NoError(t, err)
		return step.LogState
	}
	assert.Eventually(t, func() bool { return state("left") == model.LogStateOpen }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, model.LogStateFinalizing, state("taken-again"))
}

// TestTickGivesUpOnHungStore ensures one hung call cannot stall the finalizer: it is given up
// after stepTimeout and newer pipelines are still finalized.
func TestTickGivesUpOnHungStore(t *testing.T) {
	old := stepTimeout
	stepTimeout = 50 * time.Millisecond
	t.Cleanup(func() { stepTimeout = old })

	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "hung")
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "next")
	f.logs.On("StepFinished", mock.Anything, mock.MatchedBy(func(s *model.Step) bool { return s.Name == "hung" })).Return(hung)
	f.logs.On("StepFinished", mock.Anything, mock.MatchedBy(func(s *model.Step) bool { return s.Name == "next" })).Return(nil)

	// without the timeout the tick would wait for this deadline and finalize nothing
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	res, err := f.finalizer().Tick(ctx, f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 1, Failed: 1}, res)
	assert.True(t, f.finalized(t, "next"))
	assert.Equal(t, model.LogStateFinalizing, f.state(t, "hung"), "stays claimed for RecoverClaims")
}

// mockStore returns a store mock that lists the given steps of pipeline 1 as due.
func mockStore(t *testing.T, steps ...*model.Step) *store_mocks.MockStore {
	t.Helper()
	s := store_mocks.NewMockStore(t)
	s.On("StepCountByLogState", mock.Anything, mock.Anything).Return(int64(len(steps)), nil)
	s.On("StepListByLogState", mock.Anything, mock.MatchedBy(func(f *model.StepLogFilter) bool { return f.Limit == 10 })).Return(steps, nil)
	return s
}

// swap matches StepLogStateSwap calls of the step with the given ID to the given state.
func swap(s *store_mocks.MockStore, stepID int64, to model.LogState) *mock.Call {
	return s.On("StepLogStateSwap", mock.Anything, mock.MatchedBy(func(step *model.Step) bool { return step.ID == stepID }), to)
}

func TestTickMarkFailed(t *testing.T) {
	s := mockStore(t, &model.Step{ID: 1, UUID: "a", PipelineID: 1, Started: 1}, &model.Step{ID: 2, UUID: "b", PipelineID: 1, Started: 1})
	logs := log_mocks.NewMockService(t)
	logs.On("StepFinished", mock.Anything, mock.Anything).Return(nil)
	swap(s, 1, model.LogStateFinalizing).Return(true, nil)
	swap(s, 2, model.LogStateFinalizing).Return(true, nil)
	swap(s, 1, model.LogStateFinalized).Return(false, errors.New("db down"))
	swap(s, 2, model.LogStateFinalized).Return(true, nil)

	fin := &Finalizer{Store: s, Logs: logs, Interval: time.Minute, Batch: 10}
	res, err := fin.Tick(t.Context(), time.Now())
	require.NoError(t, err)
	assert.Equal(t, Result{Done: 1, Failed: 1}, res)
}

// TestTickGivesUpOnHungDatabase ensures marking a step is bound by stepTimeout like the log store call.
func TestTickGivesUpOnHungDatabase(t *testing.T) {
	old := stepTimeout
	stepTimeout = 50 * time.Millisecond
	t.Cleanup(func() { stepTimeout = old })

	s := mockStore(t, &model.Step{ID: 1, UUID: "a", PipelineID: 1, Started: 1})
	logs := log_mocks.NewMockService(t)
	logs.On("StepFinished", mock.Anything, mock.Anything).Return(nil)
	swap(s, 1, model.LogStateFinalizing).Return(true, nil)
	swap(s, 1, model.LogStateFinalized).Return(func(ctx context.Context, _ *model.Step, _ model.LogState) (bool, error) {
		// a database that never answers but gives up with ctx
		select {
		case <-ctx.Done():
			return false, context.Cause(ctx)
		case <-time.After(5 * time.Second):
			t.Error("marking the step ignored the step timeout")
			return false, errors.New("hung")
		}
	})

	fin := &Finalizer{Store: s, Logs: logs, Interval: time.Minute, Batch: 10}
	res, err := fin.Tick(t.Context(), time.Now())
	require.NoError(t, err)
	assert.Equal(t, Result{Failed: 1}, res)
}

// TestTickStopsWhileDatabaseHangs ensures a tick can be canceled while it waits for the database.
func TestTickStopsWhileDatabaseHangs(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	querying := make(chan struct{})
	s.On("StepCountByLogState", mock.Anything, mock.Anything).Return(func(ctx context.Context, _ *model.StepLogFilter) (int64, error) {
		close(querying)
		<-ctx.Done()
		return 0, context.Cause(ctx)
	})

	ctx, cancel := context.WithCancelCause(t.Context())
	errc := make(chan error, 1)
	go func() {
		_, err := (&Finalizer{Store: s, Interval: time.Minute, Batch: 10}).Tick(ctx, time.Now())
		errc <- err
	}()
	<-querying
	cancel(nil)
	select {
	case err := <-errc:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("tick still waiting for the database after cancel")
	}
}

func TestTickPaused(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "due")

	fin := f.finalizer()
	fin.Paused = true
	res, err := fin.Tick(t.Context(), f.now)
	require.NoError(t, err)
	assert.Equal(t, Result{}, res)
	assert.False(t, f.finalized(t, "due"))
}

func TestTickStopsOnCanceledContext(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "a", "b", "c")

	ctx, cancel := context.WithCancelCause(t.Context())
	var called []string
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(_ context.Context, step *model.Step) error {
		called = append(called, step.Name)
		if step.Name == "b" {
			cancel(nil) // the server shuts down
		}
		return nil
	})

	res, err := f.finalizer().Tick(ctx, f.now)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []string{"a", "b"}, called)
	// what was done before the cancel is kept, the interrupted step is handed over again later
	assert.Equal(t, Result{Done: 1, Failed: 1}, res)
	assert.True(t, f.finalized(t, "a"))
	assert.False(t, f.finalized(t, "b"))
	assert.False(t, f.finalized(t, "c"))
}

// TestRunStopsWithHungStore ensures the server can shut down while a log store call hangs.
func TestRunStopsWithHungStore(t *testing.T) {
	f := newFixture(t)
	f.pipeline(t, model.StatusSuccess, f.ago(time.Hour), "hung")
	called := make(chan struct{})
	f.logs.On("StepFinished", mock.Anything, mock.Anything).Return(func(ctx context.Context, step *model.Step) error {
		close(called)
		return hung(ctx, step)
	}).Once()

	// the interval is at least a minute in production, the test cannot wait that long
	defer func(interval time.Duration) { minInterval = interval }(minInterval)
	minInterval = time.Millisecond
	fin := f.finalizer()
	fin.Interval = time.Millisecond
	ctx, cancel := context.WithCancelCause(t.Context())
	ran := make(chan error)
	go func() { ran <- fin.Run(ctx) }()
	<-called
	cancel(nil)
	select {
	case err := <-ran:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Run still blocked after cancel")
	}
}

func TestRunRejectsInvalidConfig(t *testing.T) {
	for name, fin := range map[string]*Finalizer{
		"interval below 1m": {Interval: time.Minute - time.Second, Batch: 1},
		"zero interval":     {Interval: 0, Batch: 1},
		"negative interval": {Interval: -time.Second, Batch: 1},
		"zero batch":        {Interval: time.Minute, Batch: 0},
		"negative batch":    {Interval: time.Minute, Batch: -1},
	} {
		t.Run(name, func(t *testing.T) {
			// without the check Run would loop until the context ends and return nil
			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()
			assert.Error(t, fin.Run(ctx))
		})
	}
}
