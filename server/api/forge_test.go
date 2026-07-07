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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func testAppPrivateKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

func TestGetForge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should redact the github app private key for admins", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": "super-secret",
			},
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("user", &model.User{Admin: true})
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}

		GetForge(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusOK, w.Code)
		var response model.Forge
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "12345", response.AdditionalOptions["app-id"])
		assert.NotContains(t, response.AdditionalOptions, "app-private-key")
		assert.Equal(t, true, response.AdditionalOptions["app-private-key-set"])
	})

	t.Run("should return only public data for non-admins", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(&model.Forge{
			ID:            1,
			Type:          model.ForgeTypeGithub,
			URL:           "https://github.com",
			OAuthClientID: "oauth-client-id",
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": "super-secret",
			},
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("user", &model.User{Admin: false})
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}

		GetForge(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, float64(1), response["id"])
		assert.Equal(t, "github", response["type"])
		assert.Equal(t, "https://github.com", response["url"])
		assert.NotContains(t, response, "client")
		assert.NotContains(t, response, "additional_options")
	})

	t.Run("should reject an invalid forge id", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("user", &model.User{Admin: true})
		c.Params = gin.Params{{Key: "forge_id", Value: "invalid"}}

		GetForge(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetForges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	listForgesRequest := func(t *testing.T, user *model.User) *httptest.ResponseRecorder {
		t.Helper()
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeList", &model.ListOptions{Page: 1, PerPage: 50}).Return([]*model.Forge{{
			ID:            1,
			Type:          model.ForgeTypeGithub,
			URL:           "https://github.com",
			OAuthClientID: "oauth-client-id",
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": "super-secret",
			},
		}}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Set("user", user)
		c.Request = httptest.NewRequest(http.MethodGet, "/forges", nil)

		GetForges(c)
		c.Writer.WriteHeaderNow()
		return w
	}

	t.Run("should redact the github app private key for admins", func(t *testing.T) {
		w := listForgesRequest(t, &model.User{Admin: true})

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*model.Forge
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Len(t, response, 1)
		assert.Equal(t, "oauth-client-id", response[0].OAuthClientID)
		assert.Equal(t, "12345", response[0].AdditionalOptions["app-id"])
		assert.NotContains(t, response[0].AdditionalOptions, "app-private-key")
		assert.Equal(t, true, response[0].AdditionalOptions["app-private-key-set"])
	})

	t.Run("should return only public data for non-admins", func(t *testing.T) {
		w := listForgesRequest(t, &model.User{Admin: false})

		assert.Equal(t, http.StatusOK, w.Code)
		var response []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Len(t, response, 1)
		assert.Equal(t, float64(1), response[0]["id"])
		assert.Equal(t, "github", response[0]["type"])
		assert.Equal(t, "https://github.com", response[0]["url"])
		assert.NotContains(t, response[0], "client")
		assert.NotContains(t, response[0], "additional_options")
	})
}

func patchForgeRequest(t *testing.T, mockStore *store_mocks.MockStore, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("store", mockStore)
	c.Params = gin.Params{{Key: "forge_id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/forges/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	PatchForge(c)
	c.Writer.WriteHeaderNow()
	return w
}

func TestPatchForge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appPrivateKey := testAppPrivateKey(t)

	storedForge := func() *model.Forge {
		return &model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": appPrivateKey,
			},
		}
	}

	t.Run("should keep the stored github app private key when it is omitted", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasMarker := forge.AdditionalOptions["app-private-key-set"]
			return forge.AdditionalOptions["app-private-key"] == appPrivateKey && !hasMarker
		})).Return(nil)

		// clients echo back the redaction marker, it must not be persisted
		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": "12345", "app-private-key-set": true}}`)

		assert.Equal(t, http.StatusOK, w.Code)
		var response model.Forge
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.NotContains(t, response.AdditionalOptions, "app-private-key")
		assert.Equal(t, true, response.AdditionalOptions["app-private-key-set"])
	})

	t.Run("should drop the stored github app private key when the app id is cleared", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasKey := forge.AdditionalOptions["app-private-key"]
			return !hasKey
		})).Return(nil)

		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": ""}}`)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should drop the stored github app private key when the app id is removed", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasKey := forge.AdditionalOptions["app-private-key"]
			return !hasKey
		})).Return(nil)

		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {}}`)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject an invalid github app configuration", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
		}, nil)

		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": "12345"}}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})

	t.Run("should replace the stored github app private key when a new one is sent", func(t *testing.T) {
		newAppPrivateKey := testAppPrivateKey(t)
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			return forge.AdditionalOptions["app-private-key"] == newAppPrivateKey
		})).Return(nil)

		body, err := json.Marshal(map[string]any{
			"type": "github",
			"url":  "https://github.com",
			"additional_options": map[string]any{
				"app-id":          "12345",
				"app-private-key": newAppPrivateKey,
			},
		})
		require.NoError(t, err)

		w := patchForgeRequest(t, mockStore, string(body))

		assert.Equal(t, http.StatusOK, w.Code)
		var response model.Forge
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.NotContains(t, response.AdditionalOptions, "app-private-key")
		assert.Equal(t, true, response.AdditionalOptions["app-private-key-set"])
	})

	t.Run("should reject a numeric app id", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)

		// additional options are typed, GitHub app ids must be sent as strings
		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": 12345}}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "must be a string")
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})

	t.Run("should clear the app configuration when additional options are omitted", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasKey := forge.AdditionalOptions[model.ForgeGithubOptionAppPrivateKey]
			return !hasKey
		})).Return(nil)

		// omitting additional_options replaces them wholesale and must not panic
		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com"}`)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject a new private key without an app id", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)

		// an explicitly submitted key must not be dropped silently
		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": "", "app-private-key": "some-new-key"}}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})

	t.Run("should validate non-github forges as well", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)

		// bitbucket-dc requires a git machine account
		w := patchForgeRequest(t, mockStore, `{"type": "bitbucket-dc", "url": "https://bb.example.com", "additional_options": {}}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})

	t.Run("should not restore github options when the forge type changes", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasKey := forge.AdditionalOptions["app-private-key"]
			return forge.Type == model.ForgeTypeGitlab && !hasKey
		})).Return(nil)

		// the echoed app-id must not resurrect the github key on a gitlab forge
		w := patchForgeRequest(t, mockStore, `{"type": "gitlab", "url": "https://gitlab.com", "additional_options": {"app-id": "12345"}}`)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return a conflict when the forge cannot be updated", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(storedForge(), nil)
		mockStore.On("ForgeUpdate", mock.Anything).Return(errors.New("update failed"))

		w := patchForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": "12345"}}`)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("should reject a malformed request body", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := patchForgeRequest(t, mockStore, `{"type": "github"`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})

	t.Run("should reject an invalid forge id", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "forge_id", Value: "invalid"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/forges/invalid", strings.NewReader(`{"type": "github", "url": "https://github.com"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		PatchForge(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
	})
}

