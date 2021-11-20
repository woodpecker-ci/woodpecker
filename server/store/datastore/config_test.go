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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestConfig(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.BuildConfig), new(model.Build), new(model.Repo))
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
	if err := store.CreateRepo(repo); err != nil {
		t.Errorf("Unexpected error: insert repo: %s", err)
		return
	}

	config := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   hash,
		Name:   "default",
	}
	if err := store.ConfigCreate(config); err != nil {
		t.Errorf("Unexpected error: insert config: %s", err)
		return
	}

	build := &model.Build{
		RepoID: repo.ID,
		Status: model.StatusRunning,
		Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
	}
	if err := store.CreateBuild(build); err != nil {
		t.Errorf("Unexpected error: insert build: %s", err)
		return
	}

	if err := store.BuildConfigCreate(
		&model.BuildConfig{
			ConfigID: config.ID,
			BuildID:  build.ID,
		},
	); err != nil {
		t.Errorf("Unexpected error: insert build config: %s", err)
		return
	}

	config, err := store.ConfigFindIdentical(repo.ID, hash)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := config.ID, int64(1); got != want {
		t.Errorf("Want config id %d, got %d", want, got)
	}
	if got, want := config.RepoID, repo.ID; got != want {
		t.Errorf("Want config repo id %d, got %d", want, got)
	}
	if got, want := string(config.Data), data; got != want {
		t.Errorf("Want config data %s, got %s", want, got)
	}
	if got, want := config.Hash, hash; got != want {
		t.Errorf("Want config hash %s, got %s", want, got)
	}
	if got, want := config.Name, "default"; got != want {
		t.Errorf("Want config name %s, got %s", want, got)
	}

	loaded, err := store.ConfigsForBuild(build.ID)
	if err != nil {
		t.Errorf("Want config by id, got error %q", err)
		return
	}
	if got, want := loaded[0].ID, config.ID; got != want {
		t.Errorf("Want config by id %d, got %d", want, got)
	}
}

func TestConfigApproved(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.BuildConfig), new(model.Build), new(model.Repo))
	defer closer()

	repo := &model.Repo{
		UserID:   1,
		FullName: "bradrydzewski/test",
		Owner:    "bradrydzewski",
		Name:     "test",
	}
	if err := store.CreateRepo(repo); err != nil {
		t.Errorf("Unexpected error: insert repo: %s", err)
		return
	}

	var (
		data         = "pipeline: [ { image: golang, commands: [ go build, go test ] } ]"
		hash         = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
		buildBlocked = &model.Build{
			RepoID: repo.ID,
			Status: model.StatusBlocked,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
		buildPending = &model.Build{
			RepoID: repo.ID,
			Status: model.StatusPending,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
		buildRunning = &model.Build{
			RepoID: repo.ID,
			Status: model.StatusRunning,
			Commit: "85f8c029b902ed9400bc600bac301a0aadb144ac",
		}
	)

	if err := store.CreateBuild(buildBlocked); err != nil {
		t.Errorf("Unexpected error: insert build: %s", err)
		return
	}
	if err := store.CreateBuild(buildPending); err != nil {
		t.Errorf("Unexpected error: insert build: %s", err)
		return
	}
	conf := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   hash,
	}
	if err := store.ConfigCreate(conf); err != nil {
		t.Errorf("Unexpected error: insert config: %s", err)
		return
	}
	buildConfig := &model.BuildConfig{
		ConfigID: conf.ID,
		BuildID:  buildBlocked.ID,
	}
	if err := store.BuildConfigCreate(buildConfig); err != nil {
		t.Errorf("Unexpected error: insert build_config: %s", err)
		return
	}

	approved, err := store.ConfigFindApproved(conf)
	if !assert.NoError(t, err) {
		return
	}
	if approved != false {
		t.Errorf("Want config not approved, when blocked or pending.")
		return
	}

	assert.NoError(t, store.CreateBuild(buildRunning))
	conf2 := &model.Config{
		RepoID: repo.ID,
		Data:   []byte(data),
		Hash:   "xxx",
	}
	if err := store.ConfigCreate(conf2); err != nil {
		t.Errorf("Unexpected error: insert config: %s", err)
		return
	}
	buildConfig2 := &model.BuildConfig{
		ConfigID: conf2.ID,
		BuildID:  buildRunning.ID,
	}
	if err := store.BuildConfigCreate(buildConfig2); err != nil {
		t.Errorf("Unexpected error: insert config: %s", err)
		return
	}

	if approved, err := store.ConfigFindApproved(conf2); approved != true || err != nil {
		t.Errorf("Want config approved, when running. %v", err)
		return
	}
}

func TestConfigIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Config), new(model.Proc), new(model.Build), new(model.Repo))
	defer closer()

	var (
		data = "pipeline: [ { image: golang, commands: [ go build, go test ] } ]"
		hash = "8d8647c9aa90d893bfb79dddbe901f03e258588121e5202632f8ae5738590b26"
	)

	if err := store.ConfigCreate(
		&model.Config{
			RepoID: 2,
			Data:   []byte(data),
			Hash:   hash,
		},
	); err != nil {
		t.Errorf("Unexpected error: insert config: %s", err)
		return
	}

	// fail due to duplicate sha
	if err := store.ConfigCreate(
		&model.Config{
			RepoID: 2,
			Data:   []byte(data),
			Hash:   hash,
		},
	); err == nil {
		t.Errorf("Unexpected error: duplicate sha")
	}
}
