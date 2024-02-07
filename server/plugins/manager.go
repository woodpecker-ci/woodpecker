package plugins

import (
	"crypto"

	"github.com/urfave/cli/v2"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/environments"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type Manager struct {
	secret              secret.Service
	registry            registry.Service
	config              config.Service
	environment         environments.Service
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
}

func NewManager(c *cli.Context, store store.Store, forge forge.Forge) (*Manager, error) {
	signaturePrivateKey, signaturePublicKey, err := setupSignatureKeys(store)
	if err != nil {
		return nil, err
	}

	config, err := setupConfigService(c, store, signaturePrivateKey)
	if err != nil {
		return nil, err
	}

	return &Manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		secret:              setupSecretExtension(store),
		registry:            setupRegistryExtension(store, c.String("docker-config")),
		config:              config,
		environment:         environments.Parse(c.StringSlice("environment")),
	}, nil
}

func (e *Manager) SignaturePublicKey() crypto.PublicKey {
	return e.signaturePublicKey
}

func (e *Manager) SecretServiceFromRepo(_ *model.Repo) secret.Service {
	return e.secret
}

func (e *Manager) SecretService() secret.Service {
	return e.secret
}

func (e *Manager) RegistryServiceFromRepo(_ *model.Repo) registry.Service {
	return e.registry
}

func (e *Manager) RegistryService() registry.Service {
	return e.registry
}

func (e *Manager) ConfigServiceFromRepo(_ *model.Repo) config.Service {
	return e.config
}

func (e *Manager) EnvironmentService() environments.Service {
	return e.environment
}
