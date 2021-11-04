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

package model

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/queue"
)

// Task defines scheduled pipeline Task.
type Task struct {
	ID           string            `xorm:"PK AUTOINCR 'task_id'"`
	Data         []byte            `xorm:"'task_data'"` //MEDIUMBLOB
	Labels       map[string]string `xorm:"json 'task_labels'"`
	Dependencies []string          `xorm:"json 'task_dependencies'"`
	RunOn        []string          `xorm:"json 'task_run_on'"`
}

// TableName return database table name for xorm
func (Task) TableName() string {
	return "tasks"
}

// TaskStore defines storage for scheduled Tasks.
type TaskStore interface {
	TaskList() ([]*Task, error)
	TaskInsert(*Task) error
	TaskDelete(string) error
}

// WithTaskStore returns a queue that is backed by the TaskStore. This
// ensures the task Queue can be restored when the system starts.
func WithTaskStore(q queue.Queue, s TaskStore) queue.Queue {
	tasks, _ := s.TaskList()
	var toEnqueue []*queue.Task
	for _, task := range tasks {
		toEnqueue = append(toEnqueue, &queue.Task{
			ID:           task.ID,
			Data:         task.Data,
			Labels:       task.Labels,
			Dependencies: task.Dependencies,
			RunOn:        task.RunOn,
			DepStatus:    make(map[string]string),
		})
	}
	q.PushAtOnce(context.Background(), toEnqueue)
	return &persistentQueue{q, s}
}

type persistentQueue struct {
	queue.Queue
	store TaskStore
}

// Push pushes a task to the tail of this queue.
func (q *persistentQueue) Push(c context.Context, task *queue.Task) error {
	q.store.TaskInsert(&Task{
		ID:           task.ID,
		Data:         task.Data,
		Labels:       task.Labels,
		Dependencies: task.Dependencies,
		RunOn:        task.RunOn,
	})
	err := q.Queue.Push(c, task)
	if err != nil {
		q.store.TaskDelete(task.ID)
	}
	return err
}

// Push pushes multiple tasks to the tail of this queue.
func (q *persistentQueue) PushAtOnce(c context.Context, tasks []*queue.Task) error {
	for _, task := range tasks {
		q.store.TaskInsert(&Task{
			ID:           task.ID,
			Data:         task.Data,
			Labels:       task.Labels,
			Dependencies: task.Dependencies,
			RunOn:        task.RunOn,
		})
	}
	err := q.Queue.PushAtOnce(c, tasks)
	if err != nil {
		for _, task := range tasks {
			q.store.TaskDelete(task.ID)
		}
	}
	return err
}

// Poll retrieves and removes a task head of this queue.
func (q *persistentQueue) Poll(c context.Context, f queue.Filter) (*queue.Task, error) {
	task, err := q.Queue.Poll(c, f)
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
		q.store.TaskDelete(id)
	}
	return err
}

// Evict removes a pending task from the queue.
func (q *persistentQueue) EvictAtOnce(c context.Context, ids []string) error {
	err := q.Queue.EvictAtOnce(c, ids)
	if err == nil {
		for _, id := range ids {
			q.store.TaskDelete(id)
		}
	}
	return err
}
