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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
)

func TestPostUser(t *testing.T) {
	s := newTestStore(t)

	t.Run("missing forge_id falls back to the primary forge", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRequest(http.MethodPost, &model.User{Login: "carol"})(tc)

		PostUser(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

		created := new(model.User)
		tc.decodeJSON(t, created)
		assert.EqualValues(t, model.PrimaryForgeID, created.ForgeID, "user must never be created with forge id 0")

		// the user's org must be forge-scoped as well
		org, err := s.OrgGet(created.OrgID)
		require.NoError(t, err)
		assert.EqualValues(t, model.PrimaryForgeID, org.ForgeID, "org must never be created with forge id 0")
	})

	t.Run("explicit forge_id is kept", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRequest(http.MethodPost, &model.User{Login: "dave", ForgeID: 2})(tc)

		PostUser(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

		created := new(model.User)
		tc.decodeJSON(t, created)
		assert.EqualValues(t, 2, created.ForgeID)
	})
}

func TestPatchUser(t *testing.T) {
	s := newTestStore(t)

	admins := server.Config.Permissions.Admins
	t.Cleanup(func() { server.Config.Permissions.Admins = admins })

	patch := func(t *testing.T, login string, in *model.User) {
		t.Helper()
		tc := newTestContext(t, s)
		withRequest(http.MethodPatch, in)(tc)
		withParam("login", login)(tc)

		PatchUser(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())
	}

	t.Run("revoking a listed admin stops admin at sign in but keeps them listed", func(t *testing.T) {
		server.Config.Permissions.Admins = permissions.NewAdmins([]string{"Erin"}, false)
		require.NoError(t, s.CreateUser(&model.User{Login: "erin", ForgeID: model.PrimaryForgeID, ForgeRemoteID: "5", Hash: "erin", Admin: true}))

		patch(t, "erin", &model.User{Login: "erin", ForgeID: model.PrimaryForgeID, ForgeRemoteID: "5", Admin: false})

		// the next sign in must not grant admin again
		assert.False(t, server.Config.Permissions.Admins.IsAdmin(&model.User{Login: "erin"}))
		stored, err := s.GetUserByLogin(model.PrimaryForgeID, "erin")
		require.NoError(t, err)
		assert.False(t, stored.Admin)

		// the user list must still show that the user is in WOODPECKER_ADMIN
		tc := newTestContext(t, s)
		GetUsers(tc.Ctx)
		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())
		var users []*model.User
		tc.decodeJSON(t, &users)
		require.NotEmpty(t, users)
		for _, u := range users {
			if u.Login == "erin" {
				assert.True(t, u.AdminEnv)
				assert.False(t, u.Admin)
				assert.True(t, u.AdminRevoked)
			}
		}
	})

	t.Run("a listed user who never signed in is not shown as revoked", func(t *testing.T) {
		server.Config.Permissions.Admins = permissions.NewAdmins([]string{"grace"}, false)
		require.NoError(t, s.CreateUser(&model.User{Login: "grace", ForgeID: model.PrimaryForgeID, ForgeRemoteID: "7", Hash: "grace"}))

		tc := newTestContext(t, s)
		GetUsers(tc.Ctx)
		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())
		var users []*model.User
		tc.decodeJSON(t, &users)
		found := false
		for _, u := range users {
			if u.Login == "grace" {
				found = true
				assert.True(t, u.AdminEnv)
				assert.False(t, u.Admin)
				assert.False(t, u.AdminRevoked, "grace gets admin at the first sign in")
			}
		}
		assert.True(t, found)
	})

	t.Run("keeping admin leaves the admin list alone", func(t *testing.T) {
		server.Config.Permissions.Admins = permissions.NewAdmins([]string{"frank"}, false)
		require.NoError(t, s.CreateUser(&model.User{Login: "frank", ForgeID: model.PrimaryForgeID, ForgeRemoteID: "6", Hash: "frank", Admin: true}))

		patch(t, "frank", &model.User{Login: "frank", ForgeID: model.PrimaryForgeID, ForgeRemoteID: "6", Admin: true})

		assert.True(t, server.Config.Permissions.Admins.IsAdmin(&model.User{Login: "frank"}))
	})
}
