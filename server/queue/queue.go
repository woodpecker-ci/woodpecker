// Copyright 2023 Woodpecker Authors
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

package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

var (
	// ErrCancel indicates the task was canceled.
	ErrCancel = errors.New("queue: task canceled")

	// ErrNotFound indicates the task was not found in the queue.
	ErrNotFound = errors.New("queue: task not found")

	// ErrAgentMissMatch indicates a task is assigned to a different agent.
	ErrAgentMissMatch = errors.New("task assigned to different agent")

	// ErrTaskExpired indicates a running task exceeded its lease/deadline and was resubmitted.
	ErrTaskExpired = errors.New("queue: task expired")

	// ErrWorkerKicked worker of an agent got kicked.
	ErrWorkerKicked = errors.New("worker was kicked")
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
	} `json:"stats"`
	Paused bool `json:"paused"`
} //	@name	InfoT

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

// FilterFn filters tasks in the queue. If the Filter returns false,
// the Task is skipped and not returned to the subscriber.
// The int return value represents the matching score (higher is better).
type FilterFn func(*model.Task) (bool, int)

// Queue defines a task queue for scheduling tasks among
// a pool of workers.
type Queue interface {
	// PushAtOnce pushes multiple tasks to the tail of this queue.
	PushAtOnce(c context.Context, tasks []*model.Task) error

	// Poll retrieves and removes a task head of this queue.
	Poll(c context.Context, agentID int64, f FilterFn) (*model.Task, error)

	// Extend extends the deadline for a task.
	Extend(c context.Context, agentID int64, workflowID string) error

	// Done signals the task is complete.
	Done(c context.Context, id string, exitStatus model.StatusValue) error

	// Error signals the task is done with an error.
	Error(c context.Context, id string, err error) error

	// ErrorAtOnce signals multiple tasks are done and complete with an error.
	// If still pending they will just get removed from the queue.
	ErrorAtOnce(c context.Context, ids []string, err error) error

	// Wait waits until the task is complete.
	// Also signals via error ErrCancel if workflow got canceled.
	Wait(c context.Context, id string) error

	// Info returns internal queue information.
	Info(c context.Context) InfoT

	// Pause stops the queue from handing out new work items in Poll
	Pause()

	// Resume starts the queue again.
	Resume()

	// KickAgentWorkers kicks all workers for a given agent.
	KickAgentWorkers(agentID int64)
}

// Config holds the configuration for the queue.
type Config struct {
	Backend Type
	Store   store.Store
}

// Queue type.
type Type string

const (
	TypeMemory Type = "memory"
)

// New creates a new queue based on the provided configuration.
func New(ctx context.Context, config Config) (Queue, error) {
	var q Queue

	switch config.Backend {
	case TypeMemory:
		q = NewMemoryQueue(ctx)
		if config.Store != nil {
			q = WithTaskStore(ctx, q, config.Store)
		}
	default:
		return nil, fmt.Errorf("unsupported queue backend: %s", config.Backend)
	}

	return q, nil
}
