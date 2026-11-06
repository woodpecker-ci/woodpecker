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

package queue

import (
	"container/list"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// scanCanRunConcurrent is the original scan-based implementation, retained as
// an independent oracle for ordering, mixed limits, and within-tick updates.
func scanCanRunConcurrent(q *fifo, task *model.Task) bool {
	if task.ConcurrencyLimit <= 0 || task.ConcurrencyGroup == "" {
		return true
	}
	occupied := 0
	for _, running := range q.running {
		if running.item.ConcurrencyGroup == task.ConcurrencyGroup {
			occupied++
		}
	}
	if occupied >= task.ConcurrencyLimit {
		return false
	}
	for _, queued := range []*list.List{q.pending, q.waitingOnDeps} {
		for element := queued.Front(); element != nil; element = element.Next() {
			other, _ := element.Value.(*model.Task)
			if other.ConcurrencyGroup == task.ConcurrencyGroup && other.ID != task.ID &&
				other.PipelineID != task.PipelineID && taskOrderLess(other, task) {
				occupied++
			}
		}
	}
	return occupied < task.ConcurrencyLimit
}

func scanAssignToWorker(q *fifo) (*list.Element, *worker) {
	for element := q.pending.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		if !scanCanRunConcurrent(q, task) {
			continue
		}
		var best *worker
		bestScore := 0
		for worker := range q.workers {
			if matched, score := worker.filter(task); matched && score > bestScore {
				best, bestScore = worker, score
			}
		}
		if best != nil {
			return element, best
		}
	}
	return nil, nil
}

func TestConcurrencyIndexMatchesScan(t *testing.T) {
	random := rand.New(rand.NewPCG(42, 17))
	for fixture := range 500 {
		q := &fifo{pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry)}
		for i := range 20 + random.IntN(60) {
			task := &model.Task{
				ID: fmt.Sprintf("task-%d", i), Name: fmt.Sprintf("workflow-%d", random.IntN(4)),
				Created: int64(random.IntN(4)), PipelineID: int64(random.IntN(8)),
				ConcurrencyGroup: []string{"", "deploy", "build", "test"}[random.IntN(4)],
				ConcurrencyLimit: random.IntN(8) - 1,
			}
			switch random.IntN(5) {
			case 0:
				q.waitingOnDeps.PushBack(task)
			case 1:
				q.running[task.ID] = &entry{item: task}
			default:
				q.pending.PushBack(task)
			}
		}

		index := q.newConcurrencyIndex()
		for {
			var eligible []*list.Element
			for element := q.pending.Front(); element != nil; element = element.Next() {
				task, _ := element.Value.(*model.Task)
				want := scanCanRunConcurrent(q, task)
				require.Equal(t, want, index.canRun(task), "fixture %d, task %+v", fixture, task)
				if want {
					eligible = append(eligible, element)
				}
			}
			if len(eligible) == 0 {
				break
			}
			// Vary which eligible task finds a matching worker, including equal
			// ordering keys and tasks behind unmatched earlier candidates.
			element := eligible[random.IntN(len(eligible))]
			task, _ := element.Value.(*model.Task)
			index.start(task)
			q.pending.Remove(element)
			q.running[task.ID] = &entry{item: task}
		}
	}
}

