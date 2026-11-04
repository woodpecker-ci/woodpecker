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

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
	"encoding/hex"
	"errors"
	"os"
	"strconv"

	"github.com/tink-crypto/tink-go/v2/aead"
	insecure_clear_text_keyset "github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
	"github.com/urfave/cli/v3"
)

const (
	rawKeyFlag     = "encryption-raw-key"
	tinkKeysetFlag = "encryption-tink-keyset"

	keyIDDomainSeparation = "woodpecker-encryption-key-id/v1"
	keySize               = 32 // AES-256
)

// loadKey returns the configured key and the id of the key it encrypts with,
// or a nil key if none is configured. A new tink primary key changes the id.
func loadKey(c *cli.Command) (tink.AEAD, string, error) {
	rawKey, keysetFile := c.String(rawKeyFlag), c.String(tinkKeysetFlag)
	switch {
	case rawKey != "" && keysetFile != "":
		return nil, "", errors.New("cannot use raw encryption key and tink keyset at the same time")
	case rawKey != "":
		return rawAESKey(rawKey)
	case keysetFile != "":
		return tinkKey(keysetFile)
	}
	return nil, "", nil
}

// rawAESKey derives an AES-256-GCM key from the raw key. Its id is derived
// with domain separation, so it reveals nothing about the key.
func rawAESKey(rawKey string) (tink.AEAD, string, error) {
	key := sha3.SumSHAKE256([]byte(rawKey), keySize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, "", err
	}
	id := sha3.SumSHAKE256(append([]byte(keyIDDomainSeparation), key...), keySize)
	return aesGCM{gcm}, hex.EncodeToString(id), nil
}

// aesGCM adapts AES-GCM with random nonces to tink.AEAD.
type aesGCM struct {
	cipher.AEAD
}

func (a aesGCM) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	return a.Seal(nil, nil, plaintext, associatedData), nil
}

func (a aesGCM) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	return a.Open(nil, nil, ciphertext, associatedData)
}

// tinkKey loads a tink AEAD keyset from a JSON file. Its id is the id of the
// primary key.
func tinkKey(path string) (tink.AEAD, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	handle, err := insecure_clear_text_keyset.Read(keyset.NewJSONReader(file))
	if err != nil {
		return nil, "", err
	}
	key, err := aead.New(handle)
	if err != nil {
		return nil, "", err
	}
	return key, strconv.FormatUint(uint64(handle.KeysetInfo().GetPrimaryKeyId()), 10), nil
}
