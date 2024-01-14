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

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestConfig(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.PipelineConfig), new(model.Pipeline), new(model.Repo))
	defer closer()

	var (
		data = "pipeline: [ { image: golang, commands: [ go build, go test ] } ]"
		hash = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
	)

	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(repo))

	config := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   hash,
		Name:   "default",
	}
	assert.NoError(t, store.ConfigCreate(config))

	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusRunning,
		Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
	}
	assert.NoError(t, store.CreatePipeline(pipeline))

	assert.NoError(t, store.PipelineConfigCreate(
		&model.PipelineConfig{
			ConfigID:   config.ID,
			PipelineID: pipeline.ID,
		},
	))

	config, err := store.ConfigFindIdentical(repo.ID, hash)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, config.ID)
	assert.Equal(t, repo.ID, config.RepoID)
	assert.Equal(t, data, string(config.Data))
	assert.Equal(t, hash, config.Hash)
	assert.Equal(t, "default", config.Name)

	loaded, err := store.ConfigsForPipeline(pipeline.ID)
	assert.NoError(t, err)
	assert.Equal(t, config.ID, loaded[0].ID)
}

func TestConfigApproved(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.PipelineConfig), new(model.Pipeline), new(model.Repo))
	defer closer()

	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(repo))

	var (
		data            = "pipeline: [ { image: golang, commands: [ go build, go test ] } ]"
		hash            = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
		pipelineBlocked = &model.Pipeline{
			RepoID: repo.ID,
			Status: model.StatusBlocked,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
		pipelinePending = &model.Pipeline{
			RepoID: repo.ID,
			Status: model.StatusPending,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
		pipelineRunning = &model.Pipeline{
			RepoID: repo.ID,
			Status: model.StatusRunning,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
	)

	assert.NoError(t, store.CreatePipeline(pipelineBlocked))
	assert.NoError(t, store.CreatePipeline(pipelinePending))
	conf := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   hash,
	}
	assert.NoError(t, store.ConfigCreate(conf))
	pipelineConfig := &model.PipelineConfig{
		ConfigID:   conf.ID,
		PipelineID: pipelineBlocked.ID,
	}
	assert.NoError(t, store.PipelineConfigCreate(pipelineConfig))

	approved, err := store.ConfigFindApproved(conf)
	if !assert.NoError(t, err) {
		return
	}
	assert.False(t, approved, "want config not approved when blocked or pending.")

	assert.NoError(t, store.CreatePipeline(pipelineRunning))
	conf2 := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   "xxx",
	}
	assert.NoError(t, store.ConfigCreate(conf2))
	pipelineConfig2 := &model.PipelineConfig{
		ConfigID:   conf2.ID,
		PipelineID: pipelineRunning.ID,
	}
	assert.NoError(t, store.PipelineConfigCreate(pipelineConfig2))

	approved, err = store.ConfigFindApproved(conf2)
	assert.NoError(t, err)
	assert.True(t, approved)
}

func TestConfigIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.Step), new(model.Pipeline), new(model.Repo))
	defer closer()

	var (
		data = "pipeline: [ { image: golang, commands: [ go build, go test ] } ]"
		hash = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
	)

	assert.NoError(t, store.ConfigCreate(
		&model.Config{
			RepoID: 2,
			Data:   []byte(data),
			Hash:   hash,
		},
	))

	// fail due to duplicate sha
	assert.Error(t, store.ConfigCreate(
		&model.Config{
			RepoID: 2,
			Data:   []byte(data),
			Hash:   hash,
		},
	))
}
