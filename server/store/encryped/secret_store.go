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

package encrypted

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type EncryptedStore struct {
	store.Store
	store      store.Store
	encryption model.EncryptionService
}

// ensure wrapper match interface
var _ store.Store = new(EncryptedStore)

func NewEncryptedStore(store store.Store) *EncryptedStore {
	return &EncryptedStore{store, store, nil}
}

func (s *EncryptedStore) SetEncryptionService(service model.EncryptionService) error {
	if s.encryption != nil {
		return errors.New(errMessageInitSeveralTimes)
	}
	s.encryption = service
	return nil
}

func (e *EncryptedStore) EnableEncryption() error {
	log.Warn().Msg(logMessageEnablingSecretsEncryption)
	secrets, err := e.SecretListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEnable, err)
	}
	for _, secret := range secrets {
		if err := e.encrypt(secret); err != nil {
			return err
		}
		if err := e._save(secret); err != nil {
			return err
		}
	}
	log.Warn().Msg(logMessageEnablingSecretsEncryptionSuccess)
	return nil
}

func (e *EncryptedStore) MigrateEncryption(newEncryptionService model.EncryptionService) error {
	log.Warn().Msg(logMessageMigratingSecretsEncryption)
	secrets, err := e.SecretListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToMigrate, err)
	}
	if err := e.decryptList(secrets); err != nil {
		return err
	}
	e.encryption = newEncryptionService
	for _, secret := range secrets {
		if err := e.encrypt(secret); err != nil {
			return err
		}
		if err := e._save(secret); err != nil {
			return err
		}
	}
	log.Warn().Msg(logMessageMigratingSecretsEncryptionSuccess)
	return nil
}

func (e *EncryptedStore) encrypt(secret *model.Secret) error {
	encryptedValue, err := e.encryption.Encrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEncryptSecret, secret.ID, err)
	}
	secret.Value = encryptedValue
	return nil
}

func (e *EncryptedStore) decrypt(secret *model.Secret) error {
	decryptedValue, err := e.encryption.Decrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToDecryptSecret, secret.ID, err)
	}
	secret.Value = decryptedValue
	return nil
}

func (e *EncryptedStore) decryptList(secrets []*model.Secret) error {
	for _, secret := range secrets {
		err := e.decrypt(secret)
		if err != nil {
			return fmt.Errorf(errMessageTemplateFailedToDecryptSecret, secret.ID, err)
		}
	}
	return nil
}

func (e *EncryptedStore) _save(secret *model.Secret) error {
	err := e.SecretUpdate(secret)
	if err != nil {
		log.Err(err).Msg(errMessageTemplateStorageError)
		return err
	}
	return nil
}
