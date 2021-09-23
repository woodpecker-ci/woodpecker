package queue

import (
	"container/list"
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	StatusSkipped = "skipped"
	StatusSuccess = "success"
	StatusFailure = "failure"
)

type entry struct {
	item     *Task
	done     chan bool
	retry    int
	error    error
	deadline time.Time
}

type worker struct {
	filter  Filter
	channel chan *Task
}

type fifo struct {
	sync.Mutex

	workers       map[*worker]struct{}
	running       map[string]*entry
	pending       *list.List
	waitingOnDeps *list.List
	extension     time.Duration
	paused        bool
}

// New returns a new fifo queue.
func New() Queue {
	return &fifo{
		workers:       map[*worker]struct{}{},
		running:       map[string]*entry{},
		pending:       list.New(),
		waitingOnDeps: list.New(),
		extension:     time.Minute * 10,
		paused:        false,
	}
}

// Push pushes an item to the tail of this queue.
func (q *fifo) Push(c context.Context, task *Task) error {
	q.Lock()
	q.pending.PushBack(task)
	q.Unlock()
	go q.process()
	return nil
}

// Push pushes an item to the tail of this queue.
func (q *fifo) PushAtOnce(c context.Context, tasks []*Task) error {
	q.Lock()
	for _, task := range tasks {
		q.pending.PushBack(task)
	}
	q.Unlock()
	go q.process()
	return nil
}

// Poll retrieves and removes the head of this queue.
func (q *fifo) Poll(c context.Context, f Filter) (*Task, error) {
	q.Lock()
	w := &worker{
		channel: make(chan *Task, 1),
		filter:  f,
	}
	q.workers[w] = struct{}{}
	q.Unlock()
	go q.process()

	for {
		select {
		case <-c.Done():
			q.Lock()
			delete(q.workers, w)
			q.Unlock()
			return nil, nil
		case t := <-w.channel:
			return t, nil
		}
	}
}

// Done signals that the item is done executing.
func (q *fifo) Done(c context.Context, id string, exitStatus string) error {
	return q.finished([]string{id}, exitStatus, nil)
}

// Error signals that the item is done executing with error.
func (q *fifo) Error(c context.Context, id string, err error) error {
	return q.finished([]string{id}, StatusFailure, err)
}

// Error signals that the item is done executing with error.
func (q *fifo) ErrorAtOnce(c context.Context, id []string, err error) error {
	return q.finished(id, StatusFailure, err)
}

func (q *fifo) finished(ids []string, exitStatus string, err error) error {
	q.Lock()

	for _, id := range ids {
		taskEntry, ok := q.running[id]
		if ok {
			taskEntry.error = err
			close(taskEntry.done)
			delete(q.running, id)
		} else {
			q.removeFromPending(id)
		}
		q.updateDepStatusInQueue(id, exitStatus)
	}

	q.Unlock()
	return nil
}

// Evict removes a pending task from the queue.
func (q *fifo) Evict(c context.Context, id string) error {
	return q.EvictAtOnce(c, []string{id})
}

// Evict removes a pending task from the queue.
func (q *fifo) EvictAtOnce(c context.Context, ids []string) error {
	q.Lock()
	defer q.Unlock()

	for _, id := range ids {
		var next *list.Element
		for e := q.pending.Front(); e != nil; e = next {
			next = e.Next()
			task, ok := e.Value.(*Task)
			if ok && task.ID == id {
				q.pending.Remove(e)
				return nil
			}
		}
	}
	return ErrNotFound
}

// Wait waits until the item is done executing.
func (q *fifo) Wait(c context.Context, id string) error {
	q.Lock()
	state := q.running[id]
	q.Unlock()
	if state != nil {
		select {
		case <-c.Done():
		case <-state.done:
			return state.error
		}
	}
	return nil
}

// Extend extends the task execution deadline.
func (q *fifo) Extend(c context.Context, id string) error {
	q.Lock()
	defer q.Unlock()

	state, ok := q.running[id]
	if ok {
		state.deadline = time.Now().Add(q.extension)
		return nil
	}
	return ErrNotFound
}

