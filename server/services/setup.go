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
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/registry"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func setupRegistryService(store store.Store, dockerConfig, endpoint string, includeNetrc bool, client *utils.Client) registry.Service {
	var service registry.Service
	if dockerConfig != "" {
		service = registry.NewCombined(
			registry.NewDB(store),
			registry.NewFilesystem(dockerConfig),
		)
	} else {
		service = registry.NewDB(store)
	}

	// Wrap with global HTTP extension if configured
	if endpoint != "" {
		service = registry.NewWithExtension(service, registry.NewHTTP(endpoint, client, includeNetrc))
	}

	return service
}

func setupSecretService(store store.Store, endpoint string, client *utils.Client, includeNetrc bool) secret.Service {
	// TODO(1544): fix encrypted store
	// // encryption
	// encryptedSecretStore := encryptedStore.NewSecretStore(v)
	// err := encryption.Encryption(c, v).WithClient(encryptedSecretStore).Build()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("could not create encryption service")
	// }

	if endpoint != "" {
		return secret.NewCombined(secret.NewDB(store), secret.NewHTTP(endpoint, client, includeNetrc))
	}

	return secret.NewDB(store)
}

func setupConfigService(ctx context.Context, c *cli.Command, client *utils.Client) (config.Service, error) {
	timeout := c.Duration("forge-timeout")
	retries := c.Uint("forge-retry")
	if retries == 0 {
		return nil, fmt.Errorf("WOODPECKER_FORGE_RETRY can not be 0")
	}
	configFetcher := config.NewForge(timeout, retries, c.StringSlice("default-pipeline-configs"), c.StringSlice("default-pipeline-config-extensions"))

	extension, err := setupConfigExtension(ctx, c, client)
	if err != nil {
		return nil, err
	}
	if extension == nil {
		return configFetcher, nil
	}

	if c.Bool("config-extension-exclusive") {
		return extension, nil
	}
	return config.NewCombined(configFetcher, extension), nil
}

// setupConfigExtension returns the global configuration extension or nil if there is none.
func setupConfigExtension(ctx context.Context, c *cli.Command, client *utils.Client) (config.Service, error) {
	endpoint := c.String("config-extension-endpoint")
	wasmPath := c.String("config-extension-wasm")

	switch {
	case endpoint != "" && wasmPath != "":
		return nil, errors.New("WOODPECKER_CONFIG_EXTENSION_ENDPOINT and WOODPECKER_CONFIG_EXTENSION_WASM can not be used together")
	case endpoint != "":
		return config.NewHTTP(endpoint, client, c.Bool("config-extension-netrc")), nil
	case wasmPath != "":
		return setupConfigWasmExtension(ctx, wasmPath)
	}

	return nil, nil
}

func setupConfigWasmExtension(ctx context.Context, path string) (config.Service, error) {
	module, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read wasm config extension: %w", err)
	}

	// The runner lives as long as the server, so there is no need to close it.
	runner, err := wasm.NewRunner(ctx, module, wasm.DefaultLimits())
	if err != nil {
		return nil, fmt.Errorf("could not load wasm config extension '%s': %w", path, err)
	}

	log.Info().Str("path", path).Str("sha256", fmt.Sprintf("%x", sha256.Sum256(module))).Msg("loaded wasm config extension")

	return config.NewWasm(runner), nil
}

// setupSignatureKeys generate or load key pair to sign webhooks requests (i.e. used for service extensions).
func setupSignatureKeys(_store store.Store) (ed25519.PrivateKey, crypto.PublicKey, error) {
	privKeyID := "signature-private-key"

	privKey, err := _store.ServerConfigGet(privKeyID)
	if errors.Is(err, types.ErrRecordNotExist) {
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

func setupForgeService(c *cli.Command, _store store.Store) error {
	_forge, err := _store.ForgeGet(1)
	if err != nil && !errors.Is(err, types.ErrRecordNotExist) {
		return err
	}
	forgeExists := err == nil
	if _forge == nil {
		_forge = &model.Forge{
			ID: 0,
		}
	}
	if _forge.AdditionalOptions == nil {
		_forge.AdditionalOptions = make(map[string]any)
	}

	_forge.OAuthClientID = strings.TrimSpace(c.String("forge-oauth-client"))
	_forge.OAuthClientSecret = strings.TrimSpace(c.String("forge-oauth-secret"))
	_forge.URL = c.String("forge-url")
	_forge.SkipVerify = c.Bool("forge-skip-verify")
	_forge.OAuthHost = c.String("forge-oauth-host")

	switch {
	case c.String("addon-forge") != "":
		_forge.Type = model.ForgeTypeAddon
		_forge.AdditionalOptions["executable"] = c.String("addon-forge")
	case c.Bool("github"):
		_forge.Type = model.ForgeTypeGithub
		_forge.AdditionalOptions["merge-ref"] = c.Bool("github-merge-ref")
		_forge.AdditionalOptions["public-only"] = c.Bool("github-public-only")
		if _forge.URL == "" {
			_forge.URL = "https://github.com"
		}
	case c.Bool("gitlab"):
		_forge.Type = model.ForgeTypeGitlab
		if _forge.URL == "" {
			_forge.URL = "https://gitlab.com"
		}
	case c.Bool("gitea"):
		_forge.Type = model.ForgeTypeGitea
		if _forge.URL == "" {
			_forge.URL = "https://try.gitea.com"
		}
	case c.Bool("forgejo"):
		_forge.Type = model.ForgeTypeForgejo
		// TODO enable oauth URL with generic config option
		if _forge.URL == "" {
			_forge.URL = "https://next.forgejo.org"
		}
	case c.Bool("bitbucket"):
		_forge.Type = model.ForgeTypeBitbucket
	case c.Bool("bitbucket-dc"):
		_forge.Type = model.ForgeTypeBitbucketDatacenter
		_forge.AdditionalOptions["git-username"] = c.String("bitbucket-dc-git-username")
		_forge.AdditionalOptions["git-password"] = c.String("bitbucket-dc-git-password")
		_forge.AdditionalOptions["oauth-enable-project-admin-scope"] = c.Bool("bitbucket-dc-oauth-enable-oauth2-scope-project-admin")
	default:
		return errors.New("forge not configured")
	}

	if forgeExists {
		err := _store.ForgeUpdate(_forge)
		if err != nil {
			return err
		}
	} else {
		err := _store.ForgeCreate(_forge)
		if err != nil {
			return err
		}
	}

	return nil
}
