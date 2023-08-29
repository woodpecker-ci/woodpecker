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
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/encryption"
)

const (
	secretValueTemplate = "_%s_"
)

type EncryptedSecretService struct {
	secretSvc     model.SecretService
	encryptionSvc encryption.EncryptionService
}

func NewEncrypted(secretService model.SecretService, encryptionService encryption.EncryptionService) EncryptedSecretService {
	return EncryptedSecretService{
		secretSvc:     secretService,
		encryptionSvc: encryptionService,
	}
}

func (ess *EncryptedSecretService) EncryptSecret(secret *model.Secret) error {
	log.Debug().Int64("id", secret.ID).Str("name", secret.Name).Msg("encryption")

	encryptedValue, err := ess.encryptionSvc.Encrypt(secret.Value, secret.Name)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret id=%d: %w", secret.ID, err)
	}
	encodedValue := ess.encodeSecretValue(encryptedValue)
	secret.Value = encodedValue
	return nil
}

func (ess *EncryptedSecretService) encodeSecretValue(value string) string {
	return ess.header() + value
}

func (ess *EncryptedSecretService) DecryptSecret(secret *model.Secret) error {
	log.Debug().Int64("id", secret.ID).Str("name", secret.Name).Msg("decryption")

	decodedValue := ess.decodeSecretValue(secret.Value)
	decryptedValue, err := ess.encryptionSvc.Decrypt(decodedValue, secret.Name)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret id=%d: %w", secret.ID, err)
	}
	secret.Value = decryptedValue
	return nil
}

func (ess *EncryptedSecretService) decodeSecretValue(value string) string {
	return strings.TrimPrefix(value, ess.header())
}

func (ess *EncryptedSecretService) decryptList(secrets []*model.Secret) error {
	for _, secret := range secrets {
		err := ess.DecryptSecret(secret)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ess *EncryptedSecretService) header() string {
	return fmt.Sprintf(secretValueTemplate, ess.encryptionSvc.Algo())
}

// SecretService interface

func (ess *EncryptedSecretService) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.SecretFind(repo, name)
	if err != nil {
		return nil, err
	}
	err = ess.DecryptSecret(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *EncryptedSecretService) SecretList(repo *model.Repo, listOpt *model.ListOptions) ([]*model.Secret, error) {
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

func (ess *EncryptedSecretService) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, listOpt *model.ListOptions) ([]*model.Secret, error) {
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

func (ess *EncryptedSecretService) SecretCreate(repo *model.Repo, in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.SecretCreate(repo, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) SecretUpdate(repo *model.Repo, in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.SecretUpdate(repo, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) SecretDelete(repo *model.Repo, name string) error {
	return ess.secretSvc.SecretDelete(repo, name)
}

func (ess *EncryptedSecretService) OrgSecretFind(owner int64, name string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.OrgSecretFind(owner, name)
	if err != nil {
		return nil, err
	}
	err = ess.DecryptSecret(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *EncryptedSecretService) OrgSecretList(owner int64, listOpt *model.ListOptions) ([]*model.Secret, error) {
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

func (ess *EncryptedSecretService) OrgSecretCreate(owner int64, in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.OrgSecretCreate(owner, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) OrgSecretUpdate(owner int64, in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.OrgSecretUpdate(owner, in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) OrgSecretDelete(owner int64, name string) error {
	return ess.secretSvc.OrgSecretDelete(owner, name)
}

func (ess *EncryptedSecretService) GlobalSecretFind(owner string) (*model.Secret, error) {
	var err error
	secret, err := ess.secretSvc.GlobalSecretFind(owner)
	if err != nil {
		return nil, err
	}
	err = ess.DecryptSecret(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (ess *EncryptedSecretService) GlobalSecretList(listOpt *model.ListOptions) ([]*model.Secret, error) {
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

func (ess *EncryptedSecretService) GlobalSecretCreate(in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.GlobalSecretCreate(in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) GlobalSecretUpdate(in *model.Secret) error {
	var err error
	err = ess.EncryptSecret(in)
	if err != nil {
		return err
	}
	err = ess.secretSvc.GlobalSecretUpdate(in)
	if err != nil {
		return err
	}
	return nil
}

func (ess *EncryptedSecretService) GlobalSecretDelete(name string) error {
	return ess.secretSvc.GlobalSecretDelete(name)
}
