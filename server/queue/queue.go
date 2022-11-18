package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

var (
	// ErrCancel indicates the task was canceled.
	ErrCancel = errors.New("queue: task canceled")

	// ErrNotFound indicates the task was not found in the queue.
	ErrNotFound = errors.New("queue: task not found")
)

// Task defines a unit of work in the queue.
type Task struct {
	// ID identifies this task.
	ID string `json:"id,omitempty"`

	// Data is the actual data in the entry.
	Data []byte `json:"data"`

	// Labels represents the key-value pairs the entry is labeled with.
	Labels map[string]string `json:"labels,omitempty"`

	// Task IDs this task depend
	Dependencies []string

	// Dependency's exit status
	DepStatus map[string]string

	// RunOn failure or success
	RunOn []string
}

// ShouldRun tells if a task should be run or skipped, based on dependencies
func (t *Task) ShouldRun() bool {
	if runsOnFailure(t.RunOn) && runsOnSuccess(t.RunOn) {
		return true
	}

	if !runsOnFailure(t.RunOn) && runsOnSuccess(t.RunOn) {
		for _, status := range t.DepStatus {
			if StatusSuccess != status {
				return false
			}
		}
		return true
	}

	if runsOnFailure(t.RunOn) && !runsOnSuccess(t.RunOn) {
		for _, status := range t.DepStatus {
			if StatusSuccess == status {
				return false
			}
		}
		return true
	}

	return false
}

func (t *Task) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%s) - %s", t.ID, t.Dependencies, t.DepStatus))
	return sb.String()
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

func (t *InfoT) String() string {
	var sb strings.Builder

	for _, task := range t.Pending {
		sb.WriteString("\t" + task.String())
	}

	for _, task := range t.Running {
		sb.WriteString("\t" + task.String())
	}

	for _, task := range t.WaitingOnDeps {
		sb.WriteString("\t" + task.String())
	}

	return sb.String()
}

// Filter filters tasks in the queue. If the Filter returns false,
// the Task is skipped and not returned to the subscriber.
type FilterFn func(*Task) bool

// Queue defines a task queue for scheduling tasks among
// a pool of workers.
type Queue interface {
	// Push pushes a task to the tail of this queue.
	Push(c context.Context, task *Task) error

	// PushAtOnce pushes a task to the tail of this queue.
	PushAtOnce(c context.Context, tasks []*Task) error

	// Poll retrieves and removes a task head of this queue.
	Poll(c context.Context, f FilterFn) (*Task, error)

	// Extend extends the deadline for a task.
	Extend(c context.Context, id string) error

	// Done signals the task is complete.
	Done(c context.Context, id string, exitStatus model.StatusValue) error

	// Error signals the task is complete with errors.
	Error(c context.Context, id string, err error) error

	// ErrorAtOnce signals the task is complete with errors.
	ErrorAtOnce(c context.Context, id []string, err error) error

	EvictCurrent(c context.Context, id string, exitStatus model.StatusValue) error

	// Evict removes a pending task from the queue.
	Evict(c context.Context, id string) error

	// EvictAtOnce removes a pending task from the queue.
	EvictAtOnce(c context.Context, id []string) error

	// Wait waits until the task is complete.
	Wait(c context.Context, id string) error

	// Info returns internal queue information.
	Info(c context.Context) InfoT

	// Pause stops the queue from handing out new work items in Poll
	Pause()

	// Resume starts the queue again, Poll returns new items
	Resume()
}
