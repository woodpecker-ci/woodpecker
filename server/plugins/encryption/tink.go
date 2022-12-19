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

package encryption

import (
	"encoding/base64"
	"github.com/fsnotify/fsnotify"
	"github.com/google/tink/go/tink"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

const keyIdAAD = "Primary key id"

type tinkEncryptionService struct {
	keysetFilePath    string
	primaryKeyId      string
	encryption        tink.AEAD
	store             store.Store
	keysetFileWatcher *fsnotify.Watcher
	clients           []model.EncryptionClient
}

func (svc *tinkEncryptionService) Encrypt(plaintext string, associatedData string) string {
	msg := []byte(plaintext)
	aad := []byte(associatedData)
	ciphertext, err := svc.encryption.Encrypt(msg, aad)
	if err != nil {
		log.Fatal().Err(err).Msgf("Encryption error")
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func (svc *tinkEncryptionService) Decrypt(ciphertext string, associatedData string) string {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		log.Fatal().Err(err).Msgf("Secrets encryption: Base64 decryption error")
	}

	plaintext, err := svc.encryption.Decrypt(ct, []byte(associatedData))
	if err != nil {
		log.Fatal().Err(err).Msgf("Decryption error")
	}

	return string(plaintext)
}

func (svc *tinkEncryptionService) Disable() {
	svc.disable()
}
