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
}
