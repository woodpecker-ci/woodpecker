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

func newOrgRegistryCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("org", &model.Org{ID: orgSecretID, Name: "acme"})
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

func newOrgRegistryCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *registry_service_mocks.MockService) {
	t.Helper()
	c, rec := newOrgRegistryCtx(t, method, body)

	svc := registry_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("RegistryService").Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

func storedOrgRegistry() *model.Registry {
	return &model.Registry{
		ID:       1,
		OrgID:    orgSecretID,
		Address:  "docker.io",
		Username: "user",
		Password: "super-secret-password",
	}
}

func TestGetOrgRegistry(t *testing.T) {
	t.Run("returns registry without leaking the password", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("OrgRegistryFind", orgSecretID, "docker.io").Return(storedOrgRegistry(), nil)

		GetOrgRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "docker.io", got.Address)
		assert.Empty(t, got.Password)
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("OrgRegistryFind", orgSecretID, "nope").Return(nil, types.ErrRecordNotExist)

		GetOrgRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostOrgRegistry(t *testing.T) {
	t.Run("creates registry and never echoes the password back", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "super-secret-password"}
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("OrgRegistryCreate", orgSecretID, mock.MatchedBy(func(r *model.Registry) bool {
			return r.Address == "docker.io" && r.Password == "super-secret-password"
		})).Return(nil)

		PostOrgRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Empty(t, got.Password)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newOrgRegistryCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostOrgRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty password", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user"}
		c, rec := newOrgRegistryCtx(t, http.MethodPost, in)

		PostOrgRegistry(c)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Registry{Address: "docker.io", Username: "user", Password: "p"}
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodPost, in)
		svc.On("OrgRegistryCreate", orgSecretID, mock.Anything).Return(assert.AnError)

		PostOrgRegistry(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchOrgRegistry(t *testing.T) {
	t.Run("updates password but does not leak it", func(t *testing.T) {
		in := &model.Registry{Username: "user", Password: "rotated-password"}
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodPatch, in)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("OrgRegistryFind", orgSecretID, "docker.io").Return(storedOrgRegistry(), nil)
		svc.On("OrgRegistryUpdate", orgSecretID, mock.MatchedBy(func(r *model.Registry) bool {
			return r.Password == "rotated-password"
		})).Return(nil)

		PatchOrgRegistry(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-password")
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodPatch, &model.Registry{})
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("OrgRegistryFind", orgSecretID, "nope").Return(nil, types.ErrRecordNotExist)

		PatchOrgRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newOrgRegistryCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchOrgRegistry(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetOrgRegistryList(t *testing.T) {
	t.Run("lists registries without leaking passwords", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("OrgRegistryList", orgSecretID, mock.Anything).
			Return([]*model.Registry{storedOrgRegistry()}, nil)

		GetOrgRegistryList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-password")
		var got []*model.Registry
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Password)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodGet, nil)
		svc.On("OrgRegistryList", orgSecretID, mock.Anything).Return(nil, assert.AnError)

		GetOrgRegistryList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteOrgRegistry(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newOrgRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "docker.io"}}
		svc.On("OrgRegistryDelete", orgSecretID, "docker.io").Return(nil)

		DeleteOrgRegistry(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing registry returns not found", func(t *testing.T) {
		c, rec, svc := newOrgRegistryCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "registry", Value: "nope"}}
		svc.On("OrgRegistryDelete", orgSecretID, "nope").Return(types.ErrRecordNotExist)

		DeleteOrgRegistry(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
