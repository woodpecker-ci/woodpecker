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

package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
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
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestRepoTokenRouteRequiresPush(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := datastore.NewTestStore(t)
	repo := &model.Repo{
		ForgeID:       1,
		ForgeRemoteID: "repo-token-route",
		Owner:         "owner",
		Name:          "repo-token-route",
		FullName:      "owner/repo-token-route",
		Visibility:    model.VisibilityPrivate,
	}
	require.NoError(t, s.CreateRepo(repo))
	user := &model.User{ID: 1, Login: "user", ForgeID: repo.ForgeID}

	manager := manager_mocks.NewMockManager(t)
	manager.On("ForgeFromRepo", mock.Anything).Return(forge_mocks.NewMockForge(t), nil)
	previousManager := server.Config.Services.Manager
	server.Config.Services.Manager = manager
	t.Cleanup(func() { server.Config.Services.Manager = previousManager })

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("store", s)
		c.Set("user", user)
		c.Next()
	})
	apiRoutes(engine.Group(""))

	requestToken := func() *httptest.ResponseRecorder {
		t.Helper()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/repos/"+strconv.FormatInt(repo.ID, 10)+"/token", nil)
		engine.ServeHTTP(recorder, request)
		return recorder
	}

	perm := &model.Perm{UserID: user.ID, RepoID: repo.ID, Pull: true, Synced: time.Now().Unix()}
	require.NoError(t, s.PermUpsert(perm))

	denied := requestToken()
	assert.Equal(t, http.StatusNotFound, denied.Code)
	_, err := s.TokenFind(repo, model.TokenTypeBadge)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)

	perm.Push = true
	require.NoError(t, s.PermUpsert(perm))

	allowed := requestToken()
	assert.Equal(t, http.StatusOK, allowed.Code)
	_, err = s.TokenFind(repo, model.TokenTypeBadge)
	assert.NoError(t, err)
}

func TestRepoTokenRotateRouteRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := datastore.NewTestStore(t)
	repo := &model.Repo{
		ForgeID:       1,
		ForgeRemoteID: "repo-token-rotate-route",
		Owner:         "owner",
		Name:          "repo-token-rotate-route",
		FullName:      "owner/repo-token-rotate-route",
		Visibility:    model.VisibilityPrivate,
	}
	require.NoError(t, s.CreateRepo(repo))
	old := model.NewToken(repo.ID, model.TokenTypeBadge)
	require.NoError(t, s.TokenCreate(old))
	user := &model.User{ID: 1, Login: "user", ForgeID: repo.ForgeID}

	manager := manager_mocks.NewMockManager(t)
	manager.On("ForgeFromRepo", mock.Anything).Return(forge_mocks.NewMockForge(t), nil)
	previousManager := server.Config.Services.Manager
	server.Config.Services.Manager = manager
	t.Cleanup(func() { server.Config.Services.Manager = previousManager })

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("store", s)
		c.Set("user", user)
		c.Next()
	})
	apiRoutes(engine.Group(""))

	rotateToken := func() *httptest.ResponseRecorder {
		t.Helper()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/repos/"+strconv.FormatInt(repo.ID, 10)+"/token/rotate", nil)
		engine.ServeHTTP(recorder, request)
		return recorder
	}

	perm := &model.Perm{
		UserID: user.ID,
		RepoID: repo.ID,
		Pull:   true,
		Push:   true,
		Synced: time.Now().Unix(),
	}
	require.NoError(t, s.PermUpsert(perm))

	denied := rotateToken()
	assert.Equal(t, http.StatusForbidden, denied.Code)
	stored, err := s.TokenFind(repo, model.TokenTypeBadge)
	require.NoError(t, err)
	assert.Equal(t, old.Value, stored.Value)

	perm.Admin = true
	require.NoError(t, s.PermUpsert(perm))

	allowed := rotateToken()
	assert.Equal(t, http.StatusOK, allowed.Code)
	stored, err = s.TokenFind(repo, model.TokenTypeBadge)
	require.NoError(t, err)
	assert.NotEqual(t, old.Value, stored.Value)
}
