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
	"errors"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/google/tink/go/daead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/tink"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

type Encryption struct {
	store                store.Store
	encryption           tink.DeterministicAEAD
	primaryKeyId         string
	keysetFilePath       string
	keysetFileWatcher    *fsnotify.Watcher
	mixedDbCompatibility bool
}

func newEncryptionService(ctx *cli.Context, s store.Store) Encryption {
	filepath := ctx.String("secrets-encryption-keyset")
	mixedDb := ctx.Bool("secrets-encryption-mixed-db")

	if mixedDb {
		log.Warn().Msg("Mixed DB compatibility mode is on. Some secrets could be stored in plaintext until first read")
	}
	result := Encryption{s, nil, "",
		filepath, nil, mixedDb}
	result.reloadEncryption()
	result.initFileWatcher()

	return result
}

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
	if err != nil && svc.mixedDbCompatibility {
		// compatibility mode allows plaintext data in DB
		return ciphertext
	} else if err != nil && !svc.mixedDbCompatibility {
		log.Fatal().Err(err).Msgf("Base64 decryption error")
	}

	plaintext, err := svc.encryption.DecryptDeterministically(ct, []byte(associatedData))
	if err != nil && svc.mixedDbCompatibility {
		// compatibility mode allows plaintext data in DB
		return ciphertext
	} else if err != nil && !svc.mixedDbCompatibility {
		log.Fatal().Err(err).Msgf("Decryption error")
	}

	return string(plaintext)
}

// Watch keyset file events to detect key rotations and hot reload keys
func (svc *Encryption) initFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msgf("Error subscribing on encryption keyset file changes")
	}
	err = watcher.Add(svc.keysetFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error subscribing on encryption keyset file changes")
	}

	svc.keysetFileWatcher = watcher
	go svc.handleFileEvents()
}

func (svc *Encryption) handleFileEvents() {
	for {
		select {
		case event, ok := <-svc.keysetFileWatcher.Events:
			if !ok {
				log.Fatal().Msg("Error watching encryption keyset file changes")
			}
			if event.Op == fsnotify.Write {
				log.Info().Msgf("Modified encryption keyset file:", event.Name)
				svc.reloadEncryption()
			}
		case err, ok := <-svc.keysetFileWatcher.Errors:
			if !ok {
				log.Fatal().Err(err).Msgf("Error watching encryption keyset file changes")
			}
		}
	}
}

// Init and hot reload encryption primitive
func (svc *Encryption) reloadEncryption() {
	file, err := os.Open(svc.keysetFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error opening secret encryption keyset file")
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	jsonKeyset := keyset.NewJSONReader(file)
	keysetHandle, err := insecurecleartextkeyset.Read(jsonKeyset)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading secret encryption keyset")
	}
	svc.primaryKeyId = strconv.FormatUint(uint64(keysetHandle.KeysetInfo().PrimaryKeyId), 10)

	encryptionInstance, err := daead.New(keysetHandle)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error initializing secret encryption")
	}
	svc.encryption = encryptionInstance

	svc.validateKeyset()
}

// DB ciphertext sample
// store encrypted primaryKeyId in DB to check if used keyset is the same as used to encrypt secrets data
// and to detect keyset rotations
func (svc *Encryption) validateKeyset() {
	ciphertextSample, err := svc.store.ServerConfigGet("secrets-encryption-key-id")
	if errors.Is(err, types.RecordNotExist) {
		svc.updateCiphertextSample()
		return
	} else if err != nil {
		log.Fatal().Err(err).Msgf("Invalid secrets encryption key")
	}

	aad := "Primary key id"
	plaintext := svc.decrypt(ciphertextSample, aad)
	if err != nil {
		log.Fatal().Err(err).Msgf("Secrets encryption error")
	} else if plaintext != svc.primaryKeyId {
		svc.updateCiphertextSample()
	}
}

func (svc *Encryption) updateCiphertextSample() {
	aad := "Primary key id"
	ct := svc.encrypt(svc.primaryKeyId, aad)

	err := svc.store.ServerConfigSet("secrets-encryption-key-id", ct)
	if err != nil {
		log.Fatal().Err(err).Msgf("Storage error")
	}
}
