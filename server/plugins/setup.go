package plugins

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon"
	addonTypes "go.woodpecker-ci.org/woodpecker/v2/shared/addon/types"
)

func setupRegistryExtension(store store.Store, dockerConfig string) registry.Service {
	if dockerConfig != "" {
		return registry.NewCombined(
			registry.NewDB(store),
			registry.NewFilesystem(dockerConfig),
		)
	}

	return registry.NewDB(store)
}

func setupSecretExtension(store store.Store) secret.Service {
	// TODO(1544): fix encrypted store
	// // encryption
	// encryptedSecretStore := encryptedStore.NewSecretStore(v)
	// err := encryption.Encryption(c, v).WithClient(encryptedSecretStore).Build()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("could not create encryption service")
	// }
	// server.Config.Services.Secrets = setupSecretService(c, encryptedSecretStore)

	return secret.NewDB(store)
}

func setupConfigService(c *cli.Context, store store.Store, privateSignatureKey crypto.PrivateKey) (config.Service, error) {
	addonExt, err := addon.Load[config.Service](c.StringSlice("addons"), addonTypes.TypeConfigService)
	if err != nil {
		return nil, err
	}
	if addonExt != nil {
		return addonExt.Value, nil
	}

	if endpoint := c.String("config-service-endpoint"); endpoint != "" {
		return config.NewHTTP(endpoint, privateSignatureKey), nil
	}

	return nil, nil
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

// func setupSecretService(c *cli.Context, s model.SecretStore) (model.SecretService, error) {
// 	addonService, err := addon.Load[model.SecretService](c.StringSlice("addons"), addonTypes.TypeSecretService)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if addonService != nil {
// 		return addonService.Value, nil
// 	}

// 	return secrets.New(c.Context, s), nil
// }

// func setupRegistryService(c *cli.Context, s store.Store) (model.RegistryService, error) {
// 	addonService, err := addon.Load[model.RegistryService](c.StringSlice("addons"), addonTypes.TypeRegistryService)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if addonService != nil {
// 		return addonService.Value, nil
// 	}

// 	if c.String("docker-config") != "" {
// 		return registry.Combined(
// 			registry.New(s),
// 			registry.Filesystem(c.String("docker-config")),
// 		), nil
// 	}
// 	return registry.New(s), nil
// }

// func setupEnvironService(c *cli.Context, _ store.Store) (model.EnvironService, error) {
// 	addonService, err := addon.Load[model.EnvironService](c.StringSlice("addons"), addonTypes.TypeEnvironmentService)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if addonService != nil {
// 		return addonService.Value, nil
// 	}

// 	return environments.Parse(c.StringSlice("environment")), nil
// }
