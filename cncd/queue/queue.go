package queue

import (
	"context"
	"errors"
)

var (
	// ErrCancel indicates the task was cancelled.
	ErrCancel = errors.New("queue: task cancelled")

	// ErrNotFound indicates the task was not found in the queue.
	ErrNotFound = errors.New("queue: task not found")
)

// Task defines a unit of work in the queue.
type Task struct {
	// ID identifies this task.
	ID string `json:"id,omitempty"`

	// Data is the actual data in the entry.
	Data []byte `json:"data"`

	// Labels represents the key-value pairs the entry is lebeled with.
	Labels map[string]string `json:"labels,omitempty"`

	// Task IDs this task depend
	Dependencies []string

	// If dep finished sucessfully
	DepStatus map[string]bool

	// RunOn failure or success
	RunOn []string
}

// ShouldRun tells if a task should be run or skipped, based on dependencies
func (t *Task) ShouldRun() bool {
	if runsOnFailure(t.RunOn) && runsOnSuccess(t.RunOn) {
		return true
	}

	if !runsOnFailure(t.RunOn) && runsOnSuccess(t.RunOn) {
		for _, success := range t.DepStatus {
			if !success {
				return false
			}
		}
		return true
	}

	if runsOnFailure(t.RunOn) && !runsOnSuccess(t.RunOn) {
		for _, success := range t.DepStatus {
			if success {
				return false
			}
		}
		return true
	}

	return false
}

func runsOnFailure(runsOn []string) bool {
	for _, status := range runsOn {
		if status == "failure" {
			return true
		}
	}
	return false
}

func runsOnSuccess(runsOn []string) bool {
	if len(runsOn) == 0 {
		return true
	}

	for _, status := range runsOn {
		if status == "success" {
			return true
		}
	}
	return false
}

// InfoT provides runtime information.
type InfoT struct {
	Pending       []*Task `json:"pending"`
	WaitingOnDeps []*Task `json:"waiting_on_deps"`
	Running       []*Task `json:"running"`
	Stats         struct {
		Workers       int `json:"worker_count"`
		Pending       int `json:"pending_count"`
		WaitingOnDeps int `json:"waiting_on_deps_count"`
		Running       int `json:"running_count"`
		Complete      int `json:"completed_count"`
	} `json:"stats"`
	Paused bool
}

// Filter filters tasks in the queue. If the Filter returns false,
// the Task is skipped and not returned to the subscriber.
type Filter func(*Task) bool

// Queue defines a task queue for scheduling tasks among
// a pool of workers.
type Queue interface {
	// Push pushes a task to the tail of this queue.
	Push(c context.Context, task *Task) error

	// Push pushes a task to the tail of this queue.
	PushAtOnce(c context.Context, tasks []*Task) error

	// Poll retrieves and removes a task head of this queue.
	Poll(c context.Context, f Filter) (*Task, error)

	// Extend extends the deadline for a task.
	Extend(c context.Context, id string) error

	// Done signals the task is complete.
	Done(c context.Context, id string) error

	// Error signals the task is complete with errors.
	Error(c context.Context, id string, err error) error

	// Evict removes a pending task from the queue.
	Evict(c context.Context, id string) error

	// Wait waits until the task is complete.
	Wait(c context.Context, id string) error

	// Info returns internal queue information.
	Info(c context.Context) InfoT

	// Stops the queue from handing out new work items in Poll
	Pause()

	// Starts the queue again, Poll returns new items
	Resume()
}
