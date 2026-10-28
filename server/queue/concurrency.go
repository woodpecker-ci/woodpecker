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
	"slices"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// cspell:words Fenwick

// concurrencyIndex lives for one scheduling tick, under the queue lock. It counts
// running slots and earlier queued tasks without scanning the queue again for each
// candidate. Only starts mutate the index; other queue changes happen next tick.
type concurrencyIndex struct {
	queue *fifo
	slots map[*model.Task]concurrencySlot
}

type concurrencySlot struct {
	group                      *concurrencyGroup
	pipeline                   *concurrencyPipeline
	position, pipelinePosition int
	before, pipelineBefore     int
}

type concurrencyGroup struct {
	tasks   []*model.Task
	running int
	limit   int
	queued  concurrencyCounts
}

type concurrencyPipeline struct {
	count  int
	before int
	last   *model.Task
	queued concurrencyCounts
}

// newConcurrencyIndex defers construction until a limited candidate is checked,
// so unlimited queues keep their allocation-free, constant-time fast path.
func (q *fifo) newConcurrencyIndex() *concurrencyIndex {
	return &concurrencyIndex{queue: q}
}

// build takes O(N log N) time and O(N) space. The caller must hold the queue
// lock; the snapshot includes any unlimited tasks already started this tick.
func (index *concurrencyIndex) build() {
	q := index.queue
	var groups map[string]*concurrencyGroup
	for element := q.pending.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		if task.ConcurrencyLimit > 0 && task.ConcurrencyGroup != "" {
			if groups == nil {
				groups = make(map[string]*concurrencyGroup)
			}
			if groups[task.ConcurrencyGroup] == nil {
				groups[task.ConcurrencyGroup] = new(concurrencyGroup)
			}
			group := groups[task.ConcurrencyGroup]
			group.limit = max(group.limit, task.ConcurrencyLimit)
		}
	}
	if len(groups) == 0 {
		return
	}

	for _, running := range q.running {
		if group := groups[running.item.ConcurrencyGroup]; group != nil {
			group.running++
		}
	}

	// Include unlimited members too: they reserve order and occupy running slots
	// for limited tasks sharing their group. Fully occupied groups need no ordered
	// index: starts can only increase occupancy during this locked tick.
	for _, queued := range []*list.List{q.pending, q.waitingOnDeps} {
		for element := queued.Front(); element != nil; element = element.Next() {
			task, _ := element.Value.(*model.Task)
			if group := groups[task.ConcurrencyGroup]; group != nil && group.running < group.limit {
				group.tasks = append(group.tasks, task)
			}
		}
	}
	size := 0
	for _, group := range groups {
		size += len(group.tasks)
	}
	index.slots = make(map[*model.Task]concurrencySlot, size)
	for _, group := range groups {
		if len(group.tasks) == 0 {
			continue
		}
		slices.SortFunc(group.tasks, taskOrderCompare)
		// Allocate pipeline state together rather than once per pipeline. Large
		// backlogs commonly have one queued task per pipeline.
		pipelines := make([]concurrencyPipeline, len(group.tasks))
		pipelineByID := make(map[int64]*concurrencyPipeline, len(group.tasks))
		before := 0
		for position, task := range group.tasks {
			if position > 0 && taskOrderLess(group.tasks[position-1], task) {
				before = position
			}
			pipeline := pipelineByID[task.PipelineID]
			if pipeline == nil {
				pipeline = &pipelines[len(pipelineByID)]
				pipelineByID[task.PipelineID] = pipeline
			}
			if pipeline.last != nil && taskOrderLess(pipeline.last, task) {
				pipeline.before = pipeline.count
			}
			index.slots[task] = concurrencySlot{
				group: group, pipeline: pipeline,
				position: position, pipelinePosition: pipeline.count,
				before: before, pipelineBefore: pipeline.before,
			}
			pipeline.count++
			pipeline.last = task
		}
		// All trees share one backing allocation. Their total size is one
		// entry per task in each index, plus a sentinel per tree.
		counts := make(concurrencyCounts, 2*len(group.tasks)+len(pipelineByID)+1)
		group.queued, counts = counts[:len(group.tasks)+1], counts[len(group.tasks)+1:]
		group.queued.init()
		for _, pipeline := range pipelineByID {
			pipeline.queued, counts = counts[:pipeline.count+1], counts[pipeline.count+1:]
			pipeline.queued.init()
		}
		group.tasks = nil
	}
}

// canRun reports whether the given task may currently start without violating
// its workflow concurrency limit. Tasks without a limit always pass, keeping the
// default scheduling behavior unchanged.
//
// Slots within a concurrency group are granted in creation order (see
// taskOrderLess) rather than in the order tasks become ready. This guarantees
// that a later pipeline whose dependencies happen to finish faster cannot
// overtake an earlier one that is still waiting. Equal keys do not reserve
// order against each other.
//
// The ordering reservation only applies across pipelines. Within a single
// pipeline, execution order is already defined by depends_on, and reserving a
// slot for an earlier workflow could deadlock when it depends on a later
// workflow that shares the same group (e.g. deploy.yaml depending on
// test.yaml). Because dependencies never cross pipelines, restricting the
// reservation to other pipelines keeps the ordering guarantee while making
// such deadlocks impossible. Running tasks occupy a slot regardless of pipeline.
func (index *concurrencyIndex) canRun(task *model.Task) bool {
	if task.ConcurrencyLimit <= 0 || task.ConcurrencyGroup == "" {
		return true
	}
	if index.slots == nil {
		index.build()
	}
	slot, ok := index.slots[task]
	if !ok {
		// The group was already full at construction and stays full this tick.
		return false
	}
	ahead := slot.group.queued.prefix(slot.before) - slot.pipeline.queued.prefix(slot.pipelineBefore)
	return slot.group.running+ahead < task.ConcurrencyLimit
}

// start updates both indexes before the next assignment in the same tick.
func (index *concurrencyIndex) start(task *model.Task) {
	if slot, ok := index.slots[task]; ok {
		slot.group.queued.remove(slot.position)
		slot.pipeline.queued.remove(slot.pipelinePosition)
		slot.group.running++
	}
}

// concurrencyCounts is a Fenwick tree of queued-task counts. Initially every
// position contains one task. Prefix queries and removals take O(log N) time.
type concurrencyCounts []int

func (counts concurrencyCounts) init() {
	for i := 1; i < len(counts); i++ {
		counts[i] = i & -i
	}
}

func (counts concurrencyCounts) prefix(end int) int {
	total := 0
	for i := end; i > 0; i -= i & -i {
		total += counts[i]
	}
	return total
}

func (counts concurrencyCounts) remove(position int) {
	for i := position + 1; i < len(counts); i += i & -i {
		counts[i]--
	}
}
