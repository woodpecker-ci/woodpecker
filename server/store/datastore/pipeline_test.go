// Copyright 2022 Woodpecker Authors
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

package datastore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestPipelines(t *testing.T) {
	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}

	store, closer := newTestStore(t, new(model.Repo), new(model.Step), new(model.Pipeline))
	defer closer()

	assert.NoError(t, store.CreateRepo(repo))

	// Fail when the repo is not existing
	pipeline := model.Pipeline{
		RepoID: 100,
		Status: model.StatusSuccess,
	}
	err := store.CreatePipeline(&pipeline)
	assert.Error(t, err)

	count, err := store.GetPipelineCount()
	assert.NoError(t, err)
	assert.Zero(t, count)

	// add pipeline
	pipeline = model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusSuccess,
		Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		Branch: "some-branch",
	}
	err = store.CreatePipeline(&pipeline)
	assert.NoError(t, err)
	assert.NotZero(t, pipeline.ID)
	assert.EqualValues(t, 1, pipeline.Number)
	assert.Equal(t, "85f8c029b902ed9400bc600bac301a0aadb144ac", pipeline.Commit)

	count, err = store.GetPipelineCount()
	assert.NoError(t, err)
	assert.NotZero(t, count)

	GetPipeline, err := store.GetPipeline(pipeline.ID)
	assert.NoError(t, err)
	assert.Equal(t, pipeline.ID, GetPipeline.ID)
	assert.Equal(t, pipeline.RepoID, GetPipeline.RepoID)
	assert.Equal(t, pipeline.Status, GetPipeline.Status)

	// update pipeline
	pipeline.Status = model.StatusRunning
	err1 := store.UpdatePipeline(&pipeline)
	GetPipeline, err2 := store.GetPipeline(pipeline.ID)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, pipeline.ID, GetPipeline.ID)
	assert.Equal(t, pipeline.RepoID, GetPipeline.RepoID)
	assert.Equal(t, pipeline.Status, GetPipeline.Status)
	assert.Equal(t, pipeline.Number, GetPipeline.Number)

	pipeline2 := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusPending,
		Event:  model.EventPush,
		Branch: "main",
	}
	err2 = store.CreatePipeline(pipeline2, []*model.Step{}...)
	assert.NoError(t, err2)
	GetPipeline, err3 := store.GetPipelineNumber(&model.Repo{ID: 1}, pipeline2.Number)
	assert.NoError(t, err3)
	assert.Equal(t, pipeline2.ID, GetPipeline.ID)
	assert.Equal(t, pipeline2.RepoID, GetPipeline.RepoID)
	assert.Equal(t, pipeline2.Number, GetPipeline.Number)

	GetPipeline, err3 = store.GetPipelineLast(&model.Repo{ID: repo.ID}, pipeline2.Branch)
	assert.NoError(t, err3)
	assert.Equal(t, pipeline2.ID, GetPipeline.ID)
	assert.Equal(t, pipeline2.RepoID, GetPipeline.RepoID)
	assert.Equal(t, pipeline2.Number, GetPipeline.Number)
	assert.Equal(t, pipeline2.Status, GetPipeline.Status)

	pipeline3 := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusRunning,
		Branch: "main",
		Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
	}
	err1 = store.CreatePipeline(pipeline3, []*model.Step{}...)
	assert.NoError(t, err1)

	GetPipeline, err4 := store.GetPipelineLastBefore(&model.Repo{ID: 1}, pipeline3.Branch, pipeline3.ID)
	assert.NoError(t, err4)
	assert.Equal(t, pipeline2.ID, GetPipeline.ID)
	assert.Equal(t, pipeline2.RepoID, GetPipeline.RepoID)
	assert.Equal(t, pipeline2.Number, GetPipeline.Number)
	assert.Equal(t, pipeline2.Status, GetPipeline.Status)
	assert.Equal(t, pipeline2.Branch, GetPipeline.Branch)
	assert.Equal(t, pipeline2.Commit, GetPipeline.Commit)
}

