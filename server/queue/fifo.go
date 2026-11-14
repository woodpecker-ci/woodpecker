// Copyright 2022 Woodpecker Authors
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
	"cmp"
	"container/list"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

type entry struct {
	item     *model.Task
	done     chan bool
	error    error
	deadline time.Time
}

type worker struct {
	agentID int64
	filter  func(*model.Task) (bool, int)
	channel chan *entry
	stop    context.CancelCauseFunc
}

type fifo struct {
	sync.Mutex

	ctx           context.Context
	workers       map[*worker]struct{}
	running       map[string]*entry
	pending       *list.List
	waitingOnDeps *list.List
	queuedIDs     map[string]struct{} // reused by filterWaiting as scratchpad to have less gc
	extension     time.Duration
	paused        bool
}

// processTimeInterval is the time till the queue rearranges things,
// as the agent pull in 10 milliseconds we should also give them work asap.
const processTimeInterval = 100 * time.Millisecond

// NewMemoryQueue returns a new fifo queue.
func NewMemoryQueue(ctx context.Context) Queue {
	q := &fifo{
		ctx:           ctx,
		workers:       make(map[*worker]struct{}),
		running:       make(map[string]*entry),
		pending:       list.New(),
		waitingOnDeps: list.New(),
		extension:     constant.TaskTimeout,
		paused:        false,
	}
	go q.process()
	return q
}

// PushAtOnce pushes multiple tasks to the tail of this queue.
func (q *fifo) PushAtOnce(_ context.Context, tasks []*model.Task) error {
	q.Lock()
	for _, task := range tasks {
		q.pending.PushBack(task)
	}
	q.Unlock()
	return nil
}

// Poll retrieves and removes a task head of this queue.
func (q *fifo) Poll(c context.Context, agentID int64, filter func(*model.Task) (bool, int)) (*model.Task, error) {
	return q.pollWithValidation(c, agentID, filter, nil)
}

// pollWithValidation validates an assignment outside the queue lock, then checks
// that cancellation or expiry did not remove or replace that exact lease while
// validation was in progress.
func (q *fifo) pollWithValidation(c context.Context, agentID int64, filter func(*model.Task) (bool, int), validate func(*model.Task) *model.Task) (*model.Task, error) {
	q.Lock()
	ctx, stop := context.WithCancelCause(c)

	w := &worker{
		agentID: agentID,
		channel: make(chan *entry, 1),
		filter:  filter,
		stop:    stop,
	}
	q.workers[w] = struct{}{}
	q.Unlock()

	for {
		select {
		case <-ctx.Done():
			q.Lock()
			delete(q.workers, w)
			q.Unlock()
			return nil, ctx.Err()
		case assigned := <-w.channel:
			task := assigned.item
			if validate != nil {
				task = validate(task)
			}
			q.Lock()
			current := q.running[assigned.item.ID] == assigned
			q.Unlock()
			if !current {
				return nil, nil
			}
			return task, nil
		}
	}
}

// Done signals the task is complete.
func (q *fifo) Done(_ context.Context, id string, exitStatus model.StatusValue) error {
	return q.finished([]string{id}, exitStatus, nil)
}

// Error signals the task is done with an error.
func (q *fifo) Error(_ context.Context, id string, err error) error {
	return q.finished([]string{id}, model.StatusFailure, err)
}

// ErrorAtOnce signals multiple tasks are done and complete with an error.
// If still pending they will just get removed from the queue.
func (q *fifo) ErrorAtOnce(_ context.Context, ids []string, err error) error {
	if errors.Is(err, ErrCancel) {
		return q.finished(ids, model.StatusKilled, err)
	}
	return q.finished(ids, model.StatusFailure, err)
}

