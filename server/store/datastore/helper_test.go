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
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

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

	// Database locked errors must surface as ErrDatabaseLocked before the c == 0 check.
	for _, driverErr := range []string{
		// sqlite common errors
		"database is locked",
		"database table is locked",
	} {
		assert.ErrorIs(t, wrapDelete(0, errors.New(driverErr)), types.ErrDatabaseLocked, driverErr)
	}

	err = wrapDelete(1, errors.New("test err"))
	assert.Equal(t, "TestWrapDelete: test err", err.Error())
}

func TestWrapUpdate(t *testing.T) {
	assert.NoError(t, wrapUpdate(0, nil))

	// Database locked errors must surface as ErrDatabaseLocked.
	for _, driverErr := range []string{
		// sqlite common errors
		"database is locked",
		"database table is locked",
	} {
		assert.ErrorIs(t, wrapUpdate(0, errors.New(driverErr)), types.ErrDatabaseLocked, driverErr)
	}

	// Everything else passes through untouched.
	other := errors.New("connection refused")
	assert.Equal(t, other, wrapUpdate(0, other))
}

func TestWrapInsert(t *testing.T) {
	store, closer := newTestStore(t, new(model.Cron))
	defer closer()

	// test normal insert
	cron := &model.Cron{RepoID: 1, CreatorID: 1, Name: "sync", NextExec: 10000, Schedule: "@every 1h"}
	assert.NoError(t, wrapInsert(store.engine.Insert(cron)))

	// test insert witch should fail because of unique constraint
	err := wrapInsert(store.engine.Insert(&model.Cron{RepoID: 1, CreatorID: 1, Name: "sync", NextExec: 10000, Schedule: "@every 1h"}))
	assert.ErrorIs(t, err, types.ErrInsertDuplicateDetected)
	// the driver error must stay visible, it names the violated constraint;
	// every driver words it its own way, but all name the crons table or index
	assert.ErrorContains(t, err, "crons")

	// The store above only exercises the sqlite wording. Cover the other drivers too: callers rely on
	// errors.Is(err, ErrInsertDuplicateDetected) instead of matching driver strings themselves, so a
	// pattern missing here silently breaks them (e.g. the retry in CreatePipeline, #6067).
	for _, driverErr := range []string{
		`pq: duplicate key value violates unique constraint "UQE_pipelines_s"`, // postgres
		"Error 1062: Duplicate entry '1-2' for key 'UQE_pipelines_s'",          // mysql
		"UNIQUE violation",
		"unique constraint",
	} {
		err := wrapInsert(0, errors.New(driverErr))
		assert.ErrorIs(t, err, types.ErrInsertDuplicateDetected, driverErr)
		assert.ErrorContains(t, err, driverErr)
	}

	// Database locked errors must surface as ErrDatabaseLocked so the retry sees them as retryable.
	for _, driverErr := range []string{
		// sqlite common errors
		"database is locked",
		"database table is locked",
	} {
		assert.ErrorIs(t, wrapInsert(0, errors.New(driverErr)), types.ErrDatabaseLocked, driverErr)
	}

	// Everything else passes through untouched.
	other := errors.New("connection refused")
	assert.Equal(t, other, wrapInsert(0, other))
}

// TestWrapErrDatabaseLocked triggers a SQLite "database is locked" to assert
// wrapInsert, wrapUpdate, and wrapDelete surface it as ErrDatabaseLocked.
func TestWrapErrDatabaseLocked(t *testing.T) {
	dsn := "file::memory:?cache=shared&_busy_timeout=0"

	eng1, err := xorm.NewEngine("sqlite3", dsn)
	require.NoError(t, err)
	eng1.SetMaxOpenConns(1)
	eng1.SetMaxIdleConns(1)
	defer eng1.Close()

	eng2, err := xorm.NewEngine("sqlite3", dsn)
	require.NoError(t, err)
	eng2.SetMaxOpenConns(1)
	eng2.SetMaxIdleConns(1)
	defer eng2.Close()

	require.NoError(t, eng1.Sync(&model.Cron{}))

	sess1 := eng1.NewSession() // holds the only write conn
	defer sess1.Close()
	require.NoError(t, sess1.Begin())
	_, err = sess1.Insert(&model.Cron{RepoID: 1, CreatorID: 1, Name: "locked", NextExec: 10000, Schedule: "@every 1h"})
	require.NoError(t, err)

	// all these should fail with "database is locked".
	assert.ErrorIs(t, wrapInsert(eng2.Insert(&model.Cron{RepoID: 2, CreatorID: 1, Name: "second", NextExec: 10000, Schedule: "@every 1h"})), types.ErrDatabaseLocked)
	_, err = eng2.ID(1).AllCols().Update(&model.Cron{RepoID: 1, CreatorID: 1, Name: "updated", NextExec: 10000, Schedule: "@every 1h"})
	assert.ErrorIs(t, wrapUpdate(0, err), types.ErrDatabaseLocked)
	assert.ErrorIs(t, wrapDelete(eng2.ID(1).Delete(new(model.Cron))), types.ErrDatabaseLocked)

	require.NoError(t, sess1.Commit())
}

func TestRetryWrite(t *testing.T) {
	// Retries on ErrDatabaseLocked, then succeeds.
	calls := 0
	err := retryWrite(func() error {
		calls++
		if calls < 3 {
			return types.ErrDatabaseLocked
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, calls)

	// Does not retry on other errors.
	calls = 0
	other := errors.New("connection refused")
	err = retryWrite(func() error {
		calls++
		return other
	})
	assert.ErrorIs(t, err, other)
	assert.Equal(t, 1, calls)
}
