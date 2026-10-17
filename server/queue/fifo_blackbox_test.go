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

//go:build queue_test

package queue_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
)

// The tests in this file only use the exported Queue API, so they pin the
// scheduling behavior independent of how the in-memory queue implements it.

// settle is long enough for the queue to run several scheduling ticks.
const settle = 250 * time.Millisecond

func anyTask(*model.Task) (bool, int) { return true, 1 }

func only(id string, score int) func(*model.Task) (bool, int) {
	return func(task *model.Task) (bool, int) { return task.ID == id, score }
}

func newBlackboxQueue(t *testing.T) queue.Queue {
	t.Helper()
	ctx, cancel := context.WithCancelCause(t.Context())
	t.Cleanup(func() { cancel(nil) })
	return queue.NewMemoryQueue(ctx)
}

func task(id string, deps ...string) *model.Task {
	return &model.Task{ID: id, Name: id, Dependencies: deps, DepStatus: make(map[string]model.StatusValue)}
}

func member(id string, pipeline, created int64, group string, limit int, deps ...string) *model.Task {
	t := task(id, deps...)
	t.PipelineID = pipeline
	t.Created = created
	t.ConcurrencyGroup = group
	t.ConcurrencyLimit = limit
	return t
}

func push(t *testing.T, q queue.Queue, tasks ...*model.Task) {
	t.Helper()
	require.NoError(t, q.PushAtOnce(t.Context(), tasks))
}

// poll waits for the next task the filter accepts.
func poll(t *testing.T, q queue.Queue, agent int64, filter func(*model.Task) (bool, int)) *model.Task {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	got, err := q.Poll(ctx, agent, filter)
	require.NoError(t, err)
	return got
}

// pollNone asserts that the filter gets no task while the queue settles.
func pollNone(t *testing.T, q queue.Queue, agent int64, filter func(*model.Task) (bool, int)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), settle)
	defer cancel()
	got, err := q.Poll(ctx, agent, filter)
	require.ErrorIs(t, err, context.DeadlineExceeded, "unexpected task %v", got)
}

func TestBlackboxDependencies(t *testing.T) {
	t.Parallel()

	t.Run("independent tasks are served in push order", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"), task("b"), task("c"))
		for _, id := range []string{"a", "b", "c"} {
			assert.Equal(t, id, poll(t, q, 1, anyTask).ID)
		}
	})

	t.Run("chain runs one after another", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"), task("b", "a"), task("c", "b"))

		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		info := q.Info(t.Context())
		assert.Equal(t, 2, info.Stats.WaitingOnDeps)
		assert.Equal(t, 0, info.Stats.Pending)

		require.NoError(t, q.Done(t.Context(), "a", model.StatusSuccess))
		b := poll(t, q, 1, anyTask)
		assert.Equal(t, "b", b.ID)
		assert.Equal(t, model.StatusSuccess, b.DepStatus["a"])

		require.NoError(t, q.Done(t.Context(), "b", model.StatusSuccess))
		assert.Equal(t, "c", poll(t, q, 1, anyTask).ID)
	})

	t.Run("dependency pushed after its dependent", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("child", "parent"), task("parent"))

		assert.Equal(t, "parent", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		require.NoError(t, q.Done(t.Context(), "parent", model.StatusSuccess))
		assert.Equal(t, "child", poll(t, q, 1, anyTask).ID)
	})

	t.Run("fan-in waits for every dependency", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"), task("b"), task("c", "a", "b"))

		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
		assert.Equal(t, "b", poll(t, q, 1, anyTask).ID)
		require.NoError(t, q.Done(t.Context(), "a", model.StatusSuccess))
		pollNone(t, q, 1, anyTask)

		require.NoError(t, q.Error(t.Context(), "b", errors.New("exit code 1")))
		c := poll(t, q, 1, anyTask)
		assert.Equal(t, "c", c.ID)
		assert.Equal(t, map[string]model.StatusValue{"a": model.StatusSuccess, "b": model.StatusFailure}, c.DepStatus)
	})

	t.Run("waiting task blocks its own dependents", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"))
		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
		push(t, q, task("c", "b"), task("b", "a"))

		pollNone(t, q, 1, anyTask)
		assert.Equal(t, 2, q.Info(t.Context()).Stats.WaitingOnDeps)
		require.NoError(t, q.Done(t.Context(), "a", model.StatusSuccess))
		assert.Equal(t, "b", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
	})

	t.Run("dependency outside the queue does not block", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a", "finished-long-ago"))
		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
	})

	t.Run("canceled waiting task releases its dependents", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"), task("b", "a"), task("c", "b"))
		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)

		require.NoError(t, q.ErrorAtOnce(t.Context(), []string{"b"}, queue.ErrCancel))
		c := poll(t, q, 1, anyTask)
		assert.Equal(t, "c", c.ID)
		assert.Equal(t, model.StatusKilled, c.DepStatus["b"])
	})
}

