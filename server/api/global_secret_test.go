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
	"io"
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

// jsonBody marshals body to a reader, or an empty reader when body is nil.
// Shared by the org and global secret tests.
func jsonBody(t *testing.T, body any) io.Reader {
	t.Helper()
	if body == nil {
		return bytes.NewReader(nil)
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// newGlobalSecretCtx builds a gin test context for a global secret endpoint.
// No service is wired, for handlers that bail before reaching the service.
func newGlobalSecretCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

// newGlobalSecretCtxWithService also wires a mock secret service. The store is
// unit-tested in its own package, so it is mocked here.
func newGlobalSecretCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *secret_service_mocks.MockService) {
	t.Helper()
	c, rec := newGlobalSecretCtx(t, method, body)

	svc := secret_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("SecretService").Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

// storedGlobalSecret is a fully populated global secret as storage returns it.
func storedGlobalSecret() *model.Secret {
	return &model.Secret{
		ID:     1,
		Name:   "api_token",
		Value:  "super-secret-value",
		Events: []model.WebhookEvent{model.EventPush},
	}
}

func TestGetGlobalSecret(t *testing.T) {
	t.Run("returns secret without leaking the value", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("GlobalSecretFind", "api_token").Return(storedGlobalSecret(), nil)

		GetGlobalSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "api_token", got.Name)
		assert.Empty(t, got.Value)
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("GlobalSecretFind", "nope").Return(nil, types.ErrRecordNotExist)

		GetGlobalSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostGlobalSecret(t *testing.T) {
	t.Run("creates secret and never echoes the value back", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "super-secret-value", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodPost, in)
		svc.On("GlobalSecretCreate", mock.MatchedBy(func(s *model.Secret) bool {
			return s.Name == "api_token" && s.Value == "super-secret-value"
		})).Return(nil)

		PostGlobalSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Empty(t, got.Value)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newGlobalSecretCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostGlobalSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty value", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Events: []model.WebhookEvent{model.EventPush}}
		c, rec := newGlobalSecretCtx(t, http.MethodPost, in)

		PostGlobalSecret(c)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "v", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodPost, in)
		svc.On("GlobalSecretCreate", mock.Anything).Return(assert.AnError)

		PostGlobalSecret(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchGlobalSecret(t *testing.T) {
	t.Run("updates value but does not leak it", func(t *testing.T) {
		newValue := "rotated-value"
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{Value: &newValue})
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("GlobalSecretFind", "api_token").Return(storedGlobalSecret(), nil)
		svc.On("GlobalSecretUpdate", mock.MatchedBy(func(s *model.Secret) bool {
			return s.Value == "rotated-value"
		})).Return(nil)

		PatchGlobalSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-value")
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{})
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("GlobalSecretFind", "nope").Return(nil, types.ErrRecordNotExist)

		PatchGlobalSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newGlobalSecretCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchGlobalSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetGlobalSecretList(t *testing.T) {
	t.Run("lists secrets without leaking values", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("GlobalSecretList", mock.Anything).
			Return([]*model.Secret{storedGlobalSecret()}, nil)

		GetGlobalSecretList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got []*model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Value)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("GlobalSecretList", mock.Anything).Return(nil, assert.AnError)

		GetGlobalSecretList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteGlobalSecret(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newGlobalSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("GlobalSecretDelete", "api_token").Return(nil)

		DeleteGlobalSecret(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newGlobalSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("GlobalSecretDelete", "nope").Return(types.ErrRecordNotExist)

		DeleteGlobalSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
