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

package extensions

import (
	"crypto"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/server/extensions/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/extensions/environment"
	"go.woodpecker-ci.org/woodpecker/v2/server/extensions/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/extensions/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type Manager struct {
	secret              secret.Extension
	registry            registry.Extension
	config              config.Extension
	environment         environment.Extension
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
		secret:              setupSecretExtension(store),
		registry:            setupRegistryExtension(store, c.String("docker-config")),
		config:              setupConfigExtension(c, signaturePrivateKey),
		environment:         environment.Parse(c.StringSlice("environment")),
	}, nil
}

func (e *Manager) SignaturePublicKey() crypto.PublicKey {
	return e.signaturePublicKey
}

func (e *Manager) SecretExtensionFromRepo(_ *model.Repo) secret.Extension {
	return e.secret
}

func (e *Manager) SecretExtension() secret.Extension {
	return e.secret
}

func (e *Manager) RegistryExtensionFromRepo(_ *model.Repo) registry.Extension {
	return e.registry
}

func (e *Manager) RegistryExtension() registry.Extension {
	return e.registry
}

func (e *Manager) ConfigExtensionFromRepo(_ *model.Repo) config.Extension {
	return e.config
}

func (e *Manager) EnvironmentExtension() environment.Extension {
	return e.environment
}
