// Copyright 2026 Woodpecker Authors
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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
)

// EncryptedRegistryStore is a model.RegistryStore decorator that encrypts
// registry passwords before they are stored and decrypts them after loading.
type EncryptedRegistryStore struct {
	store      model.RegistryStore
	encryption types.EncryptionService
}

// Ensure wrapper match interface.
var _ model.RegistryStore = new(EncryptedRegistryStore)

func NewRegistryStore(registryStore model.RegistryStore) *EncryptedRegistryStore {
	return &EncryptedRegistryStore{store: registryStore}
}

func (wrapper *EncryptedRegistryStore) SetEncryptionService(service types.EncryptionService) error {
	if wrapper.encryption != nil {
		return errors.New(errMessageInitSeveralTimes)
	}
	wrapper.encryption = service
	return nil
}

func (wrapper *EncryptedRegistryStore) EnableEncryption() error {
	log.Warn().Msg(logMessageEnablingRegistriesEncryption)
	registries, err := wrapper.store.RegistryListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEnableRegistries, err)
	}
	for _, registry := range registries {
		if isMarkedValue(registry.Password) {
			// already encrypted by an earlier, interrupted run
			continue
		}
		if err := wrapper.encrypt(registry); err != nil {
			return err
		}
		if err := wrapper.save(registry); err != nil {
			return err
		}
	}
	log.Warn().Msg(logMessageEnablingRegistriesEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedRegistryStore) MigrateEncryption(newEncryptionService types.EncryptionService) error {
	log.Warn().Msg(logMessageMigratingRegistriesEncryption)
	registries, err := wrapper.store.RegistryListAll()
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToMigrateRegistries, err)
	}
	if err := wrapper.decryptList(registries); err != nil {
		return err
	}
	for _, registry := range registries {
		if err := wrapper.encryptWith(newEncryptionService, registry); err != nil {
			return err
		}
		if err := wrapper.save(registry); err != nil {
			return err
		}
	}
	wrapper.encryption = newEncryptionService
	log.Warn().Msg(logMessageMigratingRegistriesEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedRegistryStore) encrypt(registry *model.Registry) error {
	return wrapper.encryptWith(wrapper.encryption, registry)
}

func (wrapper *EncryptedRegistryStore) encryptWith(encryption types.EncryptionService, registry *model.Registry) error {
	encryptedPassword, err := encryption.Encrypt(registry.Password, strconv.Itoa(int(registry.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEncryptRegistry, registry.ID, err)
	}
	registry.Password = encryptedPassword
	return nil
}

func (wrapper *EncryptedRegistryStore) decrypt(registry *model.Registry) error {
	decryptedPassword, err := wrapper.encryption.Decrypt(registry.Password, strconv.Itoa(int(registry.ID)))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToDecryptRegistry, registry.ID, err)
	}
	registry.Password = decryptedPassword
	return nil
}

func (wrapper *EncryptedRegistryStore) decryptList(registries []*model.Registry) error {
	for _, registry := range registries {
		if err := wrapper.decrypt(registry); err != nil {
			return err
		}
	}
	return nil
}

func (wrapper *EncryptedRegistryStore) save(registry *model.Registry) error {
	if err := wrapper.store.RegistryUpdate(registry); err != nil {
		log.Err(err).Msg(errMessageTemplateRegistryStorageError)
		return err
	}
	return nil
}

func (wrapper *EncryptedRegistryStore) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	result, err := wrapper.store.RegistryFind(repo, addr)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedRegistryStore) RegistryList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptions) ([]*model.Registry, error) {
	results, err := wrapper.store.RegistryList(repo, includeGlobalAndOrg, p)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedRegistryStore) RegistryListAll() ([]*model.Registry, error) {
	results, err := wrapper.store.RegistryListAll()
	if err != nil {
		return nil, err
	}
	if err := wrapper.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedRegistryStore) RegistryCreate(registry *model.Registry) error {
	newRegistry := &model.Registry{}
	if err := wrapper.store.RegistryCreate(newRegistry); err != nil {
		return err
	}
	registry.ID = newRegistry.ID

	plainPassword := registry.Password
	defer func() { registry.Password = plainPassword }()

	if err := wrapper.encrypt(registry); err != nil {
		if deleteErr := wrapper.store.RegistryDelete(newRegistry); deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackRegistryCreation, err, deleteErr.Error())
		}
		return err
	}

	if err := wrapper.store.RegistryUpdate(registry); err != nil {
		if deleteErr := wrapper.store.RegistryDelete(newRegistry); deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackRegistryCreation, err, deleteErr.Error())
		}
		return err
	}

	return nil
}

func (wrapper *EncryptedRegistryStore) RegistryUpdate(registry *model.Registry) error {
	plainPassword := registry.Password
	defer func() { registry.Password = plainPassword }()

	if err := wrapper.encrypt(registry); err != nil {
		return err
	}

	return wrapper.store.RegistryUpdate(registry)
}

func (wrapper *EncryptedRegistryStore) RegistryDelete(registry *model.Registry) error {
	return wrapper.store.RegistryDelete(registry)
}

func (wrapper *EncryptedRegistryStore) OrgRegistryFind(orgID int64, addr string) (*model.Registry, error) {
	result, err := wrapper.store.OrgRegistryFind(orgID, addr)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedRegistryStore) OrgRegistryList(orgID int64, p *model.ListOptions) ([]*model.Registry, error) {
	results, err := wrapper.store.OrgRegistryList(orgID, p)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedRegistryStore) GlobalRegistryFind(addr string) (*model.Registry, error) {
	result, err := wrapper.store.GlobalRegistryFind(addr)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedRegistryStore) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	results, err := wrapper.store.GlobalRegistryList(p)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}
