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
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (e *EncryptedStore) SecretFind(repo *model.Repo, s string) (*model.Secret, error) {
	result, err := e.Store.SecretFind(repo, s)
	if err != nil {
		return nil, err
	}

	if err := e.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) SecretList(repo *model.Repo, b bool) ([]*model.Secret, error) {
	results, err := e.Store.SecretList(repo, b)
	if err != nil {
		return nil, err
	}

	if err := e.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (e *EncryptedStore) SecretCreate(secret *model.Secret) error {
	// make sure a new ID is created
	secret.ID = 0

	if err := e.encrypt(secret); err != nil {
		return err
	}

	if err := e.Store.SecretCreate(secret); err != nil {
		return err
	}

	return e.decrypt(secret)
}

func (e *EncryptedStore) SecretUpdate(secret *model.Secret) error {
	if err := e.encrypt(secret); err != nil {
		return err
	}

	if err := e.Store.SecretUpdate(secret); err != nil {
		return err
	}

	return e.decrypt(secret)
}

func (e *EncryptedStore) OrgSecretFind(s, s2 string) (*model.Secret, error) {
	result, err := e.Store.OrgSecretFind(s, s2)
	if err != nil {
		return nil, err
	}

	if err := e.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) OrgSecretList(s string) ([]*model.Secret, error) {
	results, err := e.Store.OrgSecretList(s)
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
	result, err := e.Store.GlobalSecretFind(s)
	if err != nil {
		return nil, err
	}

	if err := e.decrypt(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EncryptedStore) GlobalSecretList() ([]*model.Secret, error) {
	results, err := e.Store.GlobalSecretList()
	if err != nil {
		return nil, err
	}

	if err := e.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}

func (e *EncryptedStore) SecretListAll() ([]*model.Secret, error) {
	results, err := e.Store.SecretListAll()
	if err != nil {
		return nil, err
	}

	if err := e.decryptList(results); err != nil {
		return nil, err
	}
	return results, nil
}
