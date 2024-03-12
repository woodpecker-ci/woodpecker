// Copyright 2022 Woodpecker Authors
// Copyright 2019 mhmxs.
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

package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func mockStorePipeline(t *testing.T) store.Store {
	s := mocks.NewStore(t)
	s.On("UpdatePipeline", mock.Anything).Return(nil)
	return s
}

func TestUpdateToStatusRunning(t *testing.T) {
	t.Parallel()

	pipeline, _ := UpdateToStatusRunning(mockStorePipeline(t), model.Pipeline{}, int64(1))
	assert.Equal(t, model.StatusRunning, pipeline.Status)
	assert.EqualValues(t, 1, pipeline.Started)
}

func TestUpdateToStatusPending(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusPending(mockStorePipeline(t), model.Pipeline{}, "Reviewer")

	assert.Equal(t, model.StatusPending, pipeline.Status)
	assert.Equal(t, "Reviewer", pipeline.Reviewer)
	assert.LessOrEqual(t, now, pipeline.Reviewed)
}

func TestUpdateToStatusDeclined(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusDeclined(mockStorePipeline(t), model.Pipeline{}, "Reviewer")

	assert.Equal(t, model.StatusDeclined, pipeline.Status)
	assert.Equal(t, "Reviewer", pipeline.Reviewer)
	assert.LessOrEqual(t, now, pipeline.Reviewed)
}

func TestUpdateToStatusToDone(t *testing.T) {
	t.Parallel()

	pipeline, _ := UpdateStatusToDone(mockStorePipeline(t), model.Pipeline{}, "status", int64(1))

	assert.Equal(t, model.StatusValue("status"), pipeline.Status)
	assert.EqualValues(t, 1, pipeline.Finished)
}

func TestUpdateToStatusError(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusError(mockStorePipeline(t), model.Pipeline{}, errors.New("this is an error"))

	assert.Len(t, pipeline.Errors, 1)
	assert.Equal(t, "[generic] this is an error", pipeline.Errors[0].Error())
	assert.Equal(t, model.StatusError, pipeline.Status)
	assert.Equal(t, pipeline.Started, pipeline.Finished)
	assert.LessOrEqual(t, now, pipeline.Started)
	assert.Equal(t, pipeline.Started, pipeline.Finished)
}

func TestUpdateToStatusKilled(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusKilled(mockStorePipeline(t), model.Pipeline{})

	assert.Equal(t, model.StatusKilled, pipeline.Status)
	assert.LessOrEqual(t, now, pipeline.Finished)
}
