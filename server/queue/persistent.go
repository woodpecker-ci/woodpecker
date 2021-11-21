package queue

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// WithTaskStore returns a queue that is backed by the TaskStore. This
// ensures the task Queue can be restored when the system starts.
func WithTaskStore(q Queue, s model.TaskStore) Queue {
	tasks, _ := s.TaskList()
	var toEnqueue []*Task
	for _, task := range tasks {
		toEnqueue = append(toEnqueue, &Task{
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
	Queue
	store model.TaskStore
}

// Push pushes a task to the tail of this queue.
func (q *persistentQueue) Push(c context.Context, task *Task) error {
	q.store.TaskInsert(&model.Task{
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

// PushAtOnce pushes multiple tasks to the tail of this queue.
func (q *persistentQueue) PushAtOnce(c context.Context, tasks []*Task) error {
	for _, task := range tasks {
		q.store.TaskInsert(&model.Task{
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
func (q *persistentQueue) Poll(c context.Context, f Filter) (*Task, error) {
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

// EvictAtOnce removes a pending task from the queue.
func (q *persistentQueue) EvictAtOnce(c context.Context, ids []string) error {
	err := q.Queue.EvictAtOnce(c, ids)
	if err == nil {
		for _, id := range ids {
			q.store.TaskDelete(id)
		}
	}
	return err
}
