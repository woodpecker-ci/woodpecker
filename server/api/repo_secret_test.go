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
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// secretTestRepo is the repo placed in the gin session for secret endpoints.
var secretTestRepo = &model.Repo{ID: 1, FullName: "owner/repo"}

// newSecretCtx builds a gin test context for a secret endpoint with a JSON
// body. No service is wired, so it suits handlers that bail before reaching
// the secret service (bind / validation errors).
func newSecretCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("repo", secretTestRepo)

	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	c.Request = httptest.NewRequest(method, "/", reader)
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

// newSecretCtxWithService is like newSecretCtx but also wires a mock secret
// service (returned by a mock manager) into the global config. The secret CRUD
// handlers only touch the secret service; the underlying store is unit-tested
// in its own package, so it is mocked here.
func newSecretCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *secret_service_mocks.MockService) {
	t.Helper()
	c, rec := newSecretCtx(t, method, body)

	svc := secret_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("SecretServiceFromRepo", mock.Anything).Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

// a fully populated secret as the service would return it from storage.
func storedSecret() *model.Secret {
	return &model.Secret{
		ID:     1,
		RepoID: 1,
		Name:   "api_token",
		Value:  "super-secret-value",
		Events: []model.WebhookEvent{model.EventPush},
		Images: []string{"alpine"},
	}
}

func TestGetSecret(t *testing.T) {
	t.Run("returns secret without leaking the value", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("SecretFind", secretTestRepo, "api_token").Return(storedSecret(), nil)

		GetSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")

		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "api_token", got.Name)
		assert.Empty(t, got.Value)
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("SecretFind", secretTestRepo, "nope").Return(nil, types.ErrRecordNotExist)

		GetSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostSecret(t *testing.T) {
	t.Run("creates secret and never echoes the value back", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "super-secret-value", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newSecretCtxWithService(t, http.MethodPost, in)
		// The handler must pass the submitted value through to storage.
		svc.On("SecretCreate", secretTestRepo, mock.MatchedBy(func(s *model.Secret) bool {
			return s.Name == "api_token" && s.Value == "super-secret-value"
		})).Return(nil)

		PostSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")

		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "api_token", got.Name)
		assert.Empty(t, got.Value)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newSecretCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty value", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Events: []model.WebhookEvent{model.EventPush}}
		c, rec := newSecretCtx(t, http.MethodPost, in)

		PostSecret(c)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "v", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newSecretCtxWithService(t, http.MethodPost, in)
		svc.On("SecretCreate", secretTestRepo, mock.Anything).Return(assert.AnError)

		PostSecret(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchSecret(t *testing.T) {
	t.Run("updates value but does not leak it", func(t *testing.T) {
		newValue := "rotated-value"
		c, rec, svc := newSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{Value: &newValue})
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("SecretFind", secretTestRepo, "api_token").Return(storedSecret(), nil)
		svc.On("SecretUpdate", secretTestRepo, mock.MatchedBy(func(s *model.Secret) bool {
			return s.Value == "rotated-value"
		})).Return(nil)

		PatchSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-value")
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{})
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("SecretFind", secretTestRepo, "nope").Return(nil, types.ErrRecordNotExist)

		PatchSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newSecretCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetSecretList(t *testing.T) {
	t.Run("lists secrets without leaking values", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("SecretList", secretTestRepo, mock.Anything).
			Return([]*model.Secret{storedSecret()}, nil)

		GetSecretList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")

		var got []*model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Value)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("SecretList", secretTestRepo, mock.Anything).Return(nil, assert.AnError)

		GetSecretList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteSecret(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("SecretDelete", secretTestRepo, "api_token").Return(nil)

		DeleteSecret(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("SecretDelete", secretTestRepo, "nope").Return(types.ErrRecordNotExist)

		DeleteSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
