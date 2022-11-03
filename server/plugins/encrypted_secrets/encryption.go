// Copyright 2022 Woodpecker Authors
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

package encrypted_secrets

import (
	"encoding/base64"
	"github.com/fsnotify/fsnotify"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"strconv"

	"github.com/google/tink/go/tink"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type Encryption struct {
	store             store.Store
	encryption        tink.DeterministicAEAD
	primaryKeyId      string
	keysetFilePath    string
	keysetFileWatcher *fsnotify.Watcher
}

func newEncryptionService(ctx *cli.Context, s store.Store) Encryption {
	filepath := ctx.String("secrets-encryption-keyset")

	result := Encryption{s, nil, "", filepath, nil}
	result.initEncryption()
	result.initFileWatcher()

	return result
}

// Basic encrypt-decrypt functions
func (svc *Encryption) encrypt(plaintext string, associatedData string) string {
	msg := []byte(plaintext)
	aad := []byte(associatedData)
	ciphertext, err := svc.encryption.EncryptDeterministically(msg, aad)
	if err != nil {
		log.Fatal().Err(err).Msgf("Encryption error")
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func (svc *Encryption) decrypt(ciphertext string, associatedData string) string {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		log.Fatal().Err(err).Msgf("Secrets encryption: Base64 decryption error")
	}

	plaintext, err := svc.encryption.DecryptDeterministically(ct, []byte(associatedData))
	if err != nil {
		log.Fatal().Err(err).Msgf("Decryption error")
	}

	return string(plaintext)
}

// Secret-specific functions
func (svc *Encryption) encryptSecret(secret *model.Secret) {
	encryptedValue := svc.encrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	secret.Value = encryptedValue
}

func (svc *Encryption) encryptSecretList(secrets []*model.Secret) {
	for _, secret := range secrets {
		svc.encryptSecret(secret)
	}
}

func (svc *Encryption) decryptSecret(secret *model.Secret) {
	decryptedValue := svc.decrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	secret.Value = decryptedValue
}

func (svc *Encryption) decryptSecretList(secrets []*model.Secret) {
	for _, secret := range secrets {
		svc.decryptSecret(secret)
	}
}
