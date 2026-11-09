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
	forgeMu             sync.Mutex
	forgeLocks          map[int64]*forgeLock
	forges              sync.Map
}

type forgeLock struct {
	mu   sync.Mutex
	refs int
}

func (m *manager) lockForge(id int64) func() {
	m.forgeMu.Lock()
	if m.forgeLocks == nil {
		m.forgeLocks = make(map[int64]*forgeLock)
	}
	lock := m.forgeLocks[id]
	if lock == nil {
		lock = &forgeLock{}
		m.forgeLocks[id] = lock
	}
	lock.refs++
	m.forgeMu.Unlock()

	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		m.forgeMu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(m.forgeLocks, id)
		}
		m.forgeMu.Unlock()
	}
}

func closeForge(f forge.Forge) {
	if err := f.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close forge")
	}
}

func closeStoredForge(value any) {
	f, ok := value.(forge.Forge)
	if !ok {
		log.Error().Msg("failed to close forge with unexpected type")
		return
	}
	closeForge(f)
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
		forgeCache:          ttlcache.New(ttlcache.WithDisableTouchOnHit[int64, forge.Forge]()),
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
	unlock := m.lockForge(id)
	defer unlock()

	item := m.forgeCache.Get(id)
	if item != nil && !item.IsExpired() {
		return item.Value(), nil
	}
	if old, ok := m.forges.Load(id); ok {
		closeStoredForge(old)
		m.forges.Delete(id)
	} else if item != nil {
		closeForge(item.Value())
	}
	if item != nil {
		m.forgeCache.Delete(id)
	}

	forgeModel, err := m.store.ForgeGet(id)
	if err != nil {
		return nil, err
	}

	forge, err := m.setupForge(forgeModel)
	if err != nil {
		return nil, err
	}

	m.forges.Store(id, forge)
	m.forgeCache.Set(id, forge, forgeCacheTTL)

	return forge, nil
}

func (m *manager) RemoveForge(id int64) {
	unlock := m.lockForge(id)
	defer unlock()

	if old, ok := m.forges.Load(id); ok {
		closeStoredForge(old)
		m.forges.Delete(id)
	} else if item := m.forgeCache.Get(id); item != nil {
		closeForge(item.Value())
	}
	m.forgeCache.Delete(id)
}
