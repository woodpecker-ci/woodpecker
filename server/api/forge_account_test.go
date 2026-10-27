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

package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestSyncOrg(t *testing.T) {
	s := newTestStore(t)
	user := &model.User{AccessToken: "token"}

	createOrg := func(t *testing.T, org *model.Org) *model.Org {
		t.Helper()
		org.ForgeID = 1
		require.NoError(t, s.OrgCreate(org))
		return org
	}
	sync := func(t *testing.T, f *forge_mocks.MockForge, remoteID model.ForgeRemoteID, name string) (*model.Org, error) {
		t.Helper()
		return syncOrg(t.Context(), s, f, user, &model.Org{ForgeID: 1, ForgeRemoteID: remoteID, Name: name})
	}
	oldAccount := func(f *forge_mocks.MockForge, remoteID model.ForgeRemoteID, org *model.Org, err error) {
		f.On("OrgByRemoteID", mock.Anything, user, remoteID, false).Return(org, err)
	}

	t.Run("create new org", func(t *testing.T) {
		org, err := sync(t, forge_mocks.NewMockForge(t), "1", "new")
		require.NoError(t, err)

		stored, err := s.OrgFindByRemoteID(1, "1")
		require.NoError(t, err)
		assert.Equal(t, org.ID, stored.ID)
		assert.Equal(t, "new", stored.Name)
	})

	t.Run("find by forge remote id and take new name", func(t *testing.T) {
		old := createOrg(t, &model.Org{ForgeRemoteID: "2", Name: "before", ConflictForgeRemoteID: "99"})

		org, err := sync(t, forge_mocks.NewMockForge(t), "2", "after")
		require.NoError(t, err)
		assert.Equal(t, old.ID, org.ID)

		stored, err := s.OrgGet(old.ID)
		require.NoError(t, err)
		assert.Equal(t, "after", stored.Name)
		assert.Empty(t, stored.ConflictForgeRemoteID)
	})

	t.Run("take over org stored without forge remote id", func(t *testing.T) {
		old := createOrg(t, &model.Org{Name: "legacy"})

		org, err := sync(t, forge_mocks.NewMockForge(t), "3", "legacy")
		require.NoError(t, err)
		assert.Equal(t, old.ID, org.ID)

		stored, err := s.OrgFindByRemoteID(1, "3")
		require.NoError(t, err)
		assert.Equal(t, old.ID, stored.ID)
	})

	t.Run("find by name if forge has no remote ids", func(t *testing.T) {
		old := createOrg(t, &model.Org{ForgeRemoteID: "4", Name: "addon"})

		org, err := sync(t, forge_mocks.NewMockForge(t), "", "addon")
		require.NoError(t, err)
		assert.Equal(t, old.ID, org.ID)
	})

	t.Run("rename org of renamed forge account", func(t *testing.T) {
		old := createOrg(t, &model.Org{ForgeRemoteID: "5", Name: "reused", ConflictForgeRemoteID: "99"})
		f := forge_mocks.NewMockForge(t)
		oldAccount(f, "5", &model.Org{ForgeRemoteID: "5", Name: "moved"}, nil)

		org, err := sync(t, f, "6", "reused")
		require.NoError(t, err)
		assert.NotEqual(t, old.ID, org.ID)

		stored, err := s.OrgGet(old.ID)
		require.NoError(t, err)
		assert.Equal(t, "moved", stored.Name)
		assert.Empty(t, stored.ConflictForgeRemoteID)
	})

	for _, tc := range []struct {
		name     string
		remoteID model.ForgeRemoteID
		account  *model.Org
		err      error
	}{
		{name: "old forge account is gone", remoteID: "7", err: forge_types.ErrOrgNotFound},
		{name: "forge can not look up old account", remoteID: "8", err: forge_types.ErrNotImplemented},
		{name: "old forge account still has the name", remoteID: "9", account: &model.Org{ForgeRemoteID: "9", Name: "Taken-9"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := "taken-" + string(tc.remoteID)
			old := createOrg(t, &model.Org{ForgeRemoteID: tc.remoteID, Name: name})
			f := forge_mocks.NewMockForge(t)
			oldAccount(f, tc.remoteID, tc.account, tc.err)

			_, err := sync(t, f, "new-"+tc.remoteID, name)
			assert.ErrorIs(t, err, errNameConflict)

			stored, err := s.OrgGet(old.ID)
			require.NoError(t, err)
			assert.Equal(t, name, stored.Name)
			assert.Equal(t, "new-"+tc.remoteID, stored.ConflictForgeRemoteID)
			_, err = s.OrgFindByRemoteID(1, "new-"+tc.remoteID)
			assert.ErrorIs(t, err, types.ErrRecordNotExist)
		})
	}

	t.Run("forge error marks no conflict", func(t *testing.T) {
		old := createOrg(t, &model.Org{ForgeRemoteID: "10", Name: "flaky"})
		f := forge_mocks.NewMockForge(t)
		oldAccount(f, "10", nil, errors.New("forge unavailable"))

		_, err := sync(t, f, "11", "flaky")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, errNameConflict)

		stored, err := s.OrgGet(old.ID)
		require.NoError(t, err)
		assert.Empty(t, stored.ConflictForgeRemoteID)
	})
}

func TestReleaseUserLogin(t *testing.T) {
	s := newTestStore(t)
	forgeUser := &model.User{ForgeID: 1, ForgeRemoteID: "2", Login: "jane", AccessToken: "token"}

	t.Run("rename user of renamed forge account", func(t *testing.T) {
		old := &model.User{ForgeID: 1, ForgeRemoteID: "1", Login: "jane", Hash: "1"}
		require.NoError(t, s.CreateUser(old))
		f := forge_mocks.NewMockForge(t)
		f.On("OrgByRemoteID", mock.Anything, forgeUser, model.ForgeRemoteID("1"), true).Return(&model.Org{Name: "jane-old"}, nil)

		require.NoError(t, releaseUserLogin(t.Context(), s, f, forgeUser, old))

		stored, err := s.GetUser(old.ID)
		require.NoError(t, err)
		assert.Equal(t, "jane-old", stored.Login)
	})

	t.Run("mark user of gone forge account", func(t *testing.T) {
		old := &model.User{ForgeID: 1, ForgeRemoteID: "3", Login: "jane", Hash: "3"}
		require.NoError(t, s.CreateUser(old))
		f := forge_mocks.NewMockForge(t)
		f.On("OrgByRemoteID", mock.Anything, forgeUser, model.ForgeRemoteID("3"), true).Return(nil, forge_types.ErrOrgNotFound)

		assert.ErrorIs(t, releaseUserLogin(t.Context(), s, f, forgeUser, old), errNameConflict)

		stored, err := s.GetUser(old.ID)
		require.NoError(t, err)
		assert.Equal(t, "jane", stored.Login)
		assert.EqualValues(t, "2", stored.ConflictForgeRemoteID)
	})
}