func TestConcurrencyIndexStarts(t *testing.T) {
	tests := []struct {
		name              string
		first, second     *model.Task
		allowedAfterStart bool
	}{
		{
			name:   "same pipeline running task consumes a slot",
			first:  &model.Task{ID: "a", Name: "a", PipelineID: 1, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
			second: &model.Task{ID: "b", Name: "b", PipelineID: 1, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
		},
		{
			name:   "equal order running task consumes a slot",
			first:  &model.Task{ID: "a", Name: "deploy", PipelineID: 1, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
			second: &model.Task{ID: "b", Name: "deploy", PipelineID: 2, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
		},
		{
			name:              "reservation becomes running without double counting",
			first:             &model.Task{ID: "a", Created: 1, PipelineID: 1, ConcurrencyGroup: "deploy", ConcurrencyLimit: 2},
			second:            &model.Task{ID: "b", Created: 2, PipelineID: 2, ConcurrencyGroup: "deploy", ConcurrencyLimit: 2},
			allowedAfterStart: true,
		},
		{
			name:   "unlimited start before lazy index construction",
			first:  &model.Task{ID: "a", PipelineID: 1, ConcurrencyGroup: "deploy"},
			second: &model.Task{ID: "b", PipelineID: 2, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fifo{
				pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry),
				workers: map[*worker]struct{}{{filter: filterFnTrue}: {}, {filter: filterFnTrue}: {}},
			}
			q.pending.PushBack(tt.first)
			q.pending.PushBack(tt.second)
			index := q.newConcurrencyIndex()
			element, worker := q.assignToWorker(index)
			require.NotNil(t, element)
			require.Same(t, tt.first, element.Value)
			index.start(tt.first)
			q.pending.Remove(element)
			delete(q.workers, worker)
			q.running[tt.first.ID] = &entry{item: tt.first}
			element, _ = q.assignToWorker(index)
			require.Equal(t, tt.allowedAfterStart, element != nil)
		})
	}
}

func TestConcurrencyIndexUnlimitedMembers(t *testing.T) {
	q := &fifo{pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry)}
	task := &model.Task{ID: "limited", PipelineID: 2, Created: 2, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1}
	q.pending.PushBack(task)
	q.waitingOnDeps.PushBack(&model.Task{ID: "unlimited", PipelineID: 1, Created: 1, ConcurrencyGroup: "deploy"})
	require.False(t, q.newConcurrencyIndex().canRun(task))
	q.waitingOnDeps.Init()
	q.running["unlimited"] = &entry{item: &model.Task{ID: "unlimited", ConcurrencyGroup: "deploy"}}
	require.False(t, q.newConcurrencyIndex().canRun(task))
}

func TestConcurrencyIndexLazy(t *testing.T) {
	q := &fifo{pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry)}
	task := &model.Task{ID: "unlimited"}
	q.pending.PushBack(task)
	index := q.newConcurrencyIndex()
	require.True(t, index.canRun(task))
	index.start(task)
	require.Nil(t, index.slots, "unlimited scheduling must not build an index")
}

func TestConcurrencyIndexFullGroup(t *testing.T) {
	q := &fifo{pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry)}
	low := &model.Task{ID: "low", PipelineID: 1, ConcurrencyGroup: "deploy", ConcurrencyLimit: 1}
	high := &model.Task{ID: "high", PipelineID: 2, ConcurrencyGroup: "deploy", ConcurrencyLimit: 2}
	q.pending.PushBack(low)
	q.pending.PushBack(high)
	q.running["active"] = &entry{item: &model.Task{ID: "active", ConcurrencyGroup: "deploy"}}
	index := q.newConcurrencyIndex()
	require.False(t, index.canRun(low))
	require.True(t, index.canRun(high), "use the group's maximum candidate limit, not the first limit")
	q.running["other"] = &entry{item: &model.Task{ID: "other", ConcurrencyGroup: "deploy"}}
	index = q.newConcurrencyIndex()
	require.False(t, index.canRun(low))
	require.False(t, index.canRun(high))
	require.Empty(t, index.slots, "fully occupied groups need no ordered index")
	delete(q.running, "other")
	require.True(t, q.newConcurrencyIndex().canRun(high), "the next tick rebuilds occupancy")
}

func TestConcurrencyCounts(t *testing.T) {
	counts := make(concurrencyCounts, 8)
	counts.init()
	for end := range 8 {
		require.Equal(t, end, counts.prefix(end))
	}
	counts.remove(0)
	counts.remove(3)
	counts.remove(6)
	for end, want := range []int{0, 0, 1, 2, 2, 3, 4, 4} {
		require.Equal(t, want, counts.prefix(end))
	}
}
