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

package secrets

import (
	"fmt"
	"strconv"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/encryption"
)

type encryptedSecretService struct {
	secretSvc     model.SecretService
	encryptionSvc encryption.EncryptionService
}

func NewEncrypted(secretService *model.SecretService, encryptionService *encryption.EncryptionService) model.SecretService {
	return &encryptedSecretService{
		secretSvc:     *secretService,
		encryptionSvc: *encryptionService,
	}
}

func (ess *encryptedSecretService) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.SecretFind(repo, name)
	if err != nil {
		return nil, err
	}
	err = ess.decrypt(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *encryptedSecretService) SecretList(repo *model.Repo, listOpt *model.ListOptions) ([]*model.Secret, error) {
	var err error
	secrets, err := ess.secretSvc.SecretList(repo, listOpt)
	if err != nil {
		return nil, err
	}
	err = ess.decryptList(secrets)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (ess *encryptedSecretService) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, listOpt *model.ListOptions) ([]*model.Secret, error) {
	var err error
	secrets, err := ess.secretSvc.SecretListPipeline(repo, pipeline, listOpt)
	if err != nil {
		return nil, err
	}
	err = ess.decryptList(secrets)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (ess *encryptedSecretService) SecretCreate(repo *model.Repo, in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.SecretCreate(repo, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) SecretUpdate(repo *model.Repo, in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.SecretUpdate(repo, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) SecretDelete(repo *model.Repo, name string) error {
	return ess.secretSvc.SecretDelete(repo, name)
}

func (ess *encryptedSecretService) OrgSecretFind(owner int64, name string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.OrgSecretFind(owner, name)
	if err != nil {
		return nil, err
	}
	err = ess.decrypt(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *encryptedSecretService) OrgSecretList(owner int64, listOpt *model.ListOptions) ([]*model.Secret, error) {
	var err error
	secrets, err := ess.secretSvc.OrgSecretList(owner, listOpt)
	if err != nil {
		return nil, err
	}
	err = ess.decryptList(secrets)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (ess *encryptedSecretService) OrgSecretCreate(owner int64, in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.OrgSecretCreate(owner, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) OrgSecretUpdate(owner int64, in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.OrgSecretUpdate(owner, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) OrgSecretDelete(owner int64, name string) error {
	return ess.secretSvc.OrgSecretDelete(owner, name)
}

func (ess *encryptedSecretService) GlobalSecretFind(owner string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.GlobalSecretFind(owner)
	if err != nil {
		return nil, err
	}
	err = ess.decrypt(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *encryptedSecretService) GlobalSecretList(listOpt *model.ListOptions) ([]*model.Secret, error) {
	var err error
	secrets, err := ess.secretSvc.GlobalSecretList(listOpt)
	if err != nil {
		return nil, err
	}
	err = ess.decryptList(secrets)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (ess *encryptedSecretService) GlobalSecretCreate(in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.GlobalSecretCreate(in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) GlobalSecretUpdate(in *model.Secret) error {
	var err error
	err = ess.encrypt(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.GlobalSecretUpdate(in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *encryptedSecretService) GlobalSecretDelete(name string) error {
	return ess.secretSvc.GlobalSecretDelete(name)
}

func (ess *encryptedSecretService) encrypt(secret *model.Secret) error {
	encryptedValue, err := ess.encryptionSvc.Encrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf("failed to encrypt secret id=%d: %w", secret.ID, err)
	}
	secret.Value = encryptedValue
	return nil
}

func (ess *encryptedSecretService) decrypt(secret *model.Secret) error {
	decryptedValue, err := ess.encryptionSvc.Decrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	if err != nil {
		return fmt.Errorf("failed to decrypt secret id=%d: %w", secret.ID, err)
	}
	secret.Value = decryptedValue
	return nil
}

func (ess *encryptedSecretService) decryptList(secrets []*model.Secret) error {
	for _, secret := range secrets {
		err := ess.decrypt(secret)
		if err != nil {
			return err
		}
	}
	return nil
}
