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
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// testStoreDriverConfig returns the database driver and connection string used
// for tests. It defaults to in-memory sqlite but can be overridden via
// environment variables to run the same tests against mysql/postgres.
func testStoreDriverConfig() (driver, config string) {
	driver = "sqlite3"
	config = ":memory:"

	if os.Getenv("WOODPECKER_DATABASE_DRIVER") != "" {
		driver = os.Getenv("WOODPECKER_DATABASE_DRIVER")
		config = os.Getenv("WOODPECKER_DATABASE_DATASOURCE")
	}
	return driver, config
}

// NewTestStore creates a fully-migrated in-memory sqlite store for use in
// tests of other packages (e.g. server/api). It is gated behind the "test"
// build tag so it never ends up in release builds.
//
// The driver and connection string can be overridden via the
// WOODPECKER_DATABASE_DRIVER / WOODPECKER_DATABASE_DATASOURCE environment
// variables, with fallback to in-memory sqlite.
//
// The returned store is automatically closed on test cleanup.
func NewTestStore(t *testing.T) store.Store {
	t.Helper()

	driver, config := testStoreDriverConfig()

	s, err := NewEngine(&store.Opts{
		Driver: driver,
		Config: config,
		// MaxOpenConns=1 and MaxIdleConns=1 are required for in-memory sqlite:
		// without them the pool drops idle connections, destroying the in-memory
		// schema between calls and breaking migrations.
		XORM: store.XORM{
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
	require.NoError(t, err, "create test store")
	require.NoError(t, s.Ping(), "ping test store")
	require.NoError(t, s.Migrate(t.Context(), true), "migrate test store")

	t.Cleanup(func() { _ = s.Close() })
	return s
}
