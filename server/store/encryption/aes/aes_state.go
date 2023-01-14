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

package aes

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func (svc *aesEncryptionService) initClients() error {
	for _, client := range svc.clients {
		err := client.SetEncryptionService(svc)
		if err != nil {
			return fmt.Errorf(errTemplateFailedInitializingClients, err)
		}
	}
	log.Info().Msg(logMessageClientsInitialized)
	return nil
}

func (svc *aesEncryptionService) enable() error {
	err := svc.callbackOnEnable()
	if err != nil {
		return fmt.Errorf(errTemplateFailedEnablingEncryption, err)
	}
	err = svc.updateCiphertextSample()
	if err != nil {
		return fmt.Errorf(errTemplateFailedEnablingEncryption, err)
	}
	log.Warn().Msg(logMessageEncryptionEnabled)
	return nil
}

func (svc *aesEncryptionService) disable() error {
	err := svc.callbackOnDisable()
	if err != nil {
		return fmt.Errorf(errTemplateFailedDisablingEncryption, err)
	}
	err = svc.deleteCiphertextSample()
	if err != nil {
		return fmt.Errorf(errTemplateFailedDisablingEncryption, err)
	}
	log.Warn().Msg(logMessageEncryptionDisabled)
	return nil
}

func (svc *aesEncryptionService) updateCiphertextSample() error {
	ciphertext, err := svc.Encrypt(svc.keyID, keyIDAssociatedData)
	if err != nil {
		return fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}
	err = svc.store.ServerConfigSet(ciphertextSampleConfigKey, ciphertext)
	if err != nil {
		return fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}
	log.Info().Msg(logMessageEncryptionKeyRegistered)
	return nil
}

func (svc *aesEncryptionService) deleteCiphertextSample() error {
	err := svc.store.ServerConfigDelete(ciphertextSampleConfigKey)
	if err != nil {
		err = fmt.Errorf(errTemplateFailedUpdatingServerConfig, err)
	}
	return err
}

func (svc *aesEncryptionService) callbackOnEnable() error {
	for _, client := range svc.clients {
		err := client.EnableEncryption()
		if err != nil {
			return fmt.Errorf(errTemplateFailedEnablingEncryption, err)
		}
	}
	log.Info().Msg(logMessageClientsEnabled)
	return nil
}

func (svc *aesEncryptionService) callbackOnDisable() error {
	for _, client := range svc.clients {
		err := client.MigrateEncryption(&noEncryption{})
		if err != nil {
			return fmt.Errorf(errTemplateFailedDisablingEncryption, err)
		}
	}
	log.Info().Msg(logMessageEncryptionDisabled)
	return nil
}
