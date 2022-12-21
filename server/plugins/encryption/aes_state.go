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

import "github.com/rs/zerolog/log"

func (svc *aesEncryptionService) initClients() {
	for _, client := range svc.clients {
		client.SetEncryptionService(svc)
	}
	log.Info().Msg("initialized encryption on registered services")
}

func (svc *aesEncryptionService) enable() {
	svc.callbackOnEnable()
	svc.updateCiphertextSample()
	log.Warn().Msg("encryption enabled")
}

func (svc *aesEncryptionService) disable() {
	svc.callbackOnDisable()
	svc.deleteCiphertextSample()
	log.Warn().Msg("encryption disabled")
}

func (svc *aesEncryptionService) updateCiphertextSample() {
	ciphertext := svc.Encrypt(svc.keyId, keyIdAAD)
	err := svc.store.ServerConfigSet(ciphertextSampleConfigKey, ciphertext)
	if err != nil {
		log.Fatal().Err(err).Msgf("updating encryption key failed: could not update server config")
	}
	log.Info().Msg("registered new encryption key")
}

func (svc *aesEncryptionService) deleteCiphertextSample() {
	err := svc.store.ServerConfigDelete(ciphertextSampleConfigKey)
	if err != nil {
		log.Fatal().Err(err).Msg("disabling encryption failed: could not update server config")
	}
}

func (svc *aesEncryptionService) callbackOnEnable() {
	for _, client := range svc.clients {
		client.EnableEncryption()
	}
	log.Info().Msg("enabled encryption on registered services")
}

func (svc *aesEncryptionService) callbackOnDisable() {
	for _, client := range svc.clients {
		client.MigrateEncryption(&noEncryption{})
	}
	log.Info().Msg("Disabled encryption on registered services")
}
