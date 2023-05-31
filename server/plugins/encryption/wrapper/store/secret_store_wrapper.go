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

package store

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type EncryptedSecretStore struct {
	store      model.SecretStore
	encryption model.EncryptionService
}

// ensure wrapper match interface
var _ model.SecretStore = new(EncryptedSecretStore)

func NewSecretStore(secretStore model.SecretStore) *EncryptedSecretStore {
	wrapper := EncryptedSecretStore{secretStore, nil}
	return &wrapper
}

func (wrapper *EncryptedSecretStore) SetEncryptionService(service model.EncryptionService) error {
	if wrapper.encryption != nil {
		return errors.New(errMessageInitSeveralTimes)
	}
	wrapper.encryption = service
	return nil
}

func (wrapper *EncryptedSecretStore) EnableEncryption() error {
	log.Warn().Msg(logMessageEnablingSecretsEncryption)
	secrets, err := wrapper.store.SecretListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEnable, err)
	}
	for _, secret := range secrets {
		if err := wrapper.encrypt(secret); err != nil {
			return err
		}
		if err := wrapper._save(secret); err != nil {
			return err
		}
	}
	log.Warn().Msg(logMessageEnablingSecretsEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedSecretStore) MigrateEncryption(newEncryptionService model.EncryptionService) error {
	log.Warn().Msg(logMessageMigratingSecretsEncryption)
	secrets, err := wrapper.store.SecretListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToMigrate, err)
	}
	if err := wrapper.decryptList(secrets); err != nil {
		return err
	}
	wrapper.encryption = newEncryptionService
	for _, secret := range secrets {
		if err := wrapper.encrypt(secret); err != nil {
			return err
		}
		if err := wrapper._save(secret); err != nil {
			return err
		}
	}
	log.Warn().Msg(logMessageMigratingSecretsEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedSecretStore) encrypt(secret *model.Secret) error {
	encryptedValue, err := wrapper.encryption.Encrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEncryptSecret, secret.ID, err)
	}
	secret.Value = encryptedValue
	return nil
}

func (wrapper *EncryptedSecretStore) decrypt(secret *model.Secret) error {
	decryptedValue, err := wrapper.encryption.Decrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToDecryptSecret, secret.ID, err)
	}
	secret.Value = decryptedValue
	return nil
}

func (wrapper *EncryptedSecretStore) decryptList(secrets []*model.Secret) error {
	for _, secret := range secrets {
		err := wrapper.decrypt(secret)
		if err != nil {
			return fmt.Errorf(errMessageTemplateFailedToDecryptSecret, secret.ID, err)
		}
	}
	return nil
}

func (wrapper *EncryptedSecretStore) _save(secret *model.Secret) error {
	err := wrapper.store.SecretUpdate(secret)
	if err != nil {
		log.Err(err).Msg(errMessageTemplateStorageError)
		return err
	}
	return nil
}
