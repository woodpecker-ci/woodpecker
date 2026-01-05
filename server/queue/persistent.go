// Copyright 2021 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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
	"context"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// WithTaskStore returns a queue that is backed by the TaskStore. This
// ensures the task Queue can be restored when the system starts.
func WithTaskStore(ctx context.Context, q Queue, s store.Store) Queue {
	q.SetStore(s)
	tasks, _ := s.TaskList()
	if err := q.PushAtOnce(ctx, tasks); err != nil {
		log.Error().Err(err).Msg("PushAtOnce failed")
	}
	return &persistentQueue{q, s}
}

type persistentQueue struct {
	Queue
	store store.Store
}

// PushAtOnce pushes multiple tasks to the tail of this queue.
func (q *persistentQueue) PushAtOnce(c context.Context, tasks []*model.Task) error {
	// TODO: invent store.NewSession who return context including a session and make TaskInsert & TaskDelete use it
	for _, task := range tasks {
		if err := q.store.TaskInsert(task); err != nil {
			return err
		}
	}
	err := q.Queue.PushAtOnce(c, tasks)
	if err != nil {
		for _, task := range tasks {
			if err := q.store.TaskDelete(task.ID); err != nil {
				return err
			}
		}
	}
	return err
}

// Poll retrieves and removes a task head of this queue.
func (q *persistentQueue) Poll(c context.Context, agentID int64, f FilterFn) (*model.Task, error) {
	task, err := q.Queue.Poll(c, agentID, f)
	// NOTE: We intentionally do NOT delete from TaskStore here.
	// Tasks are kept in the store until Done/Error to support workflow recovery.
	// If an agent crashes after polling, the task will be reloaded from the store on server restart.
	return task, err
}

// EvictAtOnce removes multiple pending tasks from the queue.
func (q *persistentQueue) EvictAtOnce(c context.Context, ids []string) error {
	if err := q.Queue.EvictAtOnce(c, ids); err != nil {
		return err
	}
	for _, id := range ids {
		if err := q.store.TaskDelete(id); err != nil {
			return err
		}
	}
	return nil
}

// Done signals the task is complete.
func (q *persistentQueue) Done(c context.Context, id string, exitStatus model.StatusValue) error {
	if err := q.Queue.Done(c, id, exitStatus); err != nil {
		return err
	}
	return q.store.TaskDelete(id)
}

// Error signals the task is done with an error.
func (q *persistentQueue) Error(c context.Context, id string, err error) error {
	if err := q.Queue.Error(c, id, err); err != nil {
		return err
	}
	return q.store.TaskDelete(id)
}

// ErrorAtOnce signals multiple tasks are done with an error.
func (q *persistentQueue) ErrorAtOnce(c context.Context, ids []string, err error) error {
	if err := q.Queue.ErrorAtOnce(c, ids, err); err != nil {
		return err
	}
	for _, id := range ids {
		if err := q.store.TaskDelete(id); err != nil {
			return err
		}
	}
	return nil
}

func (q *persistentQueue) SetStore(s store.Store) {
	q.Queue.SetStore(s)
	q.store = s
}
