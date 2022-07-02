package extensions

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/extensions/config"
	"github.com/woodpecker-ci/woodpecker/server/extensions/environments"
	"github.com/woodpecker-ci/woodpecker/server/extensions/registry"
	"github.com/woodpecker-ci/woodpecker/server/extensions/secret"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore"
)

type Manager struct {
	secrets             secret.SecretExtension
	registries          registry.RegistryExtension
	config              config.Extension
	environ             environments.EnvironExtension
	signaturePrivateKey crypto.PrivateKey
	signaturePublicKey  crypto.PublicKey
}

func NewManager(store store.Store, remote remote.Remote, c *cli.Context) *Manager {
	signaturePrivateKey, signaturePublicKey := setupSignatureKeys(store)

	return &Manager{
		signaturePrivateKey: signaturePrivateKey,
		signaturePublicKey:  signaturePublicKey,
		secrets:             secret.NewBuiltin(store),
		registries:          setupRegistryExtension(store, c.String("docker-config")),
		config:              config.NewCombined(remote, c.String("config-service-endpoint"), signaturePrivateKey),
		environ:             environments.Parse(c.StringSlice("environment")),
	}
}

func setupRegistryExtension(store store.Store, dockerConfig string) registry.RegistryExtension {
	if dockerConfig != "" {
		return registry.NewCombined(
			registry.NewBuiltin(store),
			registry.NewFilesystem(dockerConfig),
		)
	}
	return registry.NewBuiltin(store)
}

// generate or load key pair to sign webhooks requests (i.e. used for extensions)
func setupSignatureKeys(_store store.Store) (crypto.PrivateKey, crypto.PublicKey) {
	privKeyID := "signature-private-key"

	privKey, err := _store.ServerConfigGet(privKeyID)
	if err != nil && err == datastore.RecordNotExist {
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to generate private key")
			return nil, nil
		}
		err = _store.ServerConfigSet(privKeyID, hex.EncodeToString(privKey))
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to generate private key")
			return nil, nil
		}
		log.Info().Msg("Created private key")
		return privKey, privKey.Public()
	} else if err != nil {
		log.Fatal().Err(err).Msgf("Failed to load private key")
		return nil, nil
	} else {
		privKeyStr, err := hex.DecodeString(privKey)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to decode private key")
			return nil, nil
		}
		privKey := ed25519.PrivateKey(privKeyStr)
		return privKey, privKey.Public()
	}
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

func (e *Manager) RegistriesFromRepo(repo *model.Repo) registry.RegistryExtension {
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