func TestBlackboxWorkerMatching(t *testing.T) {
	t.Parallel()

	t.Run("task waits for a matching worker", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, task("a"))
		pollNone(t, q, 1, only("other", 1))
		assert.Equal(t, "a", poll(t, q, 2, only("a", 1)).ID)
	})

	t.Run("highest score wins", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		ctx, cancel := context.WithCancelCause(t.Context())
		defer cancel(nil)
		type result struct {
			agent int64
			task  *model.Task
		}
		results := make(chan result, 2)
		for agent, score := range map[int64]int{1: 1, 2: 5} {
			go func() {
				got, err := q.Poll(ctx, agent, only("a", score))
				if err == nil {
					results <- result{agent, got}
				}
			}()
		}
		require.Eventually(t, func() bool { return q.Info(t.Context()).Stats.Workers == 2 }, 5*time.Second, 10*time.Millisecond)

		push(t, q, task("a"))
		select {
		case r := <-results:
			assert.Equal(t, int64(2), r.agent)
			assert.Equal(t, "a", r.task.ID)
		case <-time.After(5 * time.Second):
			t.Fatal("task was not dispatched")
		}
	})

	t.Run("blocked task does not hand its worker to another task", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, member("x", 1, 1, "deploy", 1))
		assert.Equal(t, "x", poll(t, q, 1, anyTask).ID)

		push(t, q, member("y", 2, 2, "deploy", 1), task("z"))
		pollNone(t, q, 2, only("y", 1))
		require.NoError(t, q.Done(t.Context(), "x", model.StatusSuccess))
		assert.Equal(t, "y", poll(t, q, 2, only("y", 1)).ID)
		assert.Equal(t, "z", poll(t, q, 3, anyTask).ID)
	})
}

func TestBlackboxConcurrencyLimits(t *testing.T) {
	t.Parallel()

	t.Run("limit one serializes a group", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, member("x", 1, 1, "deploy", 1), member("y", 2, 2, "deploy", 1))
		assert.Equal(t, "x", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		require.NoError(t, q.Done(t.Context(), "x", model.StatusSuccess))
		assert.Equal(t, "y", poll(t, q, 1, anyTask).ID)
	})

	t.Run("limit two runs two at once", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, member("x", 1, 1, "build", 2), member("y", 2, 2, "build", 2), member("z", 3, 3, "build", 2))
		assert.Equal(t, "x", poll(t, q, 1, anyTask).ID)
		assert.Equal(t, "y", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		require.NoError(t, q.Done(t.Context(), "y", model.StatusSuccess))
		assert.Equal(t, "z", poll(t, q, 1, anyTask).ID)
	})

	t.Run("groups do not limit each other", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q, member("x", 1, 1, "build", 1), member("y", 2, 2, "deploy", 1), task("z"))
		assert.Equal(t, "x", poll(t, q, 1, anyTask).ID)
		assert.Equal(t, "y", poll(t, q, 1, anyTask).ID)
		assert.Equal(t, "z", poll(t, q, 1, anyTask).ID)
	})

	t.Run("earlier pipeline waiting on dependencies keeps its slot", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		push(t, q,
			member("build-1", 1, 1, "", 0),
			member("deploy-1", 1, 1, "deploy", 1, "build-1"),
			member("deploy-2", 2, 2, "deploy", 1),
		)
		assert.Equal(t, "build-1", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)

		require.NoError(t, q.Done(t.Context(), "build-1", model.StatusSuccess))
		assert.Equal(t, "deploy-1", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		require.NoError(t, q.Done(t.Context(), "deploy-1", model.StatusSuccess))
		assert.Equal(t, "deploy-2", poll(t, q, 1, anyTask).ID)
	})

	t.Run("same pipeline does not deadlock on its own order", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		// "deploy" is ordered before "test" by name, but depends on it.
		push(t, q, member("test", 1, 1, "ci", 1), member("deploy", 1, 1, "ci", 1, "test"))
		assert.Equal(t, "test", poll(t, q, 1, anyTask).ID)
		require.NoError(t, q.Done(t.Context(), "test", model.StatusSuccess))
		assert.Equal(t, "deploy", poll(t, q, 1, anyTask).ID)
	})

	t.Run("equal order across pipelines does not deadlock", func(t *testing.T) {
		t.Parallel()
		q := newBlackboxQueue(t)
		a := member("a", 1, 5, "deploy", 1)
		b := member("b", 2, 5, "deploy", 1)
		a.Name, b.Name = "deploy", "deploy"
		push(t, q, a, b)
		assert.Equal(t, "a", poll(t, q, 1, anyTask).ID)
		pollNone(t, q, 1, anyTask)
		require.NoError(t, q.Done(t.Context(), "a", model.StatusSuccess))
		assert.Equal(t, "b", poll(t, q, 1, anyTask).ID)
	})
}

// TestBlackboxSimulation pushes random pipelines with dependencies and
// concurrency groups, serves them with a pool of agents and checks the
// scheduling guarantees after every round, until every task finished.
func TestBlackboxSimulation(t *testing.T) {
	t.Parallel()
	for seed := range uint64(6) {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			t.Parallel()
			runSimulation(t, seed)
		})
	}
}

