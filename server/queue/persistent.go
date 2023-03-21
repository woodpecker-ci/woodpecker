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

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// WithTaskStore returns a queue that is backed by the TaskStore. This
// ensures the task Queue can be restored when the system starts.
func WithTaskStore(q Queue, s model.TaskStore) Queue {
	tasks, _ := s.TaskList()
	if err := q.PushAtOnce(context.Background(), tasks); err != nil {
		log.Error().Err(err).Msg("PushAtOnce failed")
	}
	return &persistentQueue{q, s}
}

type persistentQueue struct {
	Queue
	store model.TaskStore
}

// Push pushes a task to the tail of this queue.
func (q *persistentQueue) Push(c context.Context, task *model.Task) error {
	if err := q.store.TaskInsert(task); err != nil {
		return err
	}
	err := q.Queue.Push(c, task)
	if err != nil {
		if err2 := q.store.TaskDelete(task.ID); err2 != nil {
			err = errors.Wrapf(err, "delete task '%s' failed: %v", task.ID, err2)
		}
	}
	return err
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
	if task != nil {
		log.Debug().Msgf("pull queue item: %s: remove from backup", task.ID)
		if derr := q.store.TaskDelete(task.ID); derr != nil {
			log.Error().Msgf("pull queue item: %s: failed to remove from backup: %s", task.ID, derr)
		} else {
			log.Debug().Msgf("pull queue item: %s: successfully removed from backup", task.ID)
		}
	}
	return task, err
}

// Evict removes a pending task from the queue.
func (q *persistentQueue) Evict(c context.Context, id string) error {
	err := q.Queue.Evict(c, id)
	if err == nil {
		return q.store.TaskDelete(id)
	}
	return err
}

// EvictAtOnce removes a pending task from the queue.
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
