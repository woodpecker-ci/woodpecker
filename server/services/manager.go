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

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/environment"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type Manager struct {
	secret              secret.Service
	registry            registry.Service
	config              config.Service
	environment         environment.Service
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
}

func NewManager(c *cli.Context, store store.Store) (*Manager, error) {
	signaturePrivateKey, signaturePublicKey, err := setupSignatureKeys(store)
	if err != nil {
		return nil, err
	}

	return &Manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		secret:              setupSecretService(store),
		registry:            setupRegistryService(store, c.String("docker-config")),
		config:              setupConfigService(c, signaturePrivateKey),
		environment:         environment.Parse(c.StringSlice("environment")),
	}, nil
}

func (e *Manager) SignaturePublicKey() crypto.PublicKey {
	return e.signaturePublicKey
}

func (e *Manager) SecretServiceFromRepo(_ *model.Repo) secret.Service {
	return e.SecretService()
}

func (e *Manager) SecretService() secret.Service {
	return e.secret
}

func (e *Manager) RegistryServiceFromRepo(_ *model.Repo) registry.Service {
	return e.RegistryService()
}

func (e *Manager) RegistryService() registry.Service {
	return e.registry
}

func (e *Manager) ConfigServiceFromRepo(_ *model.Repo) config.Service {
	// TODO: decied based on repo property which config service to use
	return e.config
}

func (e *Manager) EnvironmentService() environment.Service {
	return e.environment
}
