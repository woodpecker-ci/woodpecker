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
	"github.com/rs/zerolog/log"
)

func (svc *tinkEncryptionService) enable() {
	svc.callbackOnEnable()
	svc.updateCiphertextSample()
	log.Warn().Msg("Secrets encryption enabled on server")
}

func (svc *tinkEncryptionService) disable() {
	svc.callbackOnDisable()
	svc.deleteCiphertextSample()
	log.Warn().Msg("Secrets encryption disabled")
}

func (svc *tinkEncryptionService) rotate() {
	newSvc := &tinkEncryptionService{
		keysetFilePath:    svc.keysetFilePath,
		primaryKeyId:      "",
		encryption:        nil,
		store:             svc.store,
		keysetFileWatcher: nil,
		clients:           svc.clients,
	}
	newSvc.loadKeyset()

	err := newSvc.validateKeyset()
	if err == encryptionKeyRotatedError {
		newSvc.updateCiphertextSample()
	} else if err != nil {
		log.Fatal().Err(err).Msgf("rotated encryption key validation failed")
	}

	newSvc.callbackOnRotation()
	newSvc.initFileWatcher()
}

func (svc *tinkEncryptionService) updateCiphertextSample() {
	ciphertext := svc.Encrypt(svc.primaryKeyId, keyIdAAD)
	err := svc.store.ServerConfigSet(ciphertextSampleConfigKey, ciphertext)
	if err != nil {
		log.Fatal().Err(err).Msgf("Rotating secrets encryption key failed: could not update server config")
	}
	log.Info().Msg("Registered rotated secrets encryption key")
}

func (svc *tinkEncryptionService) deleteCiphertextSample() {
	err := svc.store.ServerConfigDelete(ciphertextSampleConfigKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Disabling secrets encryption failed: could not update server config")
	}
}

func (svc *tinkEncryptionService) initClients() {
	for _, client := range svc.clients {
		client.InitEncryption(svc)
	}
	log.Info().Msg("Initialized encryption on registered services")
}

func (svc *tinkEncryptionService) callbackOnEnable() {
	for _, client := range svc.clients {
		client.EnableEncryption()
	}
	log.Info().Msg("Enabled secrets encryption on registered services")
}

func (svc *tinkEncryptionService) callbackOnRotation() {
	for _, client := range svc.clients {
		client.MigrateEncryption(svc)
	}
	log.Info().Msg("Rotated secrets encryption key on registered services")
}

func (svc *tinkEncryptionService) callbackOnDisable() {
	for _, client := range svc.clients {
		client.MigrateEncryption(&noEncryption{})
	}
	log.Info().Msg("Disabled secrets encryption on registered services")
}
