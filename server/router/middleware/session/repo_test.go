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

package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

func TestSetPerm(t *testing.T) {
	s := datastore.NewTestStore(t)

	newCtx := func(user *model.User, repo *model.Repo) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("store", s)
		if user != nil {
			c.Set("user", user)
		}
		c.Set("repo", repo)
		return c, rec
	}

	t.Run("stale perm of user on another forge is not synced via the repo's forge", func(t *testing.T) {
		mgr := manager_mocks.NewMockManager(t)
		// forge of the repo; no Repo() expectation: the mock fails the test
		// if the middleware sends the foreign user's token to this forge
		_forge := forge_mocks.NewMockForge(t)
		mgr.On("ForgeFromRepo", mock.Anything).Return(_forge, nil)
		server.Config.Services.Manager = mgr

		user := &model.User{ID: 10, Login: "alice", ForgeID: 1}
		repo := &model.Repo{ID: 20, FullName: "acme/spanner", ForgeID: 2, ForgeRemoteID: "7"}
		c, _ := newCtx(user, repo)

		SetPerm()(c)

		perm := Perm(c)
		require.NotNil(t, perm)
		assert.False(t, perm.Pull)
		assert.False(t, perm.Push)
		assert.False(t, perm.Admin)
	})

	t.Run("stale perm of user on the repo's forge is synced", func(t *testing.T) {
		mgr := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		mgr.On("ForgeFromRepo", mock.Anything).Return(_forge, nil)
		server.Config.Services.Manager = mgr

		user := &model.User{ID: 11, Login: "bob", ForgeID: 2}
		repo := &model.Repo{ID: 21, FullName: "acme/spanner", ForgeID: 2, ForgeRemoteID: "7"}

		_forge.On("Repo", mock.Anything, user, repo.ForgeRemoteID, repo.Owner, repo.Name).
			Return(&model.Repo{
				ForgeRemoteID: repo.ForgeRemoteID,
				ForgeID:       repo.ForgeID,
				Perm:          &model.Perm{Pull: true, Push: true},
			}, nil).Once()

		c, _ := newCtx(user, repo)

		SetPerm()(c)

		perm := Perm(c)
		require.NotNil(t, perm)
		assert.True(t, perm.Pull)
		assert.True(t, perm.Push)
		assert.EqualValues(t, repo.ID, perm.RepoID)
		assert.EqualValues(t, user.ID, perm.UserID)
		assert.InDelta(t, time.Now().Unix(), perm.Synced, 5)
	})
}
