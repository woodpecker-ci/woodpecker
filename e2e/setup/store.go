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

package setup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

// Fixtures holds the pre-seeded database records shared across all tests.
type Fixtures struct {
	Forge *model.Forge
	Owner *model.User
	Repo  *model.Repo
}

// newStore creates a fully-migrated in-memory sqlite store.
func newStore(ctx context.Context, t *testing.T) store.Store {
	t.Helper()

	s, err := datastore.NewEngine(&store.Opts{
		Driver: "sqlite3",
		Config: ":memory:",
		// MaxOpenConns=1 and MaxIdleConns=1 are required for in-memory sqlite:
		// without them the pool drops idle connections, destroying the in-memory
		// schema between calls and breaking migrations.
		XORM: store.XORM{
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
	require.NoError(t, err, "create in-memory store")

	require.NoError(t, s.Ping(), "ping store")
	require.NoError(t, s.Migrate(ctx, true), "migrate store")

	t.Cleanup(func() { _ = s.Close() })
	return s
}

// seedFixtures creates the minimal set of DB records every test needs:
// one Forge, one owner User, one Repo linked to both.
func seedFixtures(t *testing.T, s store.Store) *Fixtures {
	t.Helper()

	forge := &model.Forge{
		Type: model.ForgeTypeGitea,
		URL:  "https://forge.example.test",
	}
	require.NoError(t, s.ForgeCreate(forge), "seed forge")

	owner := &model.User{
		ForgeID:       forge.ID,
		ForgeRemoteID: "1",
		Login:         "test-owner",
		Email:         "owner@example.test",
	}
	require.NoError(t, s.CreateUser(owner), "seed user")

	repo := &model.Repo{
		ForgeID:       forge.ID,
		ForgeRemoteID: "1",
		UserID:        owner.ID,
		FullName:      "test-owner/test-repo",
		Owner:         "test-owner",
		Name:          "test-repo",
		Clone:         "https://forge.example.test/test-owner/test-repo.git",
		Branch:        "main",
		IsActive:      true,
		AllowPull:     true,
	}
	require.NoError(t, s.CreateRepo(repo), "seed repo")

	return &Fixtures{
		Forge: forge,
		Owner: owner,
		Repo:  repo,
	}
}
