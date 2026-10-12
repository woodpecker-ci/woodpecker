// Copyright 2023 Woodpecker Authors
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestWrapGet(t *testing.T) {
	err := wrapGet(false, nil)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)

	err = wrapGet(true, errors.New("test err"))
	assert.Equal(t, "TestWrapGet: test err", err.Error())
}

func TestWrapDelete(t *testing.T) {
	err := wrapDelete(0, nil)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)

	err = wrapDelete(1, errors.New("test err"))
	assert.Equal(t, "TestWrapDelete: test err", err.Error())
}

func TestWrapInsert(t *testing.T) {
	store, closer := newTestStore(t, new(model.Cron))
	defer closer()

	// test normal insert
	cron := &model.Cron{RepoID: 1, CreatorID: 1, Name: "sync", NextExec: 10000, Schedule: "@every 1h"}
	assert.NoError(t, wrapInsert(store.engine.Insert(cron)))

	// test insert witch should fail because of unique constraint
	assert.ErrorIs(t, wrapInsert(store.engine.Insert(cron)), types.ErrInsertDuplicateDetected)

	// The store above only exercises the sqlite wording. Cover the other drivers too: callers rely on
	// errors.Is(err, ErrInsertDuplicateDetected) instead of matching driver strings themselves, so a
	// pattern missing here silently breaks them (e.g. the retry in CreatePipeline, #6067).
	for _, driverErr := range []string{
		`pq: duplicate key value violates unique constraint "UQE_pipelines_s"`, // postgres
		"Error 1062: Duplicate entry '1-2' for key 'UQE_pipelines_s'",          // mysql
		"UNIQUE violation",
		"unique constraint",
	} {
		assert.ErrorIs(t, wrapInsert(0, errors.New(driverErr)), types.ErrInsertDuplicateDetected, driverErr)
	}

	// Everything else passes through untouched.
	other := errors.New("connection refused")
	assert.Equal(t, other, wrapInsert(0, other))
}