// locks the queue itself!
func (q *fifo) finished(ids []string, exitStatus model.StatusValue, err error) error {
	q.Lock()
	defer q.Unlock()

	// it's an external error so we wrap it
	err = NewErrExternal(err)

	var errs []error
	// we first process the tasks itself
	for _, id := range ids {
		if taskEntry, ok := q.running[id]; ok {
			taskEntry.error = err
			close(taskEntry.done)
			delete(q.running, id)
		} else {
			errs = append(errs, q.removeFromPendingAndWaiting(id))
		}
	}

	// next we aim for there dependencies
	// we do this because in our ids list there could be tasks and its dependencies
	// so not to mess things up
	for _, id := range ids {
		q.updateDepStatusInQueue(id, exitStatus)
	}

	return errors.Join(errs...)
}

// Wait waits until the item leased to the agent is done executing.
// Also signals via error ErrCancel if workflow got canceled, ErrTaskExpired
// if the task is queued again and ErrAgentMissMatch if another agent holds
// the lease, so a slow agent stops instead of running next to the new one.
func (q *fifo) Wait(ctx context.Context, agentID int64, taskID string) error {
	q.Lock()
	state, err := q.leaseOf(agentID, taskID)
	q.Unlock()
	if err != nil || state == nil {
		return err
	}
	select {
	case <-ctx.Done():
	case <-state.done:
		// check if we have a wrapped cancel error and unwrap it
		if errors.Is(state.error, ErrCancel) {
			return ErrCancel
		}
		// or return queue errors and no workflow errors
		if !errors.Is(state.error, new(ErrExternal)) {
			return state.error
		}
	}
	return nil
}

// Leased reports whether the agent holds the lease of the task.
func (q *fifo) Leased(_ context.Context, agentID int64, taskID string) error {
	q.Lock()
	defer q.Unlock()

	state, err := q.leaseOf(agentID, taskID)
	if err != nil {
		return err
	}
	if state == nil {
		return ErrNotFound
	}
	return nil
}

// leaseOf returns the running entry of the task if the agent holds its lease.
// A task that is pending or waiting on dependencies again lost its lease
// (ErrTaskExpired), one running under another agent was handed out again
// (ErrAgentMissMatch), and one the queue does not know gives nil, nil.
// Expects the queue to be locked by the caller.
func (q *fifo) leaseOf(agentID int64, taskID string) (*entry, error) {
	if state, ok := q.running[taskID]; ok {
		if state.item.AgentID != agentID {
			return nil, ErrAgentMissMatch
		}
		return state, nil
	}
	if q.queued(taskID) {
		return nil, ErrTaskExpired
	}
	return nil, nil
}

// queued reports whether the task is pending or waiting on dependencies.
// Expects the queue to be locked by the caller.
func (q *fifo) queued(taskID string) bool {
	for element := q.pending.Front(); element != nil; element = element.Next() {
		if task, _ := element.Value.(*model.Task); task.ID == taskID {
			return true
		}
	}
	for element := q.waitingOnDeps.Front(); element != nil; element = element.Next() {
		if task, _ := element.Value.(*model.Task); task.ID == taskID {
			return true
		}
	}
	return false
}

// Extend extends the task execution deadline.
func (q *fifo) Extend(_ context.Context, agentID int64, taskID string) error {
	q.Lock()
	defer q.Unlock()

	state, ok := q.running[taskID]
	if ok {
		if state.item.AgentID != agentID {
			return ErrAgentMissMatch
		}

		state.deadline = time.Now().Add(q.extension)
		return nil
	}
	return ErrNotFound
}

// Info returns internal queue information.
func (q *fifo) Info(_ context.Context) InfoT {
	q.Lock()
	stats := InfoT{}
	stats.Stats.Workers = len(q.workers)
	stats.Stats.Pending = q.pending.Len()
	stats.Stats.WaitingOnDeps = q.waitingOnDeps.Len()
	stats.Stats.Running = len(q.running)

	for element := q.pending.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		stats.Pending = append(stats.Pending, task)
	}
	for element := q.waitingOnDeps.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		stats.WaitingOnDeps = append(stats.WaitingOnDeps, task)
	}
	for _, entry := range q.running {
		stats.Running = append(stats.Running, entry.item)
	}
	stats.Paused = q.paused

	q.Unlock()
	return stats
}

