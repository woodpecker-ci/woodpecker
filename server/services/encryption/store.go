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

package encryption

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/tink-crypto/tink-go/v2/tink"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// encryptedStore is a store.Store that encrypts secret values, registry
// passwords and user forge tokens on write and decrypts them on read. Without
// a key it stores plaintext and refuses to read encrypted values.
type encryptedStore struct {
	store.Store
	key tink.AEAD
}

// The associated data binds a ciphertext to its record, so it cannot be
// copied to another one.
func secretAD(id int64) string {
	return "secret/" + strconv.FormatInt(id, 10)
}

func registryAD(id int64) string {
	return "registry/" + strconv.FormatInt(id, 10)
}

func accessTokenAD(id int64) string {
	return "user/" + strconv.FormatInt(id, 10) + "/access-token"
}

func refreshTokenAD(id int64) string {
	return "user/" + strconv.FormatInt(id, 10) + "/refresh-token"
}

// rewrite stores every secret value, registry password and user token as
// convert returns it. Records convert leaves unchanged are not written.
func (s *encryptedStore) rewrite(convert func(value, associatedData string) (string, error)) error {
	secrets, err := s.Store.SecretListAll()
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		value, err := convert(secret.Value, secretAD(secret.ID))
		if err != nil {
			return fmt.Errorf("secret %d: %w", secret.ID, err)
		}
		if value != secret.Value {
			secret.Value = value
			if err := s.Store.SecretUpdate(secret); err != nil {
				return err
			}
		}
	}

	registries, err := s.Store.RegistryListAll()
	if err != nil {
		return err
	}
	for _, registry := range registries {
		password, err := convert(registry.Password, registryAD(registry.ID))
		if err != nil {
			return fmt.Errorf("registry %d: %w", registry.ID, err)
		}
		if password != registry.Password {
			registry.Password = password
			if err := s.Store.RegistryUpdate(registry); err != nil {
				return err
			}
		}
	}

	users, err := s.Store.GetUserList(&model.ListOptionsWithAll{All: true})
	if err != nil {
		return err
	}
	for _, user := range users {
		accessToken, err := convert(user.AccessToken, accessTokenAD(user.ID))
		if err != nil {
			return fmt.Errorf("user %d: %w", user.ID, err)
		}
		refreshToken, err := convert(user.RefreshToken, refreshTokenAD(user.ID))
		if err != nil {
			return fmt.Errorf("user %d: %w", user.ID, err)
		}
		if accessToken != user.AccessToken || refreshToken != user.RefreshToken {
			user.AccessToken, user.RefreshToken = accessToken, refreshToken
			if err := s.Store.UpdateUser(user); err != nil {
				return err
			}
		}
	}
	return nil
}

// Secrets.

func (s *encryptedStore) secret(secret *model.Secret, err error) (*model.Secret, error) {
	if err != nil {
		return nil, err
	}
	if secret.Value, err = decrypt(s.key, secret.Value, secretAD(secret.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt secret %d: %w", secret.ID, err)
	}
	return secret, nil
}

func (s *encryptedStore) secrets(secrets []*model.Secret, err error) ([]*model.Secret, error) {
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets {
		if _, err := s.secret(secret, nil); err != nil {
			return nil, err
		}
	}
	return secrets, nil
}

func (s *encryptedStore) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	return s.secret(s.Store.SecretFind(repo, name))
}

func (s *encryptedStore) SecretList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptionsWithAll) ([]*model.Secret, error) {
	return s.secrets(s.Store.SecretList(repo, includeGlobalAndOrg, p))
}

func (s *encryptedStore) SecretListAll() ([]*model.Secret, error) {
	return s.secrets(s.Store.SecretListAll())
}

func (s *encryptedStore) OrgSecretFind(orgID int64, name string) (*model.Secret, error) {
	return s.secret(s.Store.OrgSecretFind(orgID, name))
}

func (s *encryptedStore) OrgSecretList(orgID int64, p *model.ListOptionsWithAll) ([]*model.Secret, error) {
	return s.secrets(s.Store.OrgSecretList(orgID, p))
}

func (s *encryptedStore) GlobalSecretFind(name string) (*model.Secret, error) {
	return s.secret(s.Store.GlobalSecretFind(name))
}

func (s *encryptedStore) GlobalSecretList(p *model.ListOptionsWithAll) ([]*model.Secret, error) {
	return s.secrets(s.Store.GlobalSecretList(p))
}

// SecretCreate stores the secret without its value first, because the id the
// value is bound to only exists afterwards.
func (s *encryptedStore) SecretCreate(secret *model.Secret) error {
	if s.key == nil {
		return s.Store.SecretCreate(secret)
	}
	value := secret.Value
	secret.Value = ""
	err := s.Store.SecretCreate(secret)
	secret.Value = value
	if err != nil {
		return err
	}
	if err := s.SecretUpdate(secret); err != nil {
		return errors.Join(err, s.SecretDelete(secret))
	}
	return nil
}

func (s *encryptedStore) SecretUpdate(secret *model.Secret) error {
	encrypted := *secret
	var err error
	if encrypted.Value, err = encrypt(s.key, secret.Value, secretAD(secret.ID)); err != nil {
		return err
	}
	return s.Store.SecretUpdate(&encrypted)
}

