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

package services

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jellydator/ttlcache/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	config_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/config/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// extensionServer answers the secret, registry and config extension calls
// below /secret, /registry and /config.
func extensionServer(t *testing.T) *httptest.Server {
	t.Helper()
	respond := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/secret", respond(`{"secrets":[{"name":"token","value":"from-extension"}]}`))
	mux.HandleFunc("/registry", respond(`{"registries":[{"address":"ghcr.io","username":"ext","password":"pw"}]}`))
	mux.HandleFunc("/config", respond(`{"configs":[{"name":"extension.yaml","data":"steps: []"}]}`))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func testClient(t *testing.T) *utils.Client {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	client, err := utils.NewHTTPClient(key, "loopback")
	require.NoError(t, err)
	return client
}

func TestSecretServiceFromRepo(t *testing.T) {
	server := extensionServer(t)
	base := secret_service_mocks.NewMockService(t)
	m := &manager{secret: base, client: testClient(t)}

	t.Run("without extension the global service is used", func(t *testing.T) {
		assert.Same(t, base, m.SecretServiceFromRepo(&model.Repo{}))
	})

	t.Run("repo extension adds its secrets", func(t *testing.T) {
		base.On("SecretListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return([]*model.Secret{{Name: "base", Value: "from-db"}}, nil).Once()

		// the trailing slash is dropped from the configured endpoint
		repo := &model.Repo{SecretExtensionEndpoint: server.URL + "/secret/"}
		secrets, err := m.SecretServiceFromRepo(repo).SecretListPipeline(t.Context(), repo, &model.Pipeline{}, nil)
		require.NoError(t, err)
		assert.Equal(t, []*model.Secret{
			{Name: "token", Value: "from-extension"},
			{Name: "base", Value: "from-db"},
		}, secrets)
	})
}

func TestRegistryServiceFromRepo(t *testing.T) {
	server := extensionServer(t)
	base := registry_service_mocks.NewMockService(t)
	m := &manager{registry: base, client: testClient(t)}

	t.Run("without extension the global service is used", func(t *testing.T) {
		assert.Same(t, base, m.RegistryServiceFromRepo(&model.Repo{}))
	})

	t.Run("repo extension adds its registries", func(t *testing.T) {
		base.On("RegistryListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return([]*model.Registry{{Address: "docker.io", Username: "db"}}, nil).Once()

		repo := &model.Repo{RegistryExtensionEndpoint: server.URL + "/registry/"}
		regs, err := m.RegistryServiceFromRepo(repo).RegistryListPipeline(t.Context(), repo, &model.Pipeline{}, nil)
		require.NoError(t, err)
		require.Len(t, regs, 2)
		assert.Equal(t, "ghcr.io", regs[0].Address)
		assert.Equal(t, "ext", regs[0].Username)
		assert.Equal(t, "docker.io", regs[1].Address)
	})
}

func TestConfigServiceFromRepo(t *testing.T) {
	server := extensionServer(t)
	forgeConfigs := []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("steps: [forge]")}}
	extensionConfigs := []*forge_types.FileMeta{{Name: "extension.yaml", Data: []byte("steps: []")}}

	t.Run("without extension the global service is used", func(t *testing.T) {
		base := config_service_mocks.NewMockService(t)
		m := &manager{config: base, client: testClient(t)}
		assert.Same(t, base, m.ConfigServiceFromRepo(&model.Repo{}))
	})

	t.Run("extension runs after the forge config", func(t *testing.T) {
		base := config_service_mocks.NewMockService(t)
		base.On("Fetch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(forgeConfigs, nil).Once()
		m := &manager{config: base, client: testClient(t)}

		repo := &model.Repo{ConfigExtensionEndpoint: server.URL + "/config/"}
		configs, err := m.ConfigServiceFromRepo(repo).Fetch(t.Context(), nil, &model.User{}, repo, &model.Pipeline{}, nil, false)
		require.NoError(t, err)
		assert.Equal(t, extensionConfigs, configs)
	})

	t.Run("exclusive extension skips the forge config", func(t *testing.T) {
		// no Fetch expectation, calling the base service fails the test
		base := config_service_mocks.NewMockService(t)
		m := &manager{config: base, client: testClient(t)}

		repo := &model.Repo{ConfigExtensionEndpoint: server.URL + "/config/", ConfigExtensionExclusive: true}
		configs, err := m.ConfigServiceFromRepo(repo).Fetch(t.Context(), nil, &model.User{}, repo, &model.Pipeline{}, nil, false)
		require.NoError(t, err)
		assert.Equal(t, extensionConfigs, configs)
	})
}

func TestForgeByID(t *testing.T) {
	newManager := func(s *store_mocks.MockStore, setup SetupForge) *manager {
		return &manager{
			store:      s,
			setupForge: setup,
			forgeCache: ttlcache.New(ttlcache.WithDisableTouchOnHit[int64, forge.Forge]()),
		}
	}

	t.Run("forges are set up once and then cached", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1, Type: model.ForgeTypeGitea}, nil).Once()
		f := forge_mocks.NewMockForge(t)
		setups := 0
		m := newManager(s, func(fm *model.Forge) (forge.Forge, error) {
			setups++
			assert.Equal(t, int64(1), fm.ID)
			return f, nil
		})

		got, err := m.ForgeByID(1)
		require.NoError(t, err)
		assert.Same(t, f, got)

		got, err = m.ForgeFromRepo(&model.Repo{ForgeID: 1})
		require.NoError(t, err)
		assert.Same(t, f, got)

		got, err = m.ForgeFromUser(&model.User{ForgeID: 1})
		require.NoError(t, err)
		assert.Same(t, f, got)

		assert.Equal(t, 1, setups)
	})

	t.Run("failed setups are not cached", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil).Twice()
		setups := 0
		m := newManager(s, func(*model.Forge) (forge.Forge, error) {
			setups++
			return nil, assert.AnError
		})

		_, err := m.ForgeByID(1)
		assert.ErrorIs(t, err, assert.AnError)
		_, err = m.ForgeByID(1)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, 2, setups)
	})

	t.Run("store error", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("ForgeGet", int64(2)).Return(nil, assert.AnError)
		m := newManager(s, func(*model.Forge) (forge.Forge, error) {
			t.Error("unexpected forge setup")
			return nil, nil
		})

		_, err := m.ForgeByID(2)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