type simTask struct {
	*model.Task
	order int // position in the dispatch log, -1 while not dispatched
}

func runSimulation(t *testing.T, seed uint64) {
	rng := rand.New(rand.NewPCG(seed, 7303))
	limits := map[string]int{"build": 1 + rng.IntN(2), "deploy": 1}

	// generate pipelines: same-second creation and shared workflow names make
	// ordering ties likely, dependencies only point to earlier workflows of
	// the same pipeline, as they do in real pipelines.
	var batches [][]*model.Task
	tasks := make(map[string]*simTask)
	for pipeline := range int64(2 + rng.IntN(4)) {
		var batch []*model.Task
		for workflow := range 1 + rng.IntN(5) {
			task := &model.Task{
				ID:         fmt.Sprintf("p%d-w%d", pipeline, workflow),
				Name:       fmt.Sprintf("w%d", workflow),
				PipelineID: pipeline,
				Created:    pipeline / 2,
				DepStatus:  make(map[string]model.StatusValue),
			}
			for _, dep := range batch {
				if rng.IntN(3) == 0 {
					task.Dependencies = append(task.Dependencies, dep.ID)
				}
			}
			if group := []string{"", "build", "deploy"}[rng.IntN(3)]; group != "" {
				task.ConcurrencyGroup = group
				task.ConcurrencyLimit = limits[group]
			}
			batch = append(batch, task)
			tasks[task.ID] = &simTask{Task: task, order: -1}
		}
		batches = append(batches, batch)
	}
	rng.Shuffle(len(batches), func(i, j int) { batches[i], batches[j] = batches[j], batches[i] })

	q := newBlackboxQueue(t)
	for _, batch := range batches {
		push(t, q, batch...)
	}

	// agents report each task they get and finish it when told to.
	type assignment struct {
		task   *model.Task
		finish chan model.StatusValue
	}
	const agents = 3
	dispatched := make(chan assignment, agents)
	for agent := range int64(agents) {
		go func() {
			for {
				got, err := q.Poll(t.Context(), agent+1, anyTask)
				if err != nil {
					return
				}
				finish := make(chan model.StatusValue)
				dispatched <- assignment{got, finish}
				status := <-finish
				if status == model.StatusSuccess {
					_ = q.Done(t.Context(), got.ID, status)
				} else {
					_ = q.Error(t.Context(), got.ID, errors.New("exit code 1"))
				}
			}
		}()
	}

	finished := make(map[string]model.StatusValue)
	running := make(map[string]chan model.StatusValue)
	var log []string
	for len(finished) < len(tasks) {
		// collect everything the queue hands out until it is quiet again.
		var round []string
		for quiet := time.After(settle); quiet != nil; {
			select {
			case a := <-dispatched:
				round = append(round, a.task.ID)
				running[a.task.ID] = a.finish
				task := tasks[a.task.ID]
				require.Equal(t, -1, task.order, "%s dispatched twice", a.task.ID)
				task.order = len(log)
				log = append(log, a.task.ID)
			case <-quiet:
				quiet = nil
			}
		}
		t.Logf("round %v running %d finished %d", round, len(running), len(finished))
		require.NotEmpty(t, running, "deadlock: nothing runs, %d of %d tasks finished, log %v", len(finished), len(tasks), log)

		for _, id := range round {
			task := tasks[id]
			for _, dep := range task.Dependencies {
				status, ok := finished[dep]
				require.True(t, ok, "%s started before its dependency %s finished", id, dep)
				assert.Equal(t, status, task.DepStatus[dep], "%s sees wrong status of %s", id, dep)
			}
			if task.ConcurrencyLimit == 1 {
				// slots are granted in creation order across pipelines.
				for _, other := range tasks {
					if other.ConcurrencyGroup == task.ConcurrencyGroup && other.PipelineID != task.PipelineID && orderLess(other.Task, task.Task) {
						require.True(t, other.order >= 0 && other.order < task.order, "%s overtook earlier %s", id, other.ID)
					}
				}
			}
		}
		runningPerGroup := make(map[string]int)
		for id := range running {
			if group := tasks[id].ConcurrencyGroup; group != "" {
				runningPerGroup[group]++
				require.LessOrEqual(t, runningPerGroup[group], limits[group], "group %s over its limit", group)
			}
		}

		// finish a random, non-empty part of the running tasks.
		ids := make([]string, 0, len(running))
		for id := range running {
			ids = append(ids, id)
		}
		slices.Sort(ids)
		rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
		for _, id := range ids[:1+rng.IntN(len(ids))] {
			status := model.StatusSuccess
			if rng.IntN(4) == 0 {
				status = model.StatusFailure
			}
			finished[id] = status
			running[id] <- status
			delete(running, id)
		}
	}
	t.Logf("dispatch log %v", log)
}

// orderLess mirrors the documented queue order: pipeline creation time first,
// then the workflow name.
func orderLess(a, b *model.Task) bool {
	if a.Created != b.Created {
		return a.Created < b.Created
	}
	return a.Name < b.Name
}
