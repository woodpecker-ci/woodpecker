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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
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

// remapForge installs a forge that answers the lookup of the old forge
// account with the given result.
func remapForge(t *testing.T, admin *model.User, remoteID model.ForgeRemoteID, isUser bool, account *model.Org, err error) {
	t.Helper()
	f := forge_mocks.NewMockForge(t)
	f.On("OrgByRemoteID", mock.Anything, admin, remoteID, isUser).Return(account, err)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("ForgeByID", int64(1)).Return(f, nil)
	server.Config.Services.Manager = mgr
}

func TestRemapOrg(t *testing.T) {
	s := newTestStore(t)
	admin := &model.User{ID: 1, ForgeID: 1, Admin: true, AccessToken: "token"}

	remap := func(t *testing.T, org *model.Org) *testContext {
		t.Helper()
		tc := newTestContext(t, s)
		withUser(admin)(tc)
		tc.Ctx.Set("org", org)
		RemapOrg(tc.Ctx)
		return tc
	}

	t.Run("old forge account is gone", func(t *testing.T) {
		org := &model.Org{ForgeID: 1, ForgeRemoteID: "1", Name: "gone", ConflictForgeRemoteID: "2"}
		require.NoError(t, s.OrgCreate(org))
		remapForge(t, admin, "1", false, nil, forge_types.ErrOrgNotFound)

		tc := remap(t, org)
		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

		stored, err := s.OrgFindByRemoteID(1, "2")
		require.NoError(t, err)
		assert.Equal(t, org.ID, stored.ID)
		assert.Empty(t, stored.ConflictForgeRemoteID)
	})

	for _, tc := range []struct {
		name    string
		account *model.Org
		err     error
		code    int
	}{
		{name: "old forge account still exists", account: &model.Org{Name: "exists"}, code: http.StatusConflict},
		{name: "forge can not look up old account", err: forge_types.ErrNotImplemented, code: http.StatusUnprocessableEntity},
		{name: "forge error", err: errors.New("forge unavailable"), code: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			org := &model.Org{ForgeID: 1, ForgeRemoteID: model.ForgeRemoteID("old-" + tc.name), Name: tc.name, ConflictForgeRemoteID: "new"}
			require.NoError(t, s.OrgCreate(org))
			remapForge(t, admin, org.ForgeRemoteID, false, tc.account, tc.err)

			assert.Equal(t, tc.code, remap(t, org).Recorder.Code)

			stored, err := s.OrgGet(org.ID)
			require.NoError(t, err)
			assert.Equal(t, org.ForgeRemoteID, stored.ForgeRemoteID)
		})
	}

	t.Run("no conflict", func(t *testing.T) {
		org := &model.Org{ForgeID: 1, ForgeRemoteID: "5", Name: "fine"}
		require.NoError(t, s.OrgCreate(org))

		assert.Equal(t, http.StatusBadRequest, remap(t, org).Recorder.Code)
	})

	t.Run("admin of another forge", func(t *testing.T) {
		org := &model.Org{ForgeID: 2, ForgeRemoteID: "6", Name: "other-forge", ConflictForgeRemoteID: "7"}
		require.NoError(t, s.OrgCreate(org))

		assert.Equal(t, http.StatusBadRequest, remap(t, org).Recorder.Code)
	})
}

func TestRemapUser(t *testing.T) {
	s := newTestStore(t)
	admin := &model.User{ID: 100, ForgeID: 1, Admin: true, AccessToken: "token"}

	org := &model.Org{ForgeID: 1, ForgeRemoteID: "1", Name: "jane", IsUser: true}
	require.NoError(t, s.OrgCreate(org))
	user := &model.User{ForgeID: 1, ForgeRemoteID: "1", Login: "jane", Hash: "1", OrgID: org.ID, ConflictForgeRemoteID: "2"}
	require.NoError(t, s.CreateUser(user))
	remapForge(t, admin, "1", true, nil, forge_types.ErrOrgNotFound)

	tc := newTestContext(t, s)
	withUser(admin)(tc)
	withParam("login", "jane")(tc)
	RemapUser(tc.Ctx)
	require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

	stored, err := s.GetUserByRemoteID(1, "2")
	require.NoError(t, err)
	assert.Equal(t, user.ID, stored.ID)
	assert.Empty(t, stored.ConflictForgeRemoteID)

	// the org of the user moves along
	storedOrg, err := s.OrgFindByRemoteID(1, "2")
	require.NoError(t, err)
	assert.Equal(t, org.ID, storedOrg.ID)
}
