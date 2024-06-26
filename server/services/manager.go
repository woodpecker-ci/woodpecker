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
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/environment"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

//go:generate mockery --name Manager --output mocks --case underscore --note "+build test"

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
}

func NewManager(c *cli.Context, store store.Store, setupForge SetupForge) (Manager, error) {
	signaturePrivateKey, signaturePublicKey, err := setupSignatureKeys(store)
	if err != nil {
		return nil, err
	}

	err = setupForgeService(c, store)
	if err != nil {
		return nil, err
	}

	configService, err := setupConfigService(c, signaturePrivateKey)
	if err != nil {
		return nil, err
	}

	return &manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		store:               store,
		secret:              setupSecretService(store),
		registry:            setupRegistryService(store, c.String("docker-config")),
		config:              configService,
		environment:         environment.Parse(c.StringSlice("environment")),
		forgeCache:          ttlcache.New(ttlcache.WithDisableTouchOnHit[int64, forge.Forge]()),
		setupForge:          setupForge,
	}, nil
}

func (m *manager) SignaturePublicKey() crypto.PublicKey {
	return m.signaturePublicKey
}

func (m *manager) SecretServiceFromRepo(_ *model.Repo) secret.Service {
	return m.SecretService()
}

func (m *manager) SecretService() secret.Service {
	return m.secret
}

func (m *manager) RegistryServiceFromRepo(_ *model.Repo) registry.Service {
	return m.RegistryService()
}

func (m *manager) RegistryService() registry.Service {
	return m.registry
}

func (m *manager) ConfigServiceFromRepo(_ *model.Repo) config.Service {
	// TODO: decide based on repo property which config service to use
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
	item := m.forgeCache.Get(id)
	if item != nil && !item.IsExpired() {
		return item.Value(), nil
	}

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
