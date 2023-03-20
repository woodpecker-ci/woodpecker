package queue

import (
	"context"
	"errors"
	"strings"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

var (
	// ErrCancel indicates the task was canceled.
	ErrCancel = errors.New("queue: task canceled")

	// ErrNotFound indicates the task was not found in the queue.
	ErrNotFound = errors.New("queue: task not found")
)

// InfoT provides runtime information.
type InfoT struct {
	Pending       []*model.Task `json:"pending"`
	WaitingOnDeps []*model.Task `json:"waiting_on_deps"`
	Running       []*model.Task `json:"running"`
	Stats         struct {
		Workers       int `json:"worker_count"`
		Pending       int `json:"pending_count"`
		WaitingOnDeps int `json:"waiting_on_deps_count"`
		Running       int `json:"running_count"`
		Complete      int `json:"completed_count"`
	} `json:"stats"`
	Paused bool `json:"paused"`
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
type FilterFn func(*model.Task) bool

// Queue defines a task queue for scheduling tasks among
// a pool of workers.
type Queue interface {
	// Push pushes a task to the tail of this queue.
	Push(c context.Context, task *model.Task) error

	// PushAtOnce pushes a task to the tail of this queue.
	PushAtOnce(c context.Context, tasks []*model.Task) error

	// Poll retrieves and removes a task head of this queue.
	Poll(c context.Context, f FilterFn) (*model.Task, error)

	// Extend extends the deadline for a task.
	Extend(c context.Context, id string) error

	// Done signals the task is complete.
	Done(c context.Context, id string, exitStatus model.StatusValue) error

	// Error signals the task is complete with errors.
	Error(c context.Context, id string, err error) error

	// ErrorAtOnce signals the task is complete with errors.
	ErrorAtOnce(c context.Context, id []string, err error) error

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
