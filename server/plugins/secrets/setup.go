// Copyright 2023 Woodpecker Authors
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

package secrets

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type EncryptionMode string

const (
	EncryptionModeDisabled           EncryptionMode = "Disabled"
	EncryptionModeDisabledAndDecrypt EncryptionMode = "DisabledAndDecrypt"
	EncryptionModeEnabled            EncryptionMode = "Enabled"
	EncryptionModeEnabledAndEncrypt  EncryptionMode = "EnabledAndEncrypt"
)

func NewService(ctx *cli.Context, store model.SecretStore) (model.SecretService, error) {
	secretSvc := New(ctx.Context, store)

	encryptionMode := EncryptionMode(ctx.String("secrets-encryption-mode"))
	log.Debug().Str("mode", string(encryptionMode)).Msg("setting up secrets service")

	if encryptionMode == EncryptionModeEnabled {
		ess := NewEncrypted(secretSvc, server.Config.Services.Encryption)
		return &ess, nil
	}

	if encryptionMode == EncryptionModeEnabledAndEncrypt {
		ess := NewEncrypted(secretSvc, server.Config.Services.Encryption)
		err := encryptAll(store, &ess)
		if err != nil {
			return nil, err
		}
		return &ess, nil
	}

	if encryptionMode == EncryptionModeDisabledAndDecrypt {
		ess := NewEncrypted(secretSvc, server.Config.Services.Encryption)
		err := decryptAll(store, &ess)
		if err != nil {
			return nil, err
		}
		return &ess, nil
	}

	return secretSvc, nil
}

func encryptAll(store model.SecretStore, ess *EncryptedSecretService) error {
	log.Debug().Msg("encrypting secrets")

	secrets, err := store.SecretListAll()
	if err != nil {
		return NewAllEncryptionError(err)
	}

	for _, secret := range secrets {
		if err := ess.EncryptSecret(secret); err != nil {
			return NewAllEncryptionError(err)
		}
		if err := store.SecretUpdate(secret); err != nil {
			return NewAllEncryptionError(err)
		}
	}

	return nil
}

func decryptAll(store model.SecretStore, ess *EncryptedSecretService) error {
	log.Debug().Msg("decrypting secrets")

	secrets, err := store.SecretListAll()
	if err != nil {
		return NewAllDecryptionError(err)
	}

	for _, secret := range secrets {
		if err := ess.DecryptSecret(secret); err != nil {
			return NewAllDecryptionError(err)
		}
		if err := store.SecretUpdate(secret); err != nil {
			return NewAllDecryptionError(err)
		}
	}

	return nil
}

func NewAllEncryptionError(e error) error {
	return fmt.Errorf("cannot encrypt secrets: %w", e)
}

func NewAllDecryptionError(e error) error {
	return fmt.Errorf("cannot decrypt secrets: %w", e)
}
