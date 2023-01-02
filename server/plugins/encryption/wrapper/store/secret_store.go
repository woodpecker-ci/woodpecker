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

package store

import (
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (wrapper *EncryptedSecretStore) SecretFind(repo *model.Repo, s string) (*model.Secret, error) {
	result, err := wrapper.store.SecretFind(repo, s)
	if err != nil {
		return nil, err
	}
	err = wrapper.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedSecretStore) SecretList(repo *model.Repo, b bool) ([]*model.Secret, error) {
	results, err := wrapper.store.SecretList(repo, b)
	if err != nil {
		return nil, err
	}
	err = wrapper.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedSecretStore) SecretCreate(secret *model.Secret) error {
	newSecret := &model.Secret{}
	err := wrapper.store.SecretCreate(newSecret)
	if err != nil {
		return err
	}
	secret.ID = newSecret.ID

	err = wrapper.encrypt(secret)
	if err != nil {
		deleteErr := wrapper.store.SecretDelete(newSecret)
		if deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackSecretCreation, err, deleteErr.Error())
		}
		return err
	}

	err = wrapper.store.SecretUpdate(secret)
	if err != nil {
		deleteErr := wrapper.store.SecretDelete(newSecret)
		if deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackSecretCreation, err, deleteErr.Error())
		}
		return err
	}

	err = wrapper.decrypt(secret)
	if err != nil {
		return err
	}
	return nil
}

func (wrapper *EncryptedSecretStore) SecretUpdate(secret *model.Secret) error {
	err := wrapper.encrypt(secret)
	if err != nil {
		return err
	}

	err = wrapper.store.SecretUpdate(secret)
	if err != nil {
		return err
	}

	err = wrapper.decrypt(secret)
	if err != nil {
		return err
	}
	return nil
}

func (wrapper *EncryptedSecretStore) SecretDelete(secret *model.Secret) error {
	return wrapper.store.SecretDelete(secret)
}

func (wrapper *EncryptedSecretStore) OrgSecretFind(s, s2 string) (*model.Secret, error) {
	result, err := wrapper.store.OrgSecretFind(s, s2)
	if err != nil {
		return nil, err
	}

	err = wrapper.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedSecretStore) OrgSecretList(s string) ([]*model.Secret, error) {
	results, err := wrapper.store.OrgSecretList(s)
	if err != nil {
		return nil, err
	}

	err = wrapper.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedSecretStore) GlobalSecretFind(s string) (*model.Secret, error) {
	result, err := wrapper.store.GlobalSecretFind(s)
	if err != nil {
		return nil, err
	}

	err = wrapper.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (wrapper *EncryptedSecretStore) GlobalSecretList() ([]*model.Secret, error) {
	results, err := wrapper.store.GlobalSecretList()
	if err != nil {
		return nil, err
	}

	err = wrapper.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (wrapper *EncryptedSecretStore) SecretListAll() ([]*model.Secret, error) {
	results, err := wrapper.store.SecretListAll()
	if err != nil {
		return nil, err
	}

	err = wrapper.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
