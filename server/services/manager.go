// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"crypto"
	"strings"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/environment"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/registry"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

const forgeCacheTTL = 10 * time.Minute

type SetupForge func(forge *model.Forge) (forge.Forge, error)

type Manager interface {
	SignaturePublicKey() crypto.PublicKey
	SecretServiceFromRepo(repo *model.Repo) secret.Service
	SecretService() secret.Service
	RegistryServiceFromRepo(repo *model.Repo) registry.Service
	RegistryService() registry.Service
	ConfigServiceFromRepo(repo *model.Repo) config.Service
	EnvironmentService() environment.Service
	ForgeFromRepo(repo *model.Repo) (forge.Forge, error)
	ForgeFromUser(user *model.User) (forge.Forge, error)
	ForgeByID(forgeID int64) (forge.Forge, error)
	RemoveForge(forgeID int64)
}

type manager struct {
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
	store               store.Store
	secret              secret.Service
	registry            registry.Service
	config              config.Service
	environment         environment.Service
	forgeCache          *ttlcache.Cache[int64, forge.Forge]
	setupForge          SetupForge
	client              *utils.Client
	// forgeMu serializes creating and removing forge instances, so that
	// no instance gets replaced in the cache without being closed.
	forgeMu sync.RWMutex
}

// newForgeCache returns a forge cache that closes every forge it evicts.
// Eviction callbacks run in their own goroutine, so a slow Close (an addon
// waiting for in-flight calls) does not block the caller.
func newForgeCache() *ttlcache.Cache[int64, forge.Forge] {
	c := ttlcache.New(ttlcache.WithDisableTouchOnHit[int64, forge.Forge]())
	c.OnEviction(func(_ context.Context, _ ttlcache.EvictionReason, item *ttlcache.Item[int64, forge.Forge]) {
		if err := item.Value().Close(); err != nil {
			log.Error().Err(err).Int64("forge-id", item.Key()).Msg("failed to close forge")
		}
	})
	return c
}

func NewManager(c *cli.Command, store store.Store, setupForge SetupForge) (Manager, error) {
	signaturePrivateKey, signaturePublicKey, err := setupSignatureKeys(store)
	if err != nil {
		return nil, err
	}

	err = setupForgeService(c, store)
	if err != nil {
		return nil, err
	}

	client, err := utils.NewHTTPClient(signaturePrivateKey, c.String("extensions-allowed-hosts"))
	if err != nil {
		return nil, err
	}

	configService, err := setupConfigService(c, client)
	if err != nil {
		return nil, err
	}

	return &manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		store:               store,
		secret:              setupSecretService(store, c.String("secret-extension-endpoint"), client, c.Bool("secret-extension-netrc")),
		registry:            setupRegistryService(store, c.String("docker-config"), c.String("registry-extension-endpoint"), c.Bool("registry-extension-netrc"), client),
		config:              configService,
		environment:         environment.Parse(c.StringSlice("environment")),
		forgeCache:          newForgeCache(),
		setupForge:          setupForge,
		client:              client,
	}, nil
}

func (m *manager) SignaturePublicKey() crypto.PublicKey {
	return m.signaturePublicKey
}

func (m *manager) SecretServiceFromRepo(repo *model.Repo) secret.Service {
	if repo.SecretExtensionEndpoint != "" {
		return secret.NewCombined(m.secret, secret.NewHTTP(strings.TrimRight(repo.SecretExtensionEndpoint, "/"), m.client, repo.SecretExtensionNetrc))
	}

	return m.SecretService()
}

func (m *manager) SecretService() secret.Service {
	return m.secret
}

func (m *manager) RegistryServiceFromRepo(repo *model.Repo) registry.Service {
	if repo.RegistryExtensionEndpoint != "" {
		return registry.NewWithExtension(m.registry, registry.NewHTTP(strings.TrimRight(repo.RegistryExtensionEndpoint, "/"), m.client, repo.RegistryExtensionNetrc))
	}
	return m.RegistryService()
}

func (m *manager) RegistryService() registry.Service {
	return m.registry
}

func (m *manager) ConfigServiceFromRepo(repo *model.Repo) config.Service {
	if repo.ConfigExtensionEndpoint != "" {
		if repo.ConfigExtensionExclusive {
			return config.NewHTTP(strings.TrimRight(repo.ConfigExtensionEndpoint, "/"), m.client, repo.ConfigExtensionNetrc)
		}
		return config.NewCombined(m.config, config.NewHTTP(strings.TrimRight(repo.ConfigExtensionEndpoint, "/"), m.client, repo.ConfigExtensionNetrc))
	}

	return m.config
}

func (m *manager) EnvironmentService() environment.Service {
	return m.environment
}

func (m *manager) ForgeFromRepo(repo *model.Repo) (forge.Forge, error) {
	return m.ForgeByID(repo.ForgeID)
}

func (m *manager) ForgeFromUser(user *model.User) (forge.Forge, error) {
	return m.ForgeByID(user.ForgeID)
}

func (m *manager) ForgeByID(id int64) (forge.Forge, error) {
	// First do easy path where forge is loaded and usable
	m.forgeMu.RLock()
	if item := m.forgeCache.Get(id); item != nil {
		defer m.forgeMu.RUnlock()
		return item.Value(), nil
	}
	m.forgeMu.RUnlock()

	// Now we need write lock to clean and load forge
	m.forgeMu.Lock()
	defer m.forgeMu.Unlock()

	// another call may have created the forge while we waited for the lock
	if item := m.forgeCache.Get(id); item != nil {
		return item.Value(), nil
	}

	// Get does not return expired items, but they stay in the cache and Set
	// would overwrite them without an eviction. Delete evicts (and closes) them.
	m.forgeCache.Delete(id)

	forgeModel, err := m.store.ForgeGet(id)
	if err != nil {
		return nil, err
	}

	forge, err := m.setupForge(forgeModel)
	if err != nil {
		return nil, err
	}

	m.forgeCache.Set(id, forge, forgeCacheTTL)

	return forge, nil
}

func (m *manager) RemoveForge(id int64) {
	m.forgeMu.Lock()
	defer m.forgeMu.Unlock()

	m.forgeCache.Delete(id)
}
