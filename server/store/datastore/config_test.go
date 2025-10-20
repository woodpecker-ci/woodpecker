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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var (
	data = []byte("pipeline: [ { image: golang, commands: [ go build, go test ] } ]")
	hash = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
	name = "test"
)

func TestConfig(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.PipelineConfig), new(model.Pipeline), new(model.Repo))
	defer closer()

	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	assert.NoError(t, store.CreateRepo(repo))

	config := &model.Config{
		RepoID: repo.ID,
		Data:   data,
		Hash:   hash,
		Name:   name,
	}

	_, err := store.ConfigPersist(config)
	assert.NoError(t, err)

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

	foundConfig, err := store.configFindIdentical(store.engine.NewSession(), repo.ID, hash, name)
	assert.NoError(t, err)
	assert.EqualValues(t, config, foundConfig)

	loaded, err := store.ConfigsForPipeline(pipeline.ID)
	assert.NoError(t, err)
	assert.Equal(t, config.ID, loaded[0].ID)
}

func TestConfigPersist(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config))
	defer closer()

	conf1 := &model.Config{
		RepoID: 2,
		Data:   data,
		Hash:   hash,
		Name:   name,
	}
	conf2 := &model.Config{
		RepoID: 2,
		Data:   []byte("steps: [ { image: golang, commands: [ go generate ] } ]"),
		Name:   "generate",
	}

	conf1, err := store.ConfigPersist(conf1)
	assert.NoError(t, err)
	assert.EqualValues(t, hash, conf1.Hash)
	conf1secondInsert, err := store.ConfigPersist(conf1)
	assert.NoError(t, err)
	assert.EqualValues(t, conf1, conf1secondInsert)
	count, err := store.engine.Count(new(model.Config))
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)

	newConf2, err := store.ConfigPersist(conf2)
	assert.NoError(t, err)
	assert.EqualValues(t, "66f28f8d487a48aacf29d9feea13b0ab5dbb5025296b77a6addde93efcc4d82b", newConf2.Hash)
	count, err = store.engine.Count(new(model.Config))
	assert.NoError(t, err)
	assert.EqualValues(t, 2, count)

	// test for https://github.com/woodpecker-ci/woodpecker/issues/3093
	_, err = store.ConfigPersist(&model.Config{
		RepoID: 2,
		Data:   data,
		Hash:   hash,
		Name:   "some other",
	})
	assert.NoError(t, err)
	count, err = store.engine.Count(new(model.Config))
	assert.NoError(t, err)
	assert.EqualValues(t, 3, count)
}