// Registries.

func (s *encryptedStore) registry(registry *model.Registry, err error) (*model.Registry, error) {
	if err != nil {
		return nil, err
	}
	if registry.Password, err = decrypt(s.key, registry.Password, registryAD(registry.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt registry %d: %w", registry.ID, err)
	}
	return registry, nil
}

func (s *encryptedStore) registries(registries []*model.Registry, err error) ([]*model.Registry, error) {
	if err != nil {
		return nil, err
	}
	for _, registry := range registries {
		if _, err := s.registry(registry, nil); err != nil {
			return nil, err
		}
	}
	return registries, nil
}

func (s *encryptedStore) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	return s.registry(s.Store.RegistryFind(repo, addr))
}

func (s *encryptedStore) RegistryList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptionsWithAll) ([]*model.Registry, error) {
	return s.registries(s.Store.RegistryList(repo, includeGlobalAndOrg, p))
}

func (s *encryptedStore) RegistryListAll() ([]*model.Registry, error) {
	return s.registries(s.Store.RegistryListAll())
}

func (s *encryptedStore) OrgRegistryFind(orgID int64, addr string) (*model.Registry, error) {
	return s.registry(s.Store.OrgRegistryFind(orgID, addr))
}

func (s *encryptedStore) OrgRegistryList(orgID int64, p *model.ListOptionsWithAll) ([]*model.Registry, error) {
	return s.registries(s.Store.OrgRegistryList(orgID, p))
}

func (s *encryptedStore) GlobalRegistryFind(addr string) (*model.Registry, error) {
	return s.registry(s.Store.GlobalRegistryFind(addr))
}

func (s *encryptedStore) GlobalRegistryList(p *model.ListOptionsWithAll) ([]*model.Registry, error) {
	return s.registries(s.Store.GlobalRegistryList(p))
}

// RegistryCreate stores the registry without its password first, because the
// id the password is bound to only exists afterwards.
func (s *encryptedStore) RegistryCreate(registry *model.Registry) error {
	if s.key == nil {
		return s.Store.RegistryCreate(registry)
	}
	password := registry.Password
	registry.Password = ""
	err := s.Store.RegistryCreate(registry)
	registry.Password = password
	if err != nil {
		return err
	}
	if err := s.RegistryUpdate(registry); err != nil {
		return errors.Join(err, s.RegistryDelete(registry))
	}
	return nil
}

func (s *encryptedStore) RegistryUpdate(registry *model.Registry) error {
	encrypted := *registry
	var err error
	if encrypted.Password, err = encrypt(s.key, registry.Password, registryAD(registry.ID)); err != nil {
		return err
	}
	return s.Store.RegistryUpdate(&encrypted)
}

// Users.

func (s *encryptedStore) user(user *model.User, err error) (*model.User, error) {
	if err != nil {
		return nil, err
	}
	if user.AccessToken, err = decrypt(s.key, user.AccessToken, accessTokenAD(user.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt access token of user %d: %w", user.ID, err)
	}
	if user.RefreshToken, err = decrypt(s.key, user.RefreshToken, refreshTokenAD(user.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt refresh token of user %d: %w", user.ID, err)
	}
	return user, nil
}

func (s *encryptedStore) GetUser(id int64) (*model.User, error) {
	return s.user(s.Store.GetUser(id))
}

func (s *encryptedStore) GetUserByRemoteID(forgeID int64, remoteID model.ForgeRemoteID) (*model.User, error) {
	return s.user(s.Store.GetUserByRemoteID(forgeID, remoteID))
}

func (s *encryptedStore) GetUserByLogin(forgeID int64, login string) (*model.User, error) {
	return s.user(s.Store.GetUserByLogin(forgeID, login))
}

func (s *encryptedStore) GetUserList(p *model.ListOptionsWithAll) ([]*model.User, error) {
	users, err := s.Store.GetUserList(p)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if _, err := s.user(user, nil); err != nil {
			return nil, err
		}
	}
	return users, nil
}

// CreateUser stores the user without its tokens first, because the id the
// tokens are bound to only exists afterwards.
func (s *encryptedStore) CreateUser(user *model.User) error {
	if s.key == nil {
		return s.Store.CreateUser(user)
	}
	accessToken, refreshToken := user.AccessToken, user.RefreshToken
	user.AccessToken, user.RefreshToken = "", ""
	err := s.Store.CreateUser(user)
	user.AccessToken, user.RefreshToken = accessToken, refreshToken
	if err != nil {
		return err
	}
	if err := s.UpdateUser(user); err != nil {
		return errors.Join(err, s.DeleteUser(user))
	}
	return nil
}

func (s *encryptedStore) UpdateUser(user *model.User) error {
	encrypted := *user
	var err error
	if encrypted.AccessToken, err = encrypt(s.key, user.AccessToken, accessTokenAD(user.ID)); err != nil {
		return err
	}
	if encrypted.RefreshToken, err = encrypt(s.key, user.RefreshToken, refreshTokenAD(user.ID)); err != nil {
		return err
	}
	return s.Store.UpdateUser(&encrypted)
}
