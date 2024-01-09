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
	"fmt"
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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

	g := goblin.Goblin(t)
	g.Describe("Pipelines", func() {
		g.Before(func() {
			_, err := store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
			g.Assert(store.CreateRepo(repo)).IsNil()
		})
		g.After(func() {
			_, err := store.engine.Exec("DELETE FROM repos")
			g.Assert(err).IsNil()
		})

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			_, err := store.engine.Exec("DELETE FROM pipelines")
			g.Assert(err).IsNil()
			_, err = store.engine.Exec("DELETE FROM steps")
			g.Assert(err).IsNil()
		})

		g.It("Should Fail early when the repo is not existing", func() {
			pipeline := model.Pipeline{
				RepoID: 100,
				Status: model.StatusSuccess,
			}
			err := store.CreatePipeline(&pipeline)
			g.Assert(err).IsNotNil()

			count, err := store.GetPipelineCount()
			g.Assert(err).IsNil()
			g.Assert(count == 0).IsTrue()
			fmt.Println("GOT COUNT", count)
		})

		g.It("Should Post a Pipeline", func() {
			pipeline := model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			err := store.CreatePipeline(&pipeline)
			g.Assert(err).IsNil()
			g.Assert(pipeline.ID != 0).IsTrue()
			g.Assert(pipeline.Number).Equal(int64(1))
			g.Assert(pipeline.Commit).Equal("85f8c029b902ed9400bc600bac301a0aadb144ac")

			count, err := store.GetPipelineCount()
			g.Assert(err).IsNil()
			g.Assert(count > 0).IsTrue()
			fmt.Println("GOT COUNT", count)
		})

		g.It("Should Put a Pipeline", func() {
			pipeline := model.Pipeline{
				RepoID: repo.ID,
				Number: 5,
				Status: model.StatusSuccess,
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			err := store.CreatePipeline(&pipeline)
			g.Assert(err).IsNil()
			pipeline.Status = model.StatusRunning
			err1 := store.UpdatePipeline(&pipeline)
			GetPipeline, err2 := store.GetPipeline(pipeline.ID)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(pipeline.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline.Status).Equal(GetPipeline.Status)
			g.Assert(pipeline.Number).Equal(GetPipeline.Number)
		})

		g.It("Should Get a Pipeline", func() {
			pipeline := model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
			}
			err := store.CreatePipeline(&pipeline, []*model.Step{}...)
			g.Assert(err).IsNil()
			GetPipeline, err := store.GetPipeline(pipeline.ID)
			g.Assert(err).IsNil()
			g.Assert(pipeline.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline.Status).Equal(GetPipeline.Status)
		})

		g.It("Should Get a Pipeline by Number", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			GetPipeline, err3 := store.GetPipelineNumber(&model.Repo{ID: 1}, pipeline2.Number)
			g.Assert(err3).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
		})

		g.It("Should Get a Pipeline by Ref", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/5",
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/6",
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			GetPipeline, err3 := store.GetPipelineRef(&model.Repo{ID: 1}, "refs/pull/6")
			g.Assert(err3).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
			g.Assert(pipeline2.Ref).Equal(GetPipeline.Ref)
		})

		g.It("Should Get a Pipeline by Ref", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/5",
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Ref:    "refs/pull/6",
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			GetPipeline, err3 := store.GetPipelineRef(&model.Repo{ID: 1}, "refs/pull/6")
			g.Assert(err3).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
			g.Assert(pipeline2.Ref).Equal(GetPipeline.Ref)
		})

		g.It("Should Get a Pipeline by Commit", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusPending,
				Branch: "dev",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			GetPipeline, err3 := store.GetPipelineCommit(&model.Repo{ID: 1}, pipeline2.Commit, pipeline2.Branch)
			g.Assert(err3).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
			g.Assert(pipeline2.Commit).Equal(GetPipeline.Commit)
			g.Assert(pipeline2.Branch).Equal(GetPipeline.Branch)
		})

		g.It("Should Get the last Pipeline", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusFailure,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
				Event:  model.EventPush,
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
				Event:  model.EventPush,
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			GetPipeline, err3 := store.GetPipelineLast(&model.Repo{ID: 1}, pipeline2.Branch)
			g.Assert(err1).IsNil()
			g.Assert(err2).IsNil()
			g.Assert(err3).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
			g.Assert(pipeline2.Status).Equal(GetPipeline.Status)
			g.Assert(pipeline2.Branch).Equal(GetPipeline.Branch)
			g.Assert(pipeline2.Commit).Equal(GetPipeline.Commit)
		})

		g.It("Should Get the last Pipeline Before Pipeline N", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusFailure,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			pipeline3 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusRunning,
				Branch: "main",
				Commit: "85f8c029b902ed9400bc600bac301a0aadb144aa",
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			err3 := store.CreatePipeline(pipeline3, []*model.Step{}...)
			g.Assert(err3).IsNil()
			GetPipeline, err4 := store.GetPipelineLastBefore(&model.Repo{ID: 1}, pipeline3.Branch, pipeline3.ID)
			g.Assert(err4).IsNil()
			g.Assert(pipeline2.ID).Equal(GetPipeline.ID)
			g.Assert(pipeline2.RepoID).Equal(GetPipeline.RepoID)
			g.Assert(pipeline2.Number).Equal(GetPipeline.Number)
			g.Assert(pipeline2.Status).Equal(GetPipeline.Status)
			g.Assert(pipeline2.Branch).Equal(GetPipeline.Branch)
			g.Assert(pipeline2.Commit).Equal(GetPipeline.Commit)
		})

		g.It("Should get recent pipelines", func() {
			pipeline1 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusFailure,
			}
			pipeline2 := &model.Pipeline{
				RepoID: repo.ID,
				Status: model.StatusSuccess,
			}
			err1 := store.CreatePipeline(pipeline1, []*model.Step{}...)
			g.Assert(err1).IsNil()
			err2 := store.CreatePipeline(pipeline2, []*model.Step{}...)
			g.Assert(err2).IsNil()
			pipelines, err3 := store.GetPipelineList(&model.Repo{ID: 1}, &model.ListOptions{Page: 1, PerPage: 50})
			g.Assert(err3).IsNil()
			g.Assert(len(pipelines)).Equal(2)
			g.Assert(pipelines[0].ID).Equal(pipeline2.ID)
			g.Assert(pipelines[0].RepoID).Equal(pipeline2.RepoID)
			g.Assert(pipelines[0].Status).Equal(pipeline2.Status)
		})
	})
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
