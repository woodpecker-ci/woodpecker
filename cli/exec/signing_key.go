// Copyright 2026 Woodpecker Authors
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

package exec

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/yaronf/httpsign"
	"github.com/zalando/go-keyring"

	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

const (
	// configExtensionKeyringService is the CLI name tokens are stored under.
	configExtensionKeyringService = "woodpecker-cli"
	// configExtensionKeyringUser keeps the signing key next to tokens, under its own account.
	configExtensionKeyringUser = "config-extension-signing-key"
	// configExtensionSignatureName is the HTTP signature label. It is not the
	// server's "woodpecker-ci-extensions", so an extension can trust both keys.
	configExtensionSignatureName = "woodpecker-cli-extensions"
)

// configExtensionPrivateKey loads the ed25519 key used to sign config extension
// requests, generating and storing it in the system keyring on first use.
// The stored value is the hex encoding of the private key, the same encoding
// the server uses for its own signing key.
func configExtensionPrivateKey() (ed25519.PrivateKey, error) {
	stored, err := keyring.Get(configExtensionKeyringService, configExtensionKeyringUser)
	switch {
	case errors.Is(err, keyring.ErrNotFound):
		return generateConfigExtensionPrivateKey()
	case errors.Is(err, keyring.ErrUnsupportedPlatform):
		return nil, fmt.Errorf("cannot store the config extension signing key: %w", err)
	case err != nil:
		return nil, fmt.Errorf("could not read the config extension signing key: %w", err)
	}

	raw, err := hex.DecodeString(stored)
	if err != nil {
		return nil, fmt.Errorf("could not decode the stored config extension signing key: %w", err)
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("stored config extension signing key has length %d, want %d", len(raw), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(raw), nil
}

func generateConfigExtensionPrivateKey() (ed25519.PrivateKey, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not generate the config extension signing key: %w", err)
	}
	err = keyring.Set(configExtensionKeyringService, configExtensionKeyringUser, hex.EncodeToString(privateKey))
	if err != nil {
		return nil, fmt.Errorf("could not store the config extension signing key: %w", err)
	}
	log.Debug().Msg("generated a new config extension signing key")
	return privateKey, nil
}

// configExtensionPublicKeyPEM encodes the public key the way the server encodes
// its own at GET /api/signature/public-key, so an extension can parse either.
func configExtensionPublicKeyPEM(privateKey ed25519.PrivateKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return "", fmt.Errorf("could not encode the config extension public key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// provideConfigExtensionPublicKey prints the public key so the extension can be
// configured to trust requests from this CLI.
func provideConfigExtensionPublicKey(privateKey ed25519.PrivateKey) error {
	pemKey, err := configExtensionPublicKeyPEM(privateKey)
	if err != nil {
		return err
	}
	log.Info().Str("keyid", configExtensionSignatureName).Msg("signing config extension requests; configure the extension to trust this public key")
	if _, err := fmt.Fprint(os.Stderr, pemKey); err != nil {
		return fmt.Errorf("could not write the config extension public key: %w", err)
	}
	return nil
}

// newConfigExtensionClient signs requests the way the server signs extension
// calls: ed25519 HTTP signatures over @request-target and content-digest.
func newConfigExtensionClient(privateKey ed25519.PrivateKey) (*httpsign.Client, error) {
	signer, err := httpsign.NewEd25519Signer(
		privateKey,
		httpsign.NewSignConfig(),
		httpsign.Headers("@request-target", "content-digest"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create the config extension signer: %w", err)
	}
	httpClient := http.Client{Transport: httputil.NewUserAgentRoundTripper(nil, "cli")}
	config := httpsign.NewClientConfig().
		SetSignatureName(configExtensionSignatureName).
		SetSigner(signer)
	return httpsign.NewClient(httpClient, config), nil
}