// Pause stops the queue from handing out new work items in Poll.
func (q *fifo) Pause() {
	q.Lock()
	q.paused = true
	q.Unlock()
}

// Resume starts the queue again.
func (q *fifo) Resume() {
	q.Lock()
	q.paused = false
	q.Unlock()
}

// KickAgentWorkers kicks all workers for a given agent.
func (q *fifo) KickAgentWorkers(agentID int64) {
	q.Lock()
	defer q.Unlock()

	for worker := range q.workers {
		if worker.agentID == agentID {
			worker.stop(ErrWorkerKicked)
			delete(q.workers, worker)
		}
	}
}

// helper function that loops through the queue and attempts to
// match the item to a single subscriber until context got cancel.
func (q *fifo) process() {
	for {
		select {
		case <-time.After(processTimeInterval):
		case <-q.ctx.Done():
			return
		}

		q.Lock()
		if q.paused {
			q.Unlock()
			continue
		}

		q.resubmitExpiredPipelines()
		q.filterWaiting()
		concurrency := &concurrencyIndex{queue: q}
		for pending, worker := q.assignToWorker(concurrency); pending != nil && worker != nil; pending, worker = q.assignToWorker(concurrency) {
			task, _ := pending.Value.(*model.Task)
			concurrency.start(task)
			task.AgentID = worker.agentID
			delete(q.workers, worker)
			q.pending.Remove(pending)
			q.running[task.ID] = &entry{
				item:     task,
				done:     make(chan bool),
				deadline: time.Now().Add(q.extension),
			}
			worker.channel <- q.running[task.ID]
		}
		q.Unlock()
	}
}

// filterWaiting moves pending tasks whose dependencies are still queued or
// running to waitingOnDeps, and waiting tasks whose dependencies have cleared
// back to the tail of pending. Tasks that keep their state stay where they are,
// so a steady queue is checked without moving anything.
//
// The result is ordered as if all waiting tasks were appended to pending and
// the combined list was filtered again: newly blocked tasks go in front of the
// tasks that are still waiting, and released tasks keep their relative order.
func (q *fifo) filterWaiting() {
	// The index of queued IDs is only built when the first task with
	// dependencies shows up, which is before any task moves. It also holds the
	// waiting tasks, as they can themselves be dependencies.
	indexed := false
	blocked := func(task *model.Task) bool {
		if len(task.Dependencies) == 0 {
			return false
		}
		if !indexed {
			indexed = true
			if q.queuedIDs == nil {
				q.queuedIDs = make(map[string]struct{}, q.pending.Len()+q.waitingOnDeps.Len())
			}
			clear(q.queuedIDs)
			for _, tasks := range []*list.List{q.pending, q.waitingOnDeps} {
				for element := tasks.Front(); element != nil; element = element.Next() {
					queued, _ := element.Value.(*model.Task)
					q.queuedIDs[queued.ID] = struct{}{}
				}
			}
		}
		return q.depsInQueue(task, q.queuedIDs)
	}

	// Insert newly blocked tasks before the first task that was already
	// waiting. Inserted tasks are never revisited by the waiting loop below,
	// as it starts at this element.
	firstWaiting := q.waitingOnDeps.Front()
	for element := q.pending.Front(); element != nil; {
		next := element.Next()
		if task, _ := element.Value.(*model.Task); blocked(task) {
			q.pending.Remove(element)
			if firstWaiting != nil {
				q.waitingOnDeps.InsertBefore(task, firstWaiting)
			} else {
				q.waitingOnDeps.PushBack(task)
			}
		}
		element = next
	}

	for element := firstWaiting; element != nil; {
		next := element.Next()
		if task, _ := element.Value.(*model.Task); !blocked(task) {
			q.waitingOnDeps.Remove(element)
			q.pending.PushBack(task)
		}
		element = next
	}
}

