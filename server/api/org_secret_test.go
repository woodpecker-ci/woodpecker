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

const orgSecretID = int64(7)

// newOrgSecretCtx builds a gin test context for an org secret endpoint with the
// org placed in the session. No service is wired, for handlers that bail before
// reaching the secret service (bind / validation errors).
func newOrgSecretCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("org", &model.Org{ID: orgSecretID, Name: "acme"})
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

// newOrgSecretCtxWithService also wires a mock secret service (returned by a
// mock manager) into the global config. The store is unit-tested in its own
// package, so it is mocked here.
func newOrgSecretCtxWithService(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *secret_service_mocks.MockService) {
	t.Helper()
	c, rec := newOrgSecretCtx(t, method, body)

	svc := secret_service_mocks.NewMockService(t)
	mgr := manager_mocks.NewMockManager(t)
	mgr.On("SecretService").Return(svc)
	server.Config.Services.Manager = mgr
	return c, rec, svc
}

// storedOrgSecret is a fully populated org secret as storage would return it.
func storedOrgSecret() *model.Secret {
	return &model.Secret{
		ID:     1,
		OrgID:  orgSecretID,
		Name:   "api_token",
		Value:  "super-secret-value",
		Events: []model.WebhookEvent{model.EventPush},
	}
}

func TestGetOrgSecret(t *testing.T) {
	t.Run("returns secret without leaking the value", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("OrgSecretFind", orgSecretID, "api_token").Return(storedOrgSecret(), nil)

		GetOrgSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, "api_token", got.Name)
		assert.Empty(t, got.Value)
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodGet, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("OrgSecretFind", orgSecretID, "nope").Return(nil, types.ErrRecordNotExist)

		GetOrgSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPostOrgSecret(t *testing.T) {
	t.Run("creates secret and never echoes the value back", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "super-secret-value", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodPost, in)
		svc.On("OrgSecretCreate", orgSecretID, mock.MatchedBy(func(s *model.Secret) bool {
			return s.Name == "api_token" && s.Value == "super-secret-value"
		})).Return(nil)

		PostOrgSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Empty(t, got.Value)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newOrgSecretCtx(t, http.MethodPost, nil)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		PostOrgSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation fails on empty value", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Events: []model.WebhookEvent{model.EventPush}}
		c, rec := newOrgSecretCtx(t, http.MethodPost, in)

		PostOrgSecret(c)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		in := &model.Secret{Name: "api_token", Value: "v", Events: []model.WebhookEvent{model.EventPush}}
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodPost, in)
		svc.On("OrgSecretCreate", orgSecretID, mock.Anything).Return(assert.AnError)

		PostOrgSecret(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPatchOrgSecret(t *testing.T) {
	t.Run("updates value but does not leak it", func(t *testing.T) {
		newValue := "rotated-value"
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{Value: &newValue})
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("OrgSecretFind", orgSecretID, "api_token").Return(storedOrgSecret(), nil)
		svc.On("OrgSecretUpdate", orgSecretID, mock.MatchedBy(func(s *model.Secret) bool {
			return s.Value == "rotated-value"
		})).Return(nil)

		PatchOrgSecret(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "rotated-value")
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodPatch, &model.SecretPatch{})
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("OrgSecretFind", orgSecretID, "nope").Return(nil, types.ErrRecordNotExist)

		PatchOrgSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		c, rec := newOrgSecretCtx(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("{nope")))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchOrgSecret(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetOrgSecretList(t *testing.T) {
	t.Run("lists secrets without leaking values", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("OrgSecretList", orgSecretID, mock.Anything).
			Return([]*model.Secret{storedOrgSecret()}, nil)

		GetOrgSecretList(c)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "super-secret-value")
		var got []*model.Secret
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Empty(t, got[0].Value)
	})

	t.Run("storage error returns internal error", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodGet, nil)
		svc.On("OrgSecretList", orgSecretID, mock.Anything).Return(nil, assert.AnError)

		GetOrgSecretList(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteOrgSecret(t *testing.T) {
	t.Run("happy path returns no content", func(t *testing.T) {
		c, _, svc := newOrgSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "api_token"}}
		svc.On("OrgSecretDelete", orgSecretID, "api_token").Return(nil)

		DeleteOrgSecret(c)

		assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	})

	t.Run("missing secret returns not found", func(t *testing.T) {
		c, rec, svc := newOrgSecretCtxWithService(t, http.MethodDelete, nil)
		c.Params = gin.Params{{Key: "secret", Value: "nope"}}
		svc.On("OrgSecretDelete", orgSecretID, "nope").Return(types.ErrRecordNotExist)

		DeleteOrgSecret(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
