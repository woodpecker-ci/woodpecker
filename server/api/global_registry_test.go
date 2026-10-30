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

func newGlobalRegistryCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

func newGlobalRegistryCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *registry_service_mocks.MockService) {
	t.Helper()
	c, rec := newGlobalRegistryCtx(t, method, body)

	svc := registry_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("RegistryService").Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

func storedGlobalRegistry() *model.Registry {
	return &model.Registry{
		ID:       1,
		Address:  "docker.io",
		Username: "user",
		Password: "super-secret-password",
	}
}

func TestGetGlobalRegistry(t *testing.T) {
	t.Run("returns registry without leaking the password", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("GlobalRegistryFind", "docker.io").Return(storedGlobalRegistry(), nil)

		GetGlobalRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "docker.io", got.Address)
		assert.Empty(t, got.Password)
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("GlobalRegistryFind", "nope").Return(nil, types.ErrRecordNotExist)

		GetGlobalRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostGlobalRegistry(t *testing.T) {
	t.Run("creates registry and never echoes the password back", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "super-secret-password"}
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("GlobalRegistryCreate", mock.MatchedBy(func(r *model.Registry) bool {
			return r.Address == "docker.io" && r.Password == "super-secret-password"
		})).Return(nil)

		PostGlobalRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Empty(t, got.Password)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newGlobalRegistryCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostGlobalRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty password", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user"}
		c, rec := newGlobalRegistryCtx(t, http.MethodPost, in)

		PostGlobalRegistry(c)

		// the global handler returns 400 on validation failure
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "p"}
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("GlobalRegistryCreate", mock.Anything).Return(assert.AnError)

		PostGlobalRegistry(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchGlobalRegistry(t *testing.T) {
	t.Run("updates password but does not leak it", func(t *testing.T) {
		in := &model.Registry{Username: "user", Password: "rotated-password"}
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodPatch, in)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("GlobalRegistryFind", "docker.io").Return(storedGlobalRegistry(), nil)
		svc.On("GlobalRegistryUpdate", mock.MatchedBy(func(r *model.Registry) bool {
			return r.Password == "rotated-password"
		})).Return(nil)

		PatchGlobalRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-password")
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodPatch, &model.Registry{})
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("GlobalRegistryFind", "nope").Return(nil, types.ErrRecordNotExist)

		PatchGlobalRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newGlobalRegistryCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchGlobalRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetGlobalRegistryList(t *testing.T) {
	t.Run("lists registries without leaking passwords", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("GlobalRegistryList", mock.Anything).
			Return([]*model.Registry{storedGlobalRegistry()}, nil)

		GetGlobalRegistryList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got []*model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Password)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("GlobalRegistryList", mock.Anything).Return(nil, assert.AnError)

		GetGlobalRegistryList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteGlobalRegistry(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newGlobalRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("GlobalRegistryDelete", "docker.io").Return(nil)

		DeleteGlobalRegistry(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("GlobalRegistryDelete", "nope").Return(types.ErrRecordNotExist)

		DeleteGlobalRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