func (q *fifo) assignToWorker(concurrency *concurrencyIndex) (*list.Element, *worker) {
	if len(q.workers) == 0 {
		return nil, nil
	}
	var bestWorker *worker
	var bestScore int

	for element := q.pending.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)

		// skip tasks that would exceed their workflow concurrency limit, they
		// stay pending and are retried on the next process tick.
		if !concurrency.canRun(task) {
			continue
		}

		for worker := range q.workers {
			matched, score := worker.filter(task)
			if matched && score > bestScore {
				bestWorker = worker
				bestScore = score
			}
		}
		if bestWorker != nil {
			log.Debug().
				Str("task_id", task.ID).
				Strs("dependencies", task.Dependencies).
				Int("score", bestScore).
				Msg("queue: assigned task to worker")
			return element, bestWorker
		}
	}

	return nil, nil
}

// taskOrderLess reports whether task a was instantiated before task b. Ordering
// is by the Created timestamp (the pipeline creation time), with the workflow
// name as a deterministic tiebreaker for tasks created within the same second.
// The task ID is intentionally not used for ordering.
func taskOrderLess(a, b *model.Task) bool {
	return taskOrderCompare(a, b) < 0
}

func taskOrderCompare(a, b *model.Task) int {
	return cmp.Or(cmp.Compare(a.Created, b.Created), cmp.Compare(a.Name, b.Name))
}

func (q *fifo) resubmitExpiredPipelines() {
	for taskID, taskState := range q.running {
		if time.Now().After(taskState.deadline) {
			log.Info().Str("task_id", taskID).Msg("queue: resubmitting expired task")
			taskState.error = ErrTaskExpired
			q.pending.PushFront(taskState.item)
			delete(q.running, taskID)
			close(taskState.done)
		}
	}
}

func (q *fifo) depsInQueue(task *model.Task, queuedIDs map[string]struct{}) bool {
	for _, dep := range task.Dependencies {
		if _, ok := queuedIDs[dep]; ok {
			return true
		}
		if _, ok := q.running[dep]; ok {
			return true
		}
	}
	return false
}

// expects the q to be currently owned e.g. locked by caller!
func (q *fifo) updateDepStatusInQueue(taskID string, status model.StatusValue) {
	for element := q.pending.Front(); element != nil; element = element.Next() {
		pending, _ := element.Value.(*model.Task)
		for _, dep := range pending.Dependencies {
			if taskID == dep {
				pending.DepStatus[dep] = status
			}
		}
	}

	for _, running := range q.running {
		for _, dep := range running.item.Dependencies {
			if taskID == dep {
				running.item.DepStatus[dep] = status
			}
		}
	}

	for element := q.waitingOnDeps.Front(); element != nil; element = element.Next() {
		waiting, _ := element.Value.(*model.Task)
		for _, dep := range waiting.Dependencies {
			if taskID == dep {
				waiting.DepStatus[dep] = status
			}
		}
	}
}

// expects the q to be currently owned e.g. locked by caller!
func (q *fifo) removeFromPendingAndWaiting(taskID string) error {
	// we assume pending first
	for element := q.pending.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		if task.ID == taskID {
			log.Debug().Str("task_id", taskID).Msg("queue: task removed from pending")
			_ = q.pending.Remove(element)
			return nil
		}
	}

	// well looks like it's waiting
	for element := q.waitingOnDeps.Front(); element != nil; element = element.Next() {
		task, _ := element.Value.(*model.Task)
		if task.ID == taskID {
			log.Debug().Str("task_id", taskID).Msg("queue: task removed from waitingOnDeps")
			_ = q.waitingOnDeps.Remove(element)
			return nil
		}
	}

	// well it could not be found
	return ErrNotFound
}
