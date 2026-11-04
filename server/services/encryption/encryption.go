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

// Package encryption encrypts secret values, registry passwords and user
// forge tokens at rest.
package encryption

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/tink-crypto/tink-go/v2/tink"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

const (
	disableFlag = "encryption-disable-flag"

	// Server config holding the id of the key the data is encrypted with,
	// encrypted with that key. It exists exactly while encryption is enabled.
	keyIDConfigKey      = "encryption-ciphertext-sample"
	keyIDAssociatedData = "Primary key id"
)

// Setup brings the data at rest into the configured state and returns s
// wrapped to keep it there:
//   - without a key everything stays plaintext; encrypted data stops the start
//   - with a key every value gets encrypted with its primary key, including
//     values another server wrote with an older key meanwhile; a value whose
//     key is not in the keyset stops the start
//   - with a key and the disable flag everything gets decrypted
//
// Each step can be interrupted: the next start continues where it stopped.
func Setup(c *cli.Command, s store.Store) (store.Store, error) {
	key, keyID, err := loadKey(c)
	if err != nil {
		return nil, fmt.Errorf("could not load encryption key: %w", err)
	}

	storedKeyID, err := s.ServerConfigGet(keyIDConfigKey)
	enabled := err == nil
	if err != nil && !errors.Is(err, types.ErrRecordNotExist) {
		return nil, err
	}

	if key == nil {
		if enabled {
			return nil, errors.New("encryption is enabled, but no encryption key is configured")
		}
		return &encryptedStore{Store: s}, nil
	}

	usedKeyID := keyID
	if enabled {
		if usedKeyID, err = decrypt(key, storedKeyID, keyIDAssociatedData); err != nil {
			return nil, fmt.Errorf("encryption key does not match the key the data is encrypted with: %w", err)
		}
	}

	if c.Bool(disableFlag) {
		return disable(s, key, enabled)
	}
	return enable(s, key, keyID, enabled, usedKeyID)
}

func enable(s store.Store, key tink.AEAD, keyID string, enabled bool, usedKeyID string) (store.Store, error) {
	// record the key before the data, so a start without it fails instead
	// of serving ciphertext
	if !enabled {
		log.Warn().Msg("enabling encryption")
		if err := saveKeyID(s, key, keyID); err != nil {
			return nil, err
		}
	}
	if usedKeyID != keyID {
		log.Warn().Msg("re-encrypting data with the new primary key")
	}

	es := &encryptedStore{Store: s, key: key, keyID: keyID}
	if err := es.rewrite(); err != nil {
		return nil, fmt.Errorf("could not encrypt data: %w", err)
	}

	if usedKeyID != keyID {
		if err := saveKeyID(s, key, keyID); err != nil {
			return nil, err
		}
	}
	return es, nil
}

func disable(s store.Store, key tink.AEAD, enabled bool) (store.Store, error) {
	log.Warn().Msg("disabling encryption")
	if err := (&encryptedStore{Store: s, key: key}).rewrite(); err != nil {
		return nil, fmt.Errorf("could not decrypt data: %w", err)
	}

	// forget the key only after all data is decrypted
	if enabled {
		if err := s.ServerConfigDelete(keyIDConfigKey); err != nil {
			return nil, err
		}
	}
	return &encryptedStore{Store: s}, nil
}

func saveKeyID(s store.Store, key tink.AEAD, keyID string) error {
	encrypted, err := encrypt(key, keyID, keyIDAssociatedData)
	if err != nil {
		return err
	}
	return s.ServerConfigSet(keyIDConfigKey, encrypted)
}
