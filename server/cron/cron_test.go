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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/woodpecker-ci/woodpecker/server/model"
	mocks_remote "github.com/woodpecker-ci/woodpecker/server/remote/mocks"
	mocks_store "github.com/woodpecker-ci/woodpecker/server/store/mocks"
)

func TestCreateBuild(t *testing.T) {
	remote := mocks_remote.NewRemote(t)
	store := mocks_store.NewStore(t)
	ctx := context.Background()

	creator := &model.User{
		ID:    1,
		Login: "user1",
	}
	repo1 := &model.Repo{
		ID:       1,
		Name:     "repo1",
		Owner:    "owner1",
		FullName: "repo1/owner1",
		Branch:   "default",
	}

	// mock things
	store.On("GetRepo", mock.Anything).Return(repo1, nil)
	store.On("GetUser", mock.Anything).Return(creator, nil)
	remote.On("BranchHead", mock.Anything, creator, repo1, "default").Return("sha1", nil)

	_, build, err := createBuild(ctx, store, remote, &model.Cron{
		Name: "test",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, &model.Build{
		Event:   "cron",
		Commit:  "sha1",
		Branch:  "default",
		Ref:     "refs/heads/default",
		Message: "test",
		Sender:  "test",
	}, build)
}

func TestCalcNewNext(t *testing.T) {
	now := time.Unix(1661962369, 0)
	_, err := CalcNewNext("", now)
	assert.Error(t, err)

	schedule, err := CalcNewNext("@every 5m", now)
	assert.NoError(t, err)
	assert.EqualValues(t, 1661962669, schedule.Unix())
}
