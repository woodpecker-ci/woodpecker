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

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// newRegistryCtx builds a gin test context for a repo registry endpoint with
// the repo in the session. No service is wired, for handlers that bail before
// reaching the registry service.
func newRegistryCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("repo", secretTestRepo)
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

// newRegistryCtxWithService also wires a mock registry service. The store is
// unit-tested in its own package, so it is mocked here.
func newRegistryCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *registry_service_mocks.MockService) {
	t.Helper()
	c, rec := newRegistryCtx(t, method, body)

	svc := registry_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("RegistryServiceFromRepo", mock.Anything).Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

// storedRegistry is a fully populated registry as storage would return it,
// including the password that must never be leaked.
func storedRegistry() *model.Registry {
	return &model.Registry{
		ID:       1,
		RepoID:   1,
		Address:  "docker.io",
		Username: "user",
		Password: "super-secret-password",
	}
}

func TestGetRegistry(t *testing.T) {
	t.Run("returns registry without leaking the password", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("RegistryFind", secretTestRepo, "docker.io").Return(storedRegistry(), nil)

		GetRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "docker.io", got.Address)
		assert.Equal(t, "user", got.Username)
		assert.Empty(t, got.Password)
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("RegistryFind", secretTestRepo, "nope").Return(nil, types.ErrRecordNotExist)

		GetRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostRegistry(t *testing.T) {
	t.Run("creates registry and never echoes the password back", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "super-secret-password"}
		c, rec, svc := newRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("RegistryCreate", secretTestRepo, mock.MatchedBy(func(r *model.Registry) bool {
			return r.Address == "docker.io" && r.Password == "super-secret-password"
		})).Return(nil)

		PostRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Empty(t, got.Password)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newRegistryCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty password", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user"}
		c, rec := newRegistryCtx(t, http.MethodPost, in)

		PostRegistry(c)

		// the repo Post handler returns 400 on validation failure
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "p"}
		c, rec, svc := newRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("RegistryCreate", secretTestRepo, mock.Anything).Return(assert.AnError)

		PostRegistry(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchRegistry(t *testing.T) {
	t.Run("updates password but does not leak it", func(t *testing.T) {
		in := &model.Registry{Username: "user", Password: "rotated-password"}
		c, rec, svc := newRegistryCtxWithService(t, http.MethodPatch, in)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("RegistryFind", secretTestRepo, "docker.io").Return(storedRegistry(), nil)
		svc.On("RegistryUpdate", secretTestRepo, mock.MatchedBy(func(r *model.Registry) bool {
			return r.Password == "rotated-password"
		})).Return(nil)

		PatchRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-password")
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodPatch, &model.Registry{})
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("RegistryFind", secretTestRepo, "nope").Return(nil, types.ErrRecordNotExist)

		PatchRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newRegistryCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetRegistryList(t *testing.T) {
	t.Run("lists registries without leaking passwords", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("RegistryList", secretTestRepo, mock.Anything).
			Return([]*model.Registry{storedRegistry()}, nil)

		GetRegistryList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got []*model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Password)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("RegistryList", secretTestRepo, mock.Anything).Return(nil, assert.AnError)

		GetRegistryList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteRegistry(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("RegistryDelete", secretTestRepo, "docker.io").Return(nil)

		DeleteRegistry(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("RegistryDelete", secretTestRepo, "nope").Return(types.ErrRecordNotExist)

		DeleteRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
