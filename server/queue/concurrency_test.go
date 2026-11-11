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
	"testing"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

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
			index := &concurrencyIndex{queue: q}
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
	require.False(t, (&concurrencyIndex{queue: q}).canRun(task))
	q.waitingOnDeps.Init()
	q.running["unlimited"] = &entry{item: &model.Task{ID: "unlimited", ConcurrencyGroup: "deploy"}}
	require.False(t, (&concurrencyIndex{queue: q}).canRun(task))
}

func TestConcurrencyIndexLazy(t *testing.T) {
	q := &fifo{pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry)}
	task := &model.Task{ID: "unlimited"}
	q.pending.PushBack(task)
	index := &concurrencyIndex{queue: q}
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
	index := &concurrencyIndex{queue: q}
	require.False(t, index.canRun(low))
	require.True(t, index.canRun(high), "use the group's maximum candidate limit, not the first limit")
	q.running["other"] = &entry{item: &model.Task{ID: "other", ConcurrencyGroup: "deploy"}}
	index = &concurrencyIndex{queue: q}
	require.False(t, index.canRun(low))
	require.False(t, index.canRun(high))
	require.Empty(t, index.slots, "fully occupied groups need no ordered index")
	delete(q.running, "other")
	require.True(t, (&concurrencyIndex{queue: q}).canRun(high), "the next tick rebuilds occupancy")
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
