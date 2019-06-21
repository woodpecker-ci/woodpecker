package queue

import (
	"container/list"
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
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

	workers   map[*worker]struct{}
	running   map[string]*entry
	pending   *list.List
	extension time.Duration
}

// New returns a new fifo queue.
func New() Queue {
	return &fifo{
		workers:   map[*worker]struct{}{},
		running:   map[string]*entry{},
		pending:   list.New(),
		extension: time.Minute * 10,
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
	for range tasks {
		go q.process()
	}
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
func (q *fifo) Done(c context.Context, id string) error {
	return q.Error(c, id, nil)
}

// Error signals that the item is done executing with error.
func (q *fifo) Error(c context.Context, id string, err error) error {
	q.Lock()
	taskEntry, ok := q.running[id]
	if ok {
		q.updateDepStatusInQueue(id, err == nil)
		taskEntry.error = err
		close(taskEntry.done)
		delete(q.running, id)
	} else {
		q.removeFromPending(id)
	}
	q.Unlock()
	go q.process()
	return nil
}

// Evict removes a pending task from the queue.
func (q *fifo) Evict(c context.Context, id string) error {
	q.Lock()
	defer q.Unlock()

	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		task, ok := e.Value.(*Task)
		if ok && task.ID == id {
			q.pending.Remove(e)
			return nil
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
	stats.Stats.Running = len(q.running)

	for e := q.pending.Front(); e != nil; e = e.Next() {
		stats.Pending = append(stats.Pending, e.Value.(*Task))
	}
	for _, entry := range q.running {
		stats.Running = append(stats.Running, entry.item)
	}

	q.Unlock()
	return stats
}

// helper function that loops through the queue and attempts to
// match the item to a single subscriber.
func (q *fifo) process() {
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

	q.Lock()
	defer q.Unlock()

	q.resubmitExpiredBuilds()

	var next *list.Element
loop:
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*Task)
		logrus.Debugf("queue: trying to assign task: %v with deps %v", task.ID, task.Dependencies)
		if q.depsInQueue(task) {
			logrus.Debugf("queue: skipping due to unmet dependencies %v", task.ID)
			continue
		}
		for w := range q.workers {
			if w.filter(task) {
				delete(q.workers, w)
				q.pending.Remove(e)

				q.running[task.ID] = &entry{
					item:     task,
					done:     make(chan bool),
					deadline: time.Now().Add(q.extension),
				}

				logrus.Debugf("queue: assigned task: %v with deps %v", task.ID, task.Dependencies)
				w.channel <- task
				break loop
			}
		}
	}
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

func (q *fifo) updateDepStatusInQueue(taskID string, success bool) {
	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		pending, ok := e.Value.(*Task)
		for _, dep := range pending.Dependencies {
			if ok && taskID == dep {
				pending.DepStatus[dep] = success
			}
		}
	}
	for _, running := range q.running {
		for _, dep := range running.item.Dependencies {
			if taskID == dep {
				running.item.DepStatus[dep] = success
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
