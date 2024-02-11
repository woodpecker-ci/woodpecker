// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/registry"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func setupRegistryService(store store.Store, dockerConfig string) registry.Service {
	if dockerConfig != "" {
		return registry.NewCombined(
			registry.NewDB(store),
			registry.NewFilesystem(dockerConfig),
		)
	}

	return registry.NewDB(store)
}

func setupSecretService(store store.Store) secret.Service {
	// TODO(1544): fix encrypted store
	// // encryption
	// encryptedSecretStore := encryptedStore.NewSecretStore(v)
	// err := encryption.Encryption(c, v).WithClient(encryptedSecretStore).Build()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("could not create encryption service")
	// }

	return secret.NewDB(store)
}

func setupConfigService(c *cli.Context, privateSignatureKey crypto.PrivateKey) config.Service {
	timeout := c.Duration("forge-timeout")
	configFetcher := config.NewForge(timeout)

	if endpoint := c.String("config-extension-endpoint"); endpoint != "" {
		httpFetcher := config.NewHTTP(endpoint, privateSignatureKey)
		return config.NewCombined(configFetcher, httpFetcher)
	}

	return configFetcher
}

// setupSignatureKeys generate or load key pair to sign webhooks requests (i.e. used for service extensions)
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