func TestPostForge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	postForgeRequest := func(t *testing.T, mockStore *store_mocks.MockStore, body string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Request = httptest.NewRequest(http.MethodPost, "/forges", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		PostForge(c)
		c.Writer.WriteHeaderNow()
		return w
	}

	t.Run("should create a github forge and redact the app private key", func(t *testing.T) {
		appPrivateKey := testAppPrivateKey(t)
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeCreate", mock.MatchedBy(func(forge *model.Forge) bool {
			_, hasMarker := forge.AdditionalOptions["app-private-key-set"]
			return forge.Type == model.ForgeTypeGithub &&
				forge.AdditionalOptions["app-private-key"] == appPrivateKey &&
				!hasMarker
		})).Return(nil)

		// clients echo back the redaction marker, it must not be persisted
		body, err := json.Marshal(map[string]any{
			"type": "github",
			"url":  "https://github.com",
			"additional_options": map[string]any{
				"app-id":              "12345",
				"app-private-key":     appPrivateKey,
				"app-private-key-set": true,
			},
		})
		require.NoError(t, err)

		w := postForgeRequest(t, mockStore, string(body))

		assert.Equal(t, http.StatusOK, w.Code)
		var response model.Forge
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.NotContains(t, response.AdditionalOptions, "app-private-key")
		assert.Equal(t, true, response.AdditionalOptions["app-private-key-set"])
	})

	t.Run("should reject a github app id without a private key", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := postForgeRequest(t, mockStore, `{"type": "github", "url": "https://github.com", "additional_options": {"app-id": "12345"}}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeCreate", mock.Anything)
	})

	t.Run("should not validate app credentials for non-github forges", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeCreate", mock.MatchedBy(func(forge *model.Forge) bool {
			return forge.Type == model.ForgeTypeGitlab
		})).Return(nil)

		w := postForgeRequest(t, mockStore, `{"type": "gitlab", "url": "https://gitlab.com"}`)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return an error when the forge cannot be created", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeCreate", mock.Anything).Return(errors.New("create failed"))

		w := postForgeRequest(t, mockStore, `{"type": "gitlab", "url": "https://gitlab.com"}`)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should reject a malformed request body", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := postForgeRequest(t, mockStore, `{"type": "github"`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStore.AssertNotCalled(t, "ForgeCreate", mock.Anything)
	})
}

func TestGetForgeAppHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fixtureServer := httptest.NewServer(fixtures.Handler())
	defer fixtureServer.Close()

	appHealthRequest := func(t *testing.T, forge *model.Forge) *httptest.ResponseRecorder {
		t.Helper()
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(forge, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/forges/1/app-health", nil)

		GetForgeAppHealth(c)
		c.Writer.WriteHeaderNow()
		return w
	}

	t.Run("should report a working github app", func(t *testing.T) {
		w := appHealthRequest(t, &model.Forge{
			ID:         1,
			Type:       model.ForgeTypeGithub,
			URL:        fixtureServer.URL,
			SkipVerify: true,
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": testAppPrivateKey(t),
			},
		})

		assert.Equal(t, http.StatusOK, w.Code)
		var response ForgeAppHealth
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.True(t, response.Healthy)
		assert.Equal(t, "Woodpecker Test App", response.AppName)
		assert.Equal(t, 1, response.Installations)
	})

	t.Run("should report a missing github app configuration", func(t *testing.T) {
		w := appHealthRequest(t, &model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  fixtureServer.URL,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		var response ForgeAppHealth
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.False(t, response.Healthy)
		assert.NotEmpty(t, response.Error)
	})

	t.Run("should report an unusable github app configuration", func(t *testing.T) {
		w := appHealthRequest(t, &model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  fixtureServer.URL,
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": "not-a-private-key",
			},
		})

		assert.Equal(t, http.StatusOK, w.Code)
		var response ForgeAppHealth
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.False(t, response.Healthy)
		assert.NotEmpty(t, response.Error)
	})

	t.Run("should reject forge types without app support", func(t *testing.T) {
		w := appHealthRequest(t, &model.Forge{
			ID:   1,
			Type: model.ForgeTypeGitlab,
			URL:  "https://gitlab.com",
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "does not support app checks")
	})

	t.Run("should return not found for an unknown forge", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}

		GetForgeAppHealth(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should reject an invalid forge id", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", mockStore)
		c.Params = gin.Params{{Key: "forge_id", Value: "invalid"}}

		GetForgeAppHealth(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