// Info returns internal queue information.
func (q *fifo) Info(c context.Context) InfoT {
	q.Lock()
	stats := InfoT{}
	stats.Stats.Workers = len(q.workers)
	stats.Stats.Pending = q.pending.Len()
	stats.Stats.WaitingOnDeps = q.waitingOnDeps.Len()
	stats.Stats.Running = len(q.running)

	for e := q.pending.Front(); e != nil; e = e.Next() {
		stats.Pending = append(stats.Pending, e.Value.(*Task))
	}
	for e := q.waitingOnDeps.Front(); e != nil; e = e.Next() {
		stats.WaitingOnDeps = append(stats.WaitingOnDeps, e.Value.(*Task))
	}
	for _, entry := range q.running {
		stats.Running = append(stats.Running, entry.item)
	}
	stats.Paused = q.paused

	q.Unlock()
	return stats
}

func (q *fifo) Pause() {
	q.Lock()
	q.paused = true
	q.Unlock()
}

func (q *fifo) Resume() {
	q.Lock()
	q.paused = false
	q.Unlock()
	go q.process()
}

// helper function that loops through the queue and attempts to
// match the item to a single subscriber.
func (q *fifo) process() {
	q.Lock()
	defer q.Unlock()

	if q.paused {
		return
	}

	defer func() {
		// the risk of panic is low. This code can probably be removed
		// once the code has been used in real world installs without issue.
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("queue: unexpected panic: %v\n%s", err, buf)
		}
	}()

	q.resubmitExpiredBuilds()
	q.filterWaiting()
	for pending, worker := q.assignToWorker(); pending != nil && worker != nil; pending, worker = q.assignToWorker() {
		task := pending.Value.(*Task)
		delete(q.workers, worker)
		q.pending.Remove(pending)
		q.running[task.ID] = &entry{
			item:     task,
			done:     make(chan bool),
			deadline: time.Now().Add(q.extension),
		}
		worker.channel <- task
	}
}

func (q *fifo) filterWaiting() {
	// resubmits all waiting tasks to pending, deps may have cleared
	var nextWaiting *list.Element
	for e := q.waitingOnDeps.Front(); e != nil; e = nextWaiting {
		nextWaiting = e.Next()
		task := e.Value.(*Task)
		q.pending.PushBack(task)
	}

	// rebuild waitingDeps
	q.waitingOnDeps = list.New()
	filtered := []*list.Element{}
	var nextPending *list.Element
	for e := q.pending.Front(); e != nil; e = nextPending {
		nextPending = e.Next()
		task := e.Value.(*Task)
		if q.depsInQueue(task) {
			logrus.Debugf("queue: waiting due to unmet dependencies %v", task.ID)
			q.waitingOnDeps.PushBack(task)
			filtered = append(filtered, e)
		}
	}

	// filter waiting tasks
	for _, f := range filtered {
		q.pending.Remove(f)
	}
}

func (q *fifo) assignToWorker() (*list.Element, *worker) {
	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*Task)
		logrus.Debugf("queue: trying to assign task: %v with deps %v", task.ID, task.Dependencies)

		for w := range q.workers {
			if w.filter(task) {
				logrus.Debugf("queue: assigned task: %v with deps %v", task.ID, task.Dependencies)
				return e, w
			}
		}
	}

	return nil, nil
}

func (q *fifo) resubmitExpiredBuilds() {
	for id, state := range q.running {
		if time.Now().After(state.deadline) {
			q.pending.PushFront(state.item)
			delete(q.running, id)
			close(state.done)
		}
	}
}

func (q *fifo) depsInQueue(task *Task) bool {
	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		possibleDep, ok := e.Value.(*Task)
		logrus.Debugf("queue: pending right now: %v", possibleDep.ID)
		for _, dep := range task.Dependencies {
			if ok && possibleDep.ID == dep {
				return true
			}
		}
	}
	for possibleDepID := range q.running {
		logrus.Debugf("queue: running right now: %v", possibleDepID)
		for _, dep := range task.Dependencies {
			if possibleDepID == dep {
				return true
			}
		}
	}
	return false
}

func (q *fifo) updateDepStatusInQueue(taskID string, status string) {
	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		pending, ok := e.Value.(*Task)
		for _, dep := range pending.Dependencies {
			if ok && taskID == dep {
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

	next = nil
	for e := q.waitingOnDeps.Front(); e != nil; e = next {
		next = e.Next()
		waiting, ok := e.Value.(*Task)
		for _, dep := range waiting.Dependencies {
			if ok && taskID == dep {
				waiting.DepStatus[dep] = status
			}
		}
	}
}

func (q *fifo) removeFromPending(taskID string) {
	logrus.Debugf("queue: trying to remove %s", taskID)
	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*Task)
		if task.ID == taskID {
			logrus.Debugf("queue: %s is removed from pending", taskID)
			q.pending.Remove(e)
			return
		}
	}
}
