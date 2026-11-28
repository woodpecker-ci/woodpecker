// Copyright 2026 Woodpecker Authors
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

//go:build test

package datastore

import (
	"crypto/rand"
	"database/sql"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// NewTestStore creates a fully-migrated store for use in tests of other packages (e.g. server/api),
// see TestStoreOpts.
//
// The returned store is automatically closed on test cleanup.
func NewTestStore(t *testing.T) store.Store {
	t.Helper()

	s, err := NewEngine(TestStoreOpts(t))
	require.NoError(t, err, "create test store")
	require.NoError(t, s.Ping(), "ping test store")
	require.NoError(t, s.Migrate(t.Context(), true), "migrate test store")

	t.Cleanup(func() { _ = s.Close() })
	return s
}

// The pool of the server by default.
const (
	defaultMaxOpenConns    = 100
	defaultMaxIdleConns    = 2
	defaultConnMaxLifetime = 3 * time.Second
)

// TestStoreOpts returns the options of a test store: in-memory sqlite, or a new database on the
// server WOODPECKER_DATABASE_DRIVER and WOODPECKER_DATABASE_DATASOURCE point to, dropped on test
// cleanup, with the pool of the server.
func TestStoreOpts(t *testing.T) *store.Opts {
	t.Helper()

	driver := os.Getenv("WOODPECKER_DATABASE_DRIVER")
	if driver == "" || driver == "sqlite3" {
		return &store.Opts{
			Driver: "sqlite3",
			Config: ":memory:",
			// MaxOpenConns=1 and MaxIdleConns=1 are required for in-memory sqlite:
			// without them the pool drops idle connections, destroying the in-memory
			// schema between calls and breaking migrations.
			XORM: store.XORM{
				MaxOpenConns: 1,
				MaxIdleConns: 1,
			},
		}
	}
	return &store.Opts{
		Driver: driver,
		Config: newTestDatabase(t, driver, os.Getenv("WOODPECKER_DATABASE_DATASOURCE")),
		XORM: store.XORM{
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    defaultMaxIdleConns,
			ConnMaxLifetime: defaultConnMaxLifetime,
		},
	}
}

var pgDBName = regexp.MustCompile(`dbname=\S+`)

// newTestDatabase creates a database on the server of datasource, drops it on test cleanup
// and returns the datasource of it.
func newTestDatabase(t *testing.T, driver, datasource string) string {
	t.Helper()
	name := "wp_test_" + strings.ToLower(rand.Text()[:12])
	admin, err := sql.Open(driver, datasource)
	require.NoError(t, err)
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	// registered first, so it runs after the store is closed
	t.Cleanup(func() {
		drop := "DROP DATABASE " + name
		if driver == "postgres" {
			drop += " WITH (FORCE)"
		}
		_, err := admin.Exec(drop)
		require.NoError(t, err)
		_ = admin.Close()
	})

	if driver == "postgres" {
		return pgDBName.ReplaceAllString(datasource, "dbname="+name)
	}
	cfg, err := mysql.ParseDSN(datasource)
	require.NoError(t, err)
	cfg.DBName = name
	return cfg.FormatDSN()
}