func TestPipelineListFilter(t *testing.T) {
	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}

	store, closer := newTestStore(t, new(model.Repo), new(model.Step), new(model.Pipeline))
	defer closer()

	assert.NoError(t, store.CreateRepo(repo))

	pipeline1 := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusFailure,
		Event:  model.EventCron,
		Ref:    "refs/heads/some-branch",
		Branch: "some-branch",
	}
	pipeline2 := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusSuccess,
		Event:  model.EventPull,
		Ref:    "refs/pull/32",
		Branch: "main",
	}
	err := store.CreatePipeline(pipeline1, []*model.Step{}...)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	before := time.Now().Unix()
	err = store.CreatePipeline(pipeline2, []*model.Step{}...)
	assert.NoError(t, err)

	pipelines, err := store.GetPipelineList(&model.Repo{ID: 1}, &model.ListOptions{Page: 1, PerPage: 50}, nil)
	assert.NoError(t, err)
	assert.Len(t, (pipelines), 2)
	assert.Equal(t, pipeline2.ID, pipelines[0].ID)
	assert.Equal(t, pipeline2.RepoID, pipelines[0].RepoID)
	assert.Equal(t, pipeline2.Status, pipelines[0].Status)

	pipelines, err = store.GetPipelineList(&model.Repo{ID: 1}, nil, &model.PipelineFilter{
		Branch: "main",
	})
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
	assert.Equal(t, pipeline2.ID, pipelines[0].ID)

	pipelines, err = store.GetPipelineList(&model.Repo{ID: 1}, nil, &model.PipelineFilter{
		Events: []model.WebhookEvent{model.EventCron},
	})
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
	assert.Equal(t, pipeline1.ID, pipelines[0].ID)

	pipelines, err = store.GetPipelineList(&model.Repo{ID: 1}, nil, &model.PipelineFilter{
		Events:      []model.WebhookEvent{model.EventCron, model.EventPull},
		RefContains: "32",
	})
	assert.NoError(t, err)
	assert.Len(t, (pipelines), 1)
	assert.Equal(t, pipeline2.ID, pipelines[0].ID)

	pipelines, err3 := store.GetPipelineList(&model.Repo{ID: 1}, &model.ListOptions{Page: 1, PerPage: 50}, &model.PipelineFilter{Before: before})
	assert.NoError(t, err3)
	assert.Len(t, pipelines, 1)
	assert.Equal(t, pipeline1.ID, pipelines[0].ID)
	assert.Equal(t, pipeline1.RepoID, pipelines[0].RepoID)

	pipelines, err = store.GetPipelineList(&model.Repo{ID: 1}, nil, &model.PipelineFilter{
		Status: model.StatusSuccess,
	})
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
	assert.Equal(t, pipeline2.ID, pipelines[0].ID)
	assert.Equal(t, model.StatusSuccess, pipelines[0].Status)
}

func TestPipelineIncrement(t *testing.T) {
	store, closer := newTestStore(t, new(model.Pipeline), new(model.Repo))
	defer closer()

	assert.NoError(t, store.CreateRepo(&model.Repo{ID: 1, Owner: "1", Name: "1", FullName: "1/1"}))
	assert.NoError(t, store.CreateRepo(&model.Repo{ID: 2, Owner: "2", Name: "2", FullName: "2/2"}))

	pipelineA := &model.Pipeline{RepoID: 1}
	if !assert.NoError(t, store.CreatePipeline(pipelineA)) {
		return
	}
	assert.EqualValues(t, 1, pipelineA.Number)

	pipelineB := &model.Pipeline{RepoID: 1}
	assert.NoError(t, store.CreatePipeline(pipelineB))
	assert.EqualValues(t, 2, pipelineB.Number)

	pipelineC := &model.Pipeline{RepoID: 2}
	assert.NoError(t, store.CreatePipeline(pipelineC))
	assert.EqualValues(t, 1, pipelineC.Number)
}

func TestDeletePipeline(t *testing.T) {
	store, closer := newTestStore(t, new(model.Pipeline), new(model.Repo), new(model.Workflow),
		new(model.Step), new(model.LogEntry), new(model.PipelineConfig), new(model.Config))
	defer closer()

	_, err := store.engine.Insert(
		&model.Pipeline{
			ID:     2,
			Number: 2,
			RepoID: 7,
		},
		&model.Pipeline{
			ID:     5,
			Number: 3,
			RepoID: 7,
		},
		&model.Pipeline{
			ID:     8,
			Number: 4,
			RepoID: 7,
		},
		&model.Config{
			ID:     23,
			Hash:   "1234",
			Name:   "test",
			RepoID: 7,
		},
		&model.Config{
			ID:     25,
			Hash:   "6789",
			Name:   "test",
			RepoID: 7,
		},
		&model.PipelineConfig{
			PipelineID: 2,
			ConfigID:   23,
		},
		&model.PipelineConfig{
			PipelineID: 5,
			ConfigID:   23,
		},
		&model.PipelineConfig{
			PipelineID: 8,
			ConfigID:   25,
		},
	)
	assert.NoError(t, err)

	// delete non existing pipeline
	assert.ErrorIs(t, types.RecordNotExist, store.DeletePipeline(&model.Pipeline{ID: 1}))

	// delete pipeline with shares config
	assert.NoError(t, store.DeletePipeline(&model.Pipeline{ID: 2}))
	count, err := store.engine.Count(new(model.Config))
	assert.NoError(t, err)
	assert.EqualValues(t, 2, count)

	// delete pipeline with unique config
	assert.NoError(t, store.DeletePipeline(&model.Pipeline{ID: 8}))
	count, err = store.engine.Count(new(model.Config))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)
}
