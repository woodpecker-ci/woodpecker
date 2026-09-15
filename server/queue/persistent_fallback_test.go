// Copyright 2026 Woodpecker Authors
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// WithTaskStore can wrap queues other than FIFO. A partial Poll result must
// still be validated, without losing the wrapped queue's error for live work.
func TestPersistentQueueWrappedPollError(t *testing.T) {
	for _, state := range []model.StatusValue{model.StatusPending, model.StatusCanceled, "no-task"} {
		t.Run(string(state), func(t *testing.T) {
			ctx := t.Context()
			q := queue_mocks.NewMockQueue(t)
			store := store_mocks.NewMockStore(t)
			pollErr := errors.New("wrapped queue poll error")
			var task *model.Task
			if state != "no-task" {
				task = genDummyTask()
				store.EXPECT().TaskDelete(task.ID).Return(nil).Once()
				store.EXPECT().WorkflowLoad(int64(1)).Return(&model.Workflow{ID: 1, State: state}, nil).Once()
			}
			q.EXPECT().Poll(ctx, int64(7), mock.Anything).Return(task, pollErr).Once()
			if state == model.StatusCanceled {
				q.EXPECT().Done(ctx, task.ID, state).Return(nil).Once()
			}
			pq := &persistentQueue{Queue: q, store: store}

			got, err := pq.Poll(ctx, 7, filterFnTrue)
			if state == model.StatusCanceled {
				assert.Nil(t, got, "stale partial results must still be removed")
				assert.NoError(t, err)
			} else {
				assert.Same(t, task, got)
				assert.ErrorIs(t, err, pollErr)
			}
		})
	}
}
