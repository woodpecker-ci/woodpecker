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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// newForgeCtx builds a gin test context for a forge endpoint, with a mocked
// store. The store itself is unit-tested in its own package.
func newForgeCtx(t *testing.T, method string, body any) (*gin.Context, *httptest.ResponseRecorder, *store_mocks.MockStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/", jsonBody(t, body))
	c.Request.Header.Set("Content-Type", "application/json")

	_store := store_mocks.NewMockStore(t)
	c.Set("store", _store)
	return c, rec, _store
}

func TestPostForge(t *testing.T) {
	t.Run("should store the allowed orgs of the new forge", func(t *testing.T) {
		c, rec, _store := newForgeCtx(t, http.MethodPost, &model.ForgeWithOAuthClientSecret{
			Forge: model.Forge{
				Type: model.ForgeTypeGithub,
				URL:  "https://github.com",
				Orgs: []string{"org1", "org2"},
			},
			OAuthClientSecret: "client-secret",
		})

		var created *model.Forge
		_store.On("ForgeCreate", mock.Anything).Run(func(args mock.Arguments) {
			created, _ = args.Get(0).(*model.Forge)
		}).Return(nil)

		PostForge(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, created)
		assert.Equal(t, []string{"org1", "org2"}, created.Orgs)
	})
}

func TestPatchForge(t *testing.T) {
	t.Run("should update the allowed orgs of the forge", func(t *testing.T) {
		c, rec, _store := newForgeCtx(t, http.MethodPatch, &model.ForgeWithOAuthClientSecret{
			Forge: model.Forge{
				Type: model.ForgeTypeGithub,
				URL:  "https://github.com",
				Orgs: []string{"org2"},
			},
		})
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}

		_store.On("ForgeGet", int64(1)).Return(&model.Forge{
			ID:                1,
			Type:              model.ForgeTypeGithub,
			URL:               "https://github.com",
			OAuthClientSecret: "client-secret",
			Orgs:              []string{"org1"},
		}, nil)

		var updated *model.Forge
		_store.On("ForgeUpdate", mock.Anything).Run(func(args mock.Arguments) {
			updated, _ = args.Get(0).(*model.Forge)
		}).Return(nil)

		PatchForge(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, updated)
		assert.Equal(t, []string{"org2"}, updated.Orgs)
		// an empty client secret must not overwrite the stored one
		assert.Equal(t, "client-secret", updated.OAuthClientSecret)
	})

	t.Run("should clear the allowed orgs of the forge", func(t *testing.T) {
		c, rec, _store := newForgeCtx(t, http.MethodPatch, &model.ForgeWithOAuthClientSecret{
			Forge: model.Forge{
				Type: model.ForgeTypeGithub,
				URL:  "https://github.com",
			},
		})
		c.Params = gin.Params{{Key: "forge_id", Value: "1"}}

		_store.On("ForgeGet", int64(1)).Return(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			Orgs: []string{"org1"},
		}, nil)

		var updated *model.Forge
		_store.On("ForgeUpdate", mock.Anything).Run(func(args mock.Arguments) {
			updated, _ = args.Get(0).(*model.Forge)
		}).Return(nil)

		PatchForge(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, updated)
		assert.Empty(t, updated.Orgs)
	})
}
