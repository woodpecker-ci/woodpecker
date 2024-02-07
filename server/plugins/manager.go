package extensions

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
	environ             environments.Service
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
}

func NewManager(store store.Store, forge forge.Forge, c *cli.Context) (*Manager, error) {
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
		environ:             environments.Parse(c.StringSlice("environment")),
	}, nil
}

func (e *Manager) SignaturePublicKey() crypto.PublicKey {
	return e.signaturePublicKey
}

func (e *Manager) SecretAddonFromRepo(repo *model.Repo) secret.Service {
	// if repo.SecretEndpoint != "" {
	// 	return secret.NewHTTP(repo.SecretEndpoint, e.signaturePrivateKey)
	// }

	return e.secret
}

func (e *Manager) RegistryAddonFromRepo(repo *model.Repo) registry.Service {
	// if repo.SecretEndpoint != "" {
	// 	return registry.NewHTTP(repo.SecretEndpoint, e.signaturePrivateKey)
	// }

	return e.registry
}

func (e *Manager) ConfigAddonFromRepo(repo *model.Repo) config.Service {
	// if repo.ConfigEndpoint != "" {
	// 	return config.NewHTTP(repo.ConfigEndpoint, e.signaturePrivateKey)
	// }

	return e.config
}

func (e *Manager) EnvironAddonFromRepo(repo *model.Repo) environments.Service {
	return e.environ
}
