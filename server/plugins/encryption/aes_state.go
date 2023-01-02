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
	"fmt"

	"github.com/rs/zerolog/log"
)

func (svc *aesEncryptionService) initClients() error {
	for _, client := range svc.clients {
		err := client.SetEncryptionService(svc)
		if err != nil {
			return fmt.Errorf("failed initializing encryption clients with AES encryption: %w", err)
		}
	}
	log.Info().Msg("initialized encryption on registered services")
	return nil
}

func (svc *aesEncryptionService) enable() error {
	err := svc.callbackOnEnable()
	if err != nil {
		return fmt.Errorf("failed enabling AES encryption: %w", err)
	}
	err = svc.updateCiphertextSample()
	if err != nil {
		return fmt.Errorf("failed enabling AES encryption: %w", err)
	}
	log.Warn().Msg("encryption enabled")
	return nil
}

func (svc *aesEncryptionService) disable() error {
	err := svc.callbackOnDisable()
	if err != nil {
		return fmt.Errorf("failed disabling AES encryption: %w", err)
	}
	err = svc.deleteCiphertextSample()
	if err != nil {
		return fmt.Errorf("failed disabling AES encryption: %w", err)
	}
	log.Warn().Msg("encryption disabled")
	return nil
}

func (svc *aesEncryptionService) updateCiphertextSample() error {
	ciphertext, err := svc.Encrypt(svc.keyID, keyIDAssociatedData)
	if err != nil {
		return fmt.Errorf("failed updating server encryption configuration: %w", err)
	}
	err = svc.store.ServerConfigSet(ciphertextSampleConfigKey, ciphertext)
	if err != nil {
		return fmt.Errorf("failed updating server encryption configuration: %w", err)
	}
	log.Info().Msg("registered new encryption key")
	return nil
}

func (svc *aesEncryptionService) deleteCiphertextSample() error {
	err := svc.store.ServerConfigDelete(ciphertextSampleConfigKey)
	if err != nil {
		err = fmt.Errorf("failed updating server encryption configuration: %w", err)
	}
	return err
}

func (svc *aesEncryptionService) callbackOnEnable() error {
	for _, client := range svc.clients {
		err := client.EnableEncryption()
		if err != nil {
			return fmt.Errorf("failed enabling AES encryption: %w", err)
		}
	}
	log.Info().Msg("enabled encryption on registered services")
	return nil
}

func (svc *aesEncryptionService) callbackOnDisable() error {
	for _, client := range svc.clients {
		err := client.MigrateEncryption(&noEncryption{})
		if err != nil {
			return fmt.Errorf("failed disabling AES encryption: %w", err)
		}
	}
	log.Info().Msg("disabled encryption on registered services")
	return nil
}
