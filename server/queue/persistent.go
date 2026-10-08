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
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// WithTaskStore returns a queue that is backed by the TaskStore. This
// ensures the task Queue can be restored when the system starts.
func WithTaskStore(ctx context.Context, q Queue, s store.Store) Queue {
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

func isTerminalWorkflowState(state model.StatusValue) bool {
	switch state {
	case model.StatusSuccess,
		model.StatusFailure,
		model.StatusKilled,
		model.StatusCanceled,
		model.StatusSkipped,
		model.StatusError,
		model.StatusDeclined:
		return true
	default:
		return false
	}
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
func (q *persistentQueue) Poll(c context.Context, agentID int64, f func(*model.Task) (bool, int)) (*model.Task, error) {
	task, err := q.Queue.Poll(c, agentID, f)
	if task != nil {
		log.Debug().Msgf("pull queue item: %s: remove from backup", task.ID)
		if deleteErr := q.store.TaskDelete(task.ID); deleteErr != nil {
			if errors.Is(deleteErr, types.ErrRecordNotExist) {
				// The task is no longer in the backup store, which means it was
				// already finished/removed (e.g. its workflow was canceled) and
				// only lingered in the in-memory queue. Handing it to the agent
				// would loop forever, so drop it from the queue instead.
				log.Error().Err(deleteErr).Msgf("pull queue item: %s: not found in backup, dropping stale task", task.ID)
				if dropErr := q.Queue.Error(c, task.ID, deleteErr); dropErr != nil {
					log.Error().Err(dropErr).Msgf("pull queue item: %s: failed to drop stale task", task.ID)
				}
				return nil, nil
			}
			log.Error().Err(deleteErr).Msgf("pull queue item: %s: failed to remove from backup", task.ID)
		} else {
			log.Debug().Msgf("pull queue item: %s: successfully removed from backup", task.ID)

			workflowID, parseErr := strconv.ParseInt(task.ID, 10, 64)
			if parseErr != nil {
				log.Error().Err(parseErr).Msgf("pull queue item: %s: invalid workflow id", task.ID)
				return task, err
			}

			workflow, loadErr := q.store.WorkflowLoad(workflowID)
			if loadErr != nil {
				if errors.Is(loadErr, types.ErrRecordNotExist) {
					// A queued task without its workflow can never be initialized.
					// Drop it now instead of letting the agent reject and later
					// re-poll the same task after its lease expires.
					log.Error().Err(loadErr).Msgf("pull queue item: %s: workflow missing, dropping stale task", task.ID)
					if dropErr := q.Queue.Error(c, task.ID, loadErr); dropErr != nil {
						log.Error().Err(dropErr).Msgf("pull queue item: %s: failed to drop stale task", task.ID)
					}
					return nil, nil
				}
				log.Error().Err(loadErr).Msgf("pull queue item: %s: failed to load workflow", task.ID)
				return task, err
			}

			if isTerminalWorkflowState(workflow.State) {
				// The task can still exist in the persistent queue while its
				// workflow was completed through another path. Returning it to an
				// agent makes RPC.Init reject it as already finished, while the
				// in-memory task remains running and is resubmitted after timeout.
				log.Warn().Str("state", string(workflow.State)).Msgf("pull queue item: %s: workflow already terminal, dropping stale task", task.ID)
				if dropErr := q.Queue.Done(c, task.ID, workflow.State); dropErr != nil {
					log.Error().Err(dropErr).Msgf("pull queue item: %s: failed to drop terminal task", task.ID)
				}
				return nil, nil
			}
		}
	}
	return task, err
}

// Done signals the task is complete.
func (q *persistentQueue) Done(c context.Context, id string, exitStatus model.StatusValue) error {
	if err := q.Queue.Done(c, id, exitStatus); err != nil {
		return err
	}

	if deleteErr := q.store.TaskDelete(id); deleteErr != nil {
		if !errors.Is(deleteErr, types.ErrRecordNotExist) {
			return deleteErr
		}
		log.Debug().Msgf("task %s already removed from store", id)
	}
	return nil
}

// Error signals the task is done with an error.
func (q *persistentQueue) Error(c context.Context, id string, err error) error {
	if err := q.Queue.Error(c, id, err); err != nil {
		return err
	}

	if deleteErr := q.store.TaskDelete(id); deleteErr != nil {
		if !errors.Is(deleteErr, types.ErrRecordNotExist) {
			return deleteErr
		}
		log.Debug().Msgf("task %s already removed from store", id)
	}
	return nil
}

// ErrorAtOnce signals multiple tasks are done and complete with an error.
// If still pending they will just get removed from the queue.
func (q *persistentQueue) ErrorAtOnce(c context.Context, ids []string, err error) error {
	if err := q.Queue.ErrorAtOnce(c, ids, err); err != nil {
		return err
	}

	var errs []error
	for _, id := range ids {
		if deleteErr := q.store.TaskDelete(id); deleteErr != nil && !errors.Is(deleteErr, types.ErrRecordNotExist) {
			errs = append(errs, fmt.Errorf("task id [%s]: %w", id, deleteErr))
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("failed to delete tasks from persistent store: %w", errors.Join(errs...))
	}
	return nil
}
