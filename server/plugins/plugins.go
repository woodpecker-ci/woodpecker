package extensions

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/environments"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

type Manager struct {
	secrets             secret.SecretExtension
	registries          registry.RegistryService
	config              config.Extension
	environ             environments.EnvironPlugin
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
}

func NewManager(store store.Store, forge forge.Forge, c *cli.Context) (*Manager, error) {
	signaturePrivateKey, signaturePublicKey, err := setupSignatureKeys(store)
	if err != nil {
		return nil, err
	}

	return &Manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		secrets:             secret.NewBuiltin(store),
		registries:          setupRegistryExtension(store, c.String("docker-config")),
		config:              config.NewCombined(forge, c.String("config-service-endpoint"), signaturePrivateKey),
		environ:             environments.Parse(c.StringSlice("environment")),
	}, nil
}

func setupRegistryExtension(store store.Store, dockerConfig string) registry.RegistryService {
	if dockerConfig != "" {
		return registry.NewCombined(
			registry.NewBuiltin(store),
			registry.NewFilesystem(dockerConfig),
		)
	}
	return registry.NewBuiltin(store)
}

// setupSignatureKeys generate or load key pair to sign webhooks requests (i.e. used for extensions)
func setupSignatureKeys(_store store.Store) (crypto.PrivateKey, crypto.PublicKey, error) {
	privKeyID := "signature-private-key"

	privKey, err := _store.ServerConfigGet(privKeyID)
	if errors.Is(err, types.RecordNotExist) {
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
		}
		err = _store.ServerConfigSet(privKeyID, hex.EncodeToString(privKey))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to store private key: %w", err)
		}
		log.Debug().Msg("created private key")
		return privKey, privKey.Public(), nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to load private key: %w", err)
	}
	privKeyStr, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	privateKey := ed25519.PrivateKey(privKeyStr)
	return privateKey, privateKey.Public(), nil
}

func (e *Manager) SignaturePublicKey() crypto.PublicKey {
	return e.signaturePublicKey
}

func (e *Manager) SecretsFromRepo(repo *model.Repo) secret.SecretExtension {
	if repo.SecretEndpoint != "" {
		return secret.NewHTTP(repo.SecretEndpoint, e.signaturePrivateKey)
	}

	return e.secrets
}

func (e *Manager) RegistriesFromRepo(repo *model.Repo) registry.RegistryService {
	if repo.SecretEndpoint != "" {
		return registry.NewHTTP(repo.SecretEndpoint, e.signaturePrivateKey)
	}

	return e.registries
}

func (e *Manager) Config() config.Extension {
	return e.config
}

func (e *Manager) ConfigExtensionsFromRepo(repo *model.Repo) *config.HttpFetcher {
	if repo.ConfigEndpoint != "" {
		return config.NewHTTP(repo.ConfigEndpoint, e.signaturePrivateKey)
	}

	return nil
}

func (e *Manager) Environ() environments.EnvironExtension {
	return e.environ
}
