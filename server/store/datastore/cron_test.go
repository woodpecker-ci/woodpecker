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

package datastore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestCronCreate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Cron))
	defer closer()

	repo := &model.Repo{ID: 1, Name: "repo"}
	cron1 := &model.Cron{RepoID: repo.ID, CreatorID: 1, Name: "sync", NextExec: 10000, Schedule: "@every 1h"}
	assert.NoError(t, store.CronCreate(cron1))
	assert.NotEqualValues(t, 0, cron1.ID)

	// can not insert cron job with same repoID and title
	assert.Error(t, store.CronCreate(cron1))

	oldID := cron1.ID
	assert.NoError(t, store.CronDelete(repo, oldID))
	cron1.ID = 0
	assert.NoError(t, store.CronCreate(cron1))
	assert.NotEqual(t, oldID, cron1.ID)
}

func TestCronListNextExecute(t *testing.T) {
	store, closer := newTestStore(t, new(model.Cron))
	defer closer()

	jobs, err := store.CronListNextExecute(0, 10)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	now := time.Now().Unix()

	assert.NoError(t, store.CronCreate(&model.Cron{Schedule: "@every 1h", Name: "some", RepoID: 1, NextExec: now}))
	assert.NoError(t, store.CronCreate(&model.Cron{Schedule: "@every 1h", Name: "aaaa", RepoID: 1, NextExec: now}))
	assert.NoError(t, store.CronCreate(&model.Cron{Schedule: "@every 1h", Name: "bbbb", RepoID: 1, NextExec: now}))
	assert.NoError(t, store.CronCreate(&model.Cron{Schedule: "@every 1h", Name: "none", RepoID: 1, NextExec: now + 1000}))
	assert.NoError(t, store.CronCreate(&model.Cron{Schedule: "@every 1h", Name: "test", RepoID: 1, NextExec: now + 2000}))

	jobs, err = store.CronListNextExecute(now, 10)
	assert.NoError(t, err)
	assert.Len(t, jobs, 3)

	jobs, err = store.CronListNextExecute(now+1500, 10)
	assert.NoError(t, err)
	assert.Len(t, jobs, 4)
}

func TestCronGetLock(t *testing.T) {
	store, closer := newTestStore(t, new(model.Cron))
	defer closer()

	nonExistingJob := &model.Cron{ID: 1000, Name: "locales", NextExec: 10000}
	gotLock, err := store.CronGetLock(nonExistingJob, time.Now().Unix()+100)
	assert.NoError(t, err)
	assert.False(t, gotLock)

	cron1 := &model.Cron{RepoID: 1, Name: "some-title", NextExec: 10000, Schedule: "@every 1h"}
	assert.NoError(t, store.CronCreate(cron1))

	oldJob := *cron1
	gotLock, err = store.CronGetLock(cron1, cron1.NextExec+1000)
	assert.NoError(t, err)
	assert.True(t, gotLock)
	assert.NotEqualValues(t, oldJob.NextExec, cron1.NextExec)

	gotLock, err = store.CronGetLock(&oldJob, oldJob.NextExec+1000)
	assert.NoError(t, err)
	assert.False(t, gotLock)
	assert.EqualValues(t, oldJob.NextExec, oldJob.NextExec)
}
