// Copyright 2022 Woodpecker Authors
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

package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCreatePipeline(t *testing.T) {
	_manager := manager_mocks.NewMockManager(t)
	_forge := forge_mocks.NewMockForge(t)
	store := store_mocks.NewMockStore(t)
	ctx := t.Context()

	repoUser := &model.User{
		ID:    1,
		Login: "user1",
	}
	repo1 := &model.Repo{
		ID:       1,
		Name:     "repo1",
		Owner:    "owner1",
		FullName: "repo1/owner1",
		Branch:   "default",
		UserID:   repoUser.ID,
	}

	// mock things
	store.On("GetRepo", mock.Anything).Return(repo1, nil)
	store.On("GetUser", mock.Anything).Return(repoUser, nil)
	_forge.On("BranchHead", mock.Anything, repoUser, repo1, "default").Return(&model.Commit{
		ForgeURL: "https://example.com/sha1",
		SHA:      "sha1",
	}, nil)
	_manager.On("ForgeFromRepo", repo1).Return(_forge, nil)
	server.Config.Services.Manager = _manager

	_, pipeline, err := CreatePipeline(ctx, store, &model.Cron{
		Name: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, &model.Pipeline{
		Branch: "default",
		Commit: &model.Commit{
			ForgeURL: "https://example.com/sha1",
			SHA:      "sha1",
		},
		Event:    "cron",
		ForgeURL: "https://example.com/sha1",
		Ref:      "refs/heads/default",
		Cron:     "test",
	}, pipeline)
}

func TestCalcNewNext(t *testing.T) {
	now := time.Unix(1661962369, 0)
	_, err := CalcNewNext("", now)
	assert.Error(t, err)

	schedule, err := CalcNewNext("@every 5m", now)
	assert.NoError(t, err)
	assert.EqualValues(t, 1661962669, schedule.Unix())
}
