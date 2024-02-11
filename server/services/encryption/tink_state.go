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

package encryption

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (svc *tinkEncryptionService) enable() error {
	if err := svc.callbackOnEnable(); err != nil {
		return fmt.Errorf(errTemplateFailedEnablingEncryption, err)
	}

	if err := svc.updateCiphertextSample(); err != nil {
		return fmt.Errorf(errTemplateFailedEnablingEncryption, err)
	}

	log.Warn().Msg(logMessageEncryptionEnabled)
	return nil
}

func (svc *tinkEncryptionService) disable() error {
	if err := svc.callbackOnDisable(); err != nil {
		return fmt.Errorf(errTemplateFailedDisablingEncryption, err)
	}

	if err := svc.deleteCiphertextSample(); err != nil {
		return fmt.Errorf(errTemplateFailedDisablingEncryption, err)
	}

	log.Warn().Msg(logMessageEncryptionDisabled)
	return nil
}

func (svc *tinkEncryptionService) rotate() error {
	newSvc := &tinkEncryptionService{
		keysetFilePath:    svc.keysetFilePath,
		primaryKeyID:      "",
		encryption:        nil,
		store:             svc.store,
		keysetFileWatcher: nil,
		clients:           svc.clients,
	}

	if err := newSvc.loadKeyset(); err != nil {
		return fmt.Errorf(errTemplateFailedRotatingEncryption, err)
	}

	err := newSvc.validateKeyset()
	if errors.Is(err, errEncryptionKeyRotated) {
		err = newSvc.updateCiphertextSample()
	}
	if err != nil {
		return fmt.Errorf(errTemplateFailedRotatingEncryption, err)
	}

	if err := newSvc.callbackOnRotation(); err != nil {
		return fmt.Errorf(errTemplateFailedRotatingEncryption, err)
	}

	if err := newSvc.initFileWatcher(); err != nil {
		return fmt.Errorf(errTemplateFailedRotatingEncryption, err)
	}
	return nil
}

func (svc *tinkEncryptionService) updateCiphertextSample() error {
	ciphertext, err := svc.Encrypt(svc.primaryKeyID, keyIDAssociatedData)
	if err != nil {
		return fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}

	if err := svc.store.ServerConfigSet(ciphertextSampleConfigKey, ciphertext); err != nil {
		return fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}

	log.Info().Msg(logMessageEncryptionKeyRegistered)
	return nil
}

func (svc *tinkEncryptionService) deleteCiphertextSample() error {
	if err := svc.store.ServerConfigDelete(ciphertextSampleConfigKey); err != nil {
		return fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}
	return nil
}

func (svc *tinkEncryptionService) initClients() error {
	for _, client := range svc.clients {
		if err := client.SetEncryptionService(svc); err != nil {
			return err
		}
	}
	log.Info().Msg(logMessageClientsInitialized)
	return nil
}

func (svc *tinkEncryptionService) callbackOnEnable() error {
	for _, client := range svc.clients {
		if err := client.EnableEncryption(); err != nil {
			return err
		}
	}
	log.Info().Msg(logMessageClientsEnabled)
	return nil
}

func (svc *tinkEncryptionService) callbackOnRotation() error {
	for _, client := range svc.clients {
		if err := client.MigrateEncryption(svc); err != nil {
			return err
		}
	}
	log.Info().Msg(logMessageClientsRotated)
	return nil
}

func (svc *tinkEncryptionService) callbackOnDisable() error {
	for _, client := range svc.clients {
		if err := client.MigrateEncryption(&noEncryption{}); err != nil {
			return err
		}
	}
	log.Info().Msg(logMessageClientsDecrypted)
	return nil
}
