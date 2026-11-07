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
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

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

	// Lock conflicts of every supported driver must surface as ErrDatabaseLocked.
	for _, driverErr := range []error{
		errors.New("database is locked"),
		errors.New("database table is locked"),
		&mysql.MySQLError{Number: 1213}, // Deadlock victim.
		&mysql.MySQLError{Number: 1205}, // Lock wait timeout.
		&mysql.MySQLError{Number: 1020}, // MariaDB snapshot isolation conflict.
		&pq.Error{Code: "40P01"},        // Deadlock victim.
		&pq.Error{Code: "40001"},        // Serialization failure.
	} {
		assert.ErrorIs(t, wrapUpdate(0, driverErr), types.ErrDatabaseLocked, driverErr)
	}

	// Everything else passes through untouched.
	for _, other := range []error{
		errors.New("connection refused"),
		&mysql.MySQLError{Number: 1064}, // Syntax error.
		&pq.Error{Code: "42601"},        // Syntax error.
	} {
		assert.Same(t, other, wrapUpdate(0, other))
	}
}

// newLockTestStore returns a store with the rows "a" and "b" on the given database driver,
// and skips the test for any other driver.
func newLockTestStore(t *testing.T, driver string) *storage {
	t.Helper()
	if configured, _ := testDriverConfig(); configured != driver {
		t.Skipf("needs WOODPECKER_DATABASE_DRIVER=%s", driver)
	}
	store, closer := newTestStore(t, new(model.ServerConfig))
	t.Cleanup(closer)
	// Concurrent transactions need a connection each.
	store.engine.SetMaxOpenConns(4)
	store.engine.SetMaxIdleConns(4)
	_, err := store.engine.Insert(&model.ServerConfig{Key: "a"}, &model.ServerConfig{Key: "b"})
	require.NoError(t, err)
	return store
}

// lockRow writes the row, which keeps it locked until the transaction ends.
func lockRow(ctx context.Context, sess *xorm.Session, key string) error {
	_, err := sess.Context(ctx).ID(key).Update(&model.ServerConfig{Value: "locked"})
	return err
}

// deadlock lets two transactions lock the rows in opposite order and
// returns the error of the one the database picked as deadlock victim.
func deadlock(t *testing.T, store *storage) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	first, second := store.engine.NewSession(), store.engine.NewSession()
	defer first.Close()
	defer second.Close()
	require.NoError(t, first.Begin())
	require.NoError(t, second.Begin())
	require.NoError(t, lockRow(ctx, first, "a"))
	require.NoError(t, lockRow(ctx, second, "b"))

	errs := make(chan error, 2)
	go func() { errs <- lockRow(ctx, first, "b") }()
	go func() { errs <- lockRow(ctx, second, "a") }()
	// The survivor continues once the victim is rolled back and returns nil.
	return errors.Join(<-errs, <-errs)
}

// lockWaitTimeout lets a transaction wait for a row that another one keeps locked.
func lockWaitTimeout(t *testing.T, store *storage) error {
	t.Helper()
	first, second := store.engine.NewSession(), store.engine.NewSession()
	defer first.Close()
	defer second.Close()
	require.NoError(t, first.Begin())
	require.NoError(t, second.Begin())
	_, err := second.Exec("SET SESSION innodb_lock_wait_timeout = 1")
	require.NoError(t, err)
	require.NoError(t, lockRow(t.Context(), first, "a"))
	return lockRow(t.Context(), second, "a")
}

// serializationFailure lets a repeatable-read transaction write a row
// that another transaction changed after the first one started reading.
func serializationFailure(t *testing.T, store *storage) error {
	t.Helper()
	sess := store.engine.NewSession()
	defer sess.Close()
	require.NoError(t, sess.Begin())
	if store.engine.Dialect().URI().DBType == schemas.POSTGRES {
		// MySQL/MariaDB default to repeatable read, PostgreSQL to read committed.
		_, err := sess.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")
		require.NoError(t, err)
	}
	_, err := sess.ID("a").Get(new(model.ServerConfig))
	require.NoError(t, err)
	_, err = store.engine.ID("a").Update(&model.ServerConfig{Value: "other"})
	require.NoError(t, err)
	return lockRow(t.Context(), sess, "a")
}

// snapshotIsolationConflict is a serializationFailure on MariaDB, which reports it only
// with innodb_snapshot_isolation (the default since MariaDB 11.6). MySQL does not have it.
func snapshotIsolationConflict(t *testing.T, store *storage) error {
	t.Helper()
	var enabled bool
	if has, err := store.engine.SQL("SELECT @@innodb_snapshot_isolation").Get(&enabled); err != nil || !has || !enabled {
		t.Skip("needs MariaDB with innodb_snapshot_isolation")
	}
	return serializationFailure(t, store)
}

// The drivers must return the lock errors that isErrDatabaseLocked looks for.
func TestDatabaseLockErrors(t *testing.T) {
	mysqlNumber := func(number uint16) func(error) bool {
		return func(err error) bool {
			var mysqlErr *mysql.MySQLError
			return errors.As(err, &mysqlErr) && mysqlErr.Number == number
		}
	}
	postgresCode := func(code string) func(error) bool {
		return func(err error) bool {
			var pqErr *pq.Error
			return errors.As(err, &pqErr) && pqErr.SQLState() == code
		}
	}

	tests := []struct {
		name    string
		driver  string
		trigger func(*testing.T, *storage) error
		isError func(error) bool
	}{
		{"mysql deadlock", DriverMysql, deadlock, mysqlNumber(1213)},
		{"mysql lock wait timeout", DriverMysql, lockWaitTimeout, mysqlNumber(1205)},
		{"mariadb snapshot isolation conflict", DriverMysql, snapshotIsolationConflict, mysqlNumber(1020)},
		{"postgres deadlock", DriverPostgres, deadlock, postgresCode("40P01")},
		{"postgres serialization failure", DriverPostgres, serializationFailure, postgresCode("40001")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trigger(t, newLockTestStore(t, tt.driver))
			require.Error(t, err)
			assert.True(t, tt.isError(err), "unexpected error %v", err)
			assert.ErrorIs(t, wrapUpdate(0, err), types.ErrDatabaseLocked)
		})
	}
}

// A deadlock victim must be retried instead of failing the write.
func TestRetryWriteDeadlock(t *testing.T) {
	for _, driver := range []string{DriverMysql, DriverPostgres} {
		t.Run(driver, func(t *testing.T) {
			store := newLockTestStore(t, driver)

			var attempts atomic.Int32
			var bothLocked sync.WaitGroup
			bothLocked.Add(2)
			write := func(first, second string) error {
				waited := false
				return retryWrite(func() error {
					attempts.Add(1)
					sess := store.engine.NewSession()
					defer sess.Close()
					if err := sess.Begin(); err != nil {
						return err
					}
					if err := wrapUpdate(0, lockRow(t.Context(), sess, first)); err != nil {
						return err
					}
					if !waited {
						// Only the first attempts must deadlock, the retry has to succeed.
						waited = true
						bothLocked.Done()
						bothLocked.Wait()
					}
					if err := wrapUpdate(0, lockRow(t.Context(), sess, second)); err != nil {
						return err
					}
					return wrapUpdate(0, sess.Commit())
				})
			}

			var wg sync.WaitGroup
			var errA, errB error
			wg.Go(func() { errA = write("a", "b") })
			wg.Go(func() { errB = write("b", "a") })
			wg.Wait()

			require.NoError(t, errA)
			require.NoError(t, errB)
			assert.EqualValues(t, 3, attempts.Load(), "the deadlock victim must run once more")
		})
	}
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
