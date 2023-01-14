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
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (e *EncryptedStore) SecretFind(repo *model.Repo, s string) (*model.Secret, error) {
	result, err := e.store.SecretFind(repo, s)
	if err != nil {
		return nil, err
	}
	err = e.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) SecretList(repo *model.Repo, b bool) ([]*model.Secret, error) {
	results, err := e.store.SecretList(repo, b)
	if err != nil {
		return nil, err
	}
	err = e.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *EncryptedStore) SecretCreate(secret *model.Secret) error {
	newSecret := &model.Secret{}
	err := e.store.SecretCreate(newSecret)
	if err != nil {
		return err
	}
	secret.ID = newSecret.ID

	err = e.encrypt(secret)
	if err != nil {
		deleteErr := e.SecretDelete(newSecret)
		if deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackSecretCreation, err, deleteErr.Error())
		}
		return err
	}

	err = e.SecretUpdate(secret)
	if err != nil {
		deleteErr := e.SecretDelete(newSecret)
		if deleteErr != nil {
			return fmt.Errorf(errMessageTemplateFailedToRollbackSecretCreation, err, deleteErr.Error())
		}
		return err
	}

	err = e.decrypt(secret)
	if err != nil {
		return err
	}
	return nil
}

func (e *EncryptedStore) SecretUpdate(secret *model.Secret) error {
	err := e.encrypt(secret)
	if err != nil {
		return err
	}

	err = e.store.SecretUpdate(secret)
	if err != nil {
		return err
	}

	err = e.decrypt(secret)
	if err != nil {
		return err
	}
	return nil
}

func (e *EncryptedStore) OrgSecretFind(s, s2 string) (*model.Secret, error) {
	result, err := e.store.OrgSecretFind(s, s2)
	if err != nil {
		return nil, err
	}

	err = e.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) OrgSecretList(s string) ([]*model.Secret, error) {
	results, err := e.store.OrgSecretList(s)
	if err != nil {
		return nil, err
	}

	err = e.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *EncryptedStore) GlobalSecretFind(s string) (*model.Secret, error) {
	result, err := e.store.GlobalSecretFind(s)
	if err != nil {
		return nil, err
	}

	err = e.decrypt(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) GlobalSecretList() ([]*model.Secret, error) {
	results, err := e.store.GlobalSecretList()
	if err != nil {
		return nil, err
	}

	err = e.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *EncryptedStore) SecretListAll() ([]*model.Secret, error) {
	results, err := e.store.SecretListAll()
	if err != nil {
		return nil, err
	}

	err = e.decryptList(results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
