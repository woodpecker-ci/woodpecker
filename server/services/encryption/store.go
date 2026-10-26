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

const maxRewritePasses = 10

var errNoKey = errors.New("value is encrypted, but no encryption key is configured")

// encryptedStore is a store.Store that encrypts secret values, registry
// passwords and user forge tokens on write and decrypts them on read. Each
// record stores the id of the key its value is encrypted with, empty for
// plaintext. The records it returns always hold plaintext and an empty key
// id.
type encryptedStore struct {
	store.Store
	// key decrypts values of every key the data may be encrypted with. It is
	// nil without encryption.
	key tink.AEAD
	// keyID identifies the key new values are encrypted with. Empty stores
	// plaintext.
	keyID string
}

// seal returns value and key id as they are stored.
func (s *encryptedStore) seal(value, associatedData string) (string, string, error) {
	if s.keyID == "" {
		return value, "", nil
	}
	encrypted, err := encrypt(s.key, value, associatedData)
	return encrypted, s.keyID, err
}

// open returns the plaintext of a stored value.
func (s *encryptedStore) open(value, keyID, associatedData string) (string, error) {
	switch {
	case keyID == "":
		return value, nil
	case s.key == nil:
		return "", errNoKey
	}
	plaintext, err := decrypt(s.key, value, associatedData)
	if err != nil {
		return "", fmt.Errorf("key %s: %w", keyID, err)
	}
	return plaintext, nil
}

// The associated data binds a ciphertext to its record, so it does not decrypt
// when copied to another record or moved to another repo or org. Secrets and
// registries are bound to their unique scope and name, which exist before the
// record is inserted, so a create is a single insert.

func secretAD(secret *model.Secret) string {
	return fmt.Sprintf("secret/%d/%d/%s", secret.OrgID, secret.RepoID, secret.Name)
}

func registryAD(registry *model.Registry) string {
	return fmt.Sprintf("registry/%d/%d/%s", registry.OrgID, registry.RepoID, registry.Address)
}

func accessTokenAD(id int64) string {
	return "user/" + strconv.FormatInt(id, 10) + "/access-token"
}

func refreshTokenAD(id int64) string {
	return "user/" + strconv.FormatInt(id, 10) + "/refresh-token"
}

// rewrite brings every stored value into the form s writes. It writes only
// the encrypted columns and only if they did not change since they were read;
// values another server changed meanwhile are rewritten in the next pass.
func (s *encryptedStore) rewrite() error {
	for range maxRewritePasses {
		conflicts, err := s.rewritePass()
		if err != nil || conflicts == 0 {
			return err
		}
	}
	return errors.New("data kept changing while it was being converted")
}

func (s *encryptedStore) rewritePass() (conflicts int, err error) {
	secrets, err := s.Store.SecretListAll()
	if err != nil {
		return 0, err
	}
	for _, secret := range secrets {
		if secret.EncryptionKeyID == s.keyID {
			continue
		}
		value, err := s.open(secret.Value, secret.EncryptionKeyID, secretAD(secret))
		if err != nil {
			return 0, fmt.Errorf("could not decrypt secret %d: %w", secret.ID, err)
		}
		oldValue, oldKeyID := secret.Value, secret.EncryptionKeyID
		if secret.Value, secret.EncryptionKeyID, err = s.seal(value, secretAD(secret)); err != nil {
			return 0, err
		}
		swapped, err := s.SecretSwapValue(secret, oldValue, oldKeyID)
		if err != nil {
			return 0, err
		}
		if !swapped {
			conflicts++
		}
	}

	registries, err := s.Store.RegistryListAll()
	if err != nil {
		return 0, err
	}
	for _, registry := range registries {
		if registry.EncryptionKeyID == s.keyID {
			continue
		}
		password, err := s.open(registry.Password, registry.EncryptionKeyID, registryAD(registry))
		if err != nil {
			return 0, fmt.Errorf("could not decrypt registry %d: %w", registry.ID, err)
		}
		oldPassword, oldKeyID := registry.Password, registry.EncryptionKeyID
		if registry.Password, registry.EncryptionKeyID, err = s.seal(password, registryAD(registry)); err != nil {
			return 0, err
		}
		swapped, err := s.RegistrySwapPassword(registry, oldPassword, oldKeyID)
		if err != nil {
			return 0, err
		}
		if !swapped {
			conflicts++
		}
	}

	users, err := s.Store.GetUserList(&model.ListOptionsWithAll{All: true})
	if err != nil {
		return 0, err
	}
	for _, user := range users {
		if user.EncryptionKeyID == s.keyID {
			continue
		}
		oldAccessToken, oldRefreshToken, oldKeyID := user.AccessToken, user.RefreshToken, user.EncryptionKeyID
		if _, err := s.user(user, nil); err != nil {
			return 0, err
		}
		sealed, err := s.sealUser(user)
		if err != nil {
			return 0, err
		}
		swapped, err := s.UserSwapTokens(sealed, oldAccessToken, oldRefreshToken, oldKeyID)
		if err != nil {
			return 0, err
		}
		if !swapped {
			conflicts++
		}
	}
	return conflicts, nil
}

// Secrets.

func (s *encryptedStore) secret(secret *model.Secret, err error) (*model.Secret, error) {
	if err != nil {
		return nil, err
	}
	if secret.Value, err = s.open(secret.Value, secret.EncryptionKeyID, secretAD(secret)); err != nil {
		return nil, fmt.Errorf("could not decrypt secret %d: %w", secret.ID, err)
	}
	secret.EncryptionKeyID = ""
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

func (s *encryptedStore) sealSecret(secret *model.Secret) (*model.Secret, error) {
	sealed := *secret
	var err error
	sealed.Value, sealed.EncryptionKeyID, err = s.seal(secret.Value, secretAD(secret))
	return &sealed, err
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

func (s *encryptedStore) SecretCreate(secret *model.Secret) error {
	sealed, err := s.sealSecret(secret)
	if err != nil {
		return err
	}
	if err := s.Store.SecretCreate(sealed); err != nil {
		return err
	}
	sealed.Value, sealed.EncryptionKeyID = secret.Value, ""
	*secret = *sealed
	return nil
}

func (s *encryptedStore) SecretUpdate(secret *model.Secret) error {
	sealed, err := s.sealSecret(secret)
	if err != nil {
		return err
	}
	return s.Store.SecretUpdate(sealed)
}

// Registries.

func (s *encryptedStore) registry(registry *model.Registry, err error) (*model.Registry, error) {
	if err != nil {
		return nil, err
	}
	if registry.Password, err = s.open(registry.Password, registry.EncryptionKeyID, registryAD(registry)); err != nil {
		return nil, fmt.Errorf("could not decrypt registry %d: %w", registry.ID, err)
	}
	registry.EncryptionKeyID = ""
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

func (s *encryptedStore) sealRegistry(registry *model.Registry) (*model.Registry, error) {
	sealed := *registry
	var err error
	sealed.Password, sealed.EncryptionKeyID, err = s.seal(registry.Password, registryAD(registry))
	return &sealed, err
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

func (s *encryptedStore) RegistryCreate(registry *model.Registry) error {
	sealed, err := s.sealRegistry(registry)
	if err != nil {
		return err
	}
	if err := s.Store.RegistryCreate(sealed); err != nil {
		return err
	}
	sealed.Password, sealed.EncryptionKeyID = registry.Password, ""
	*registry = *sealed
	return nil
}

func (s *encryptedStore) RegistryUpdate(registry *model.Registry) error {
	sealed, err := s.sealRegistry(registry)
	if err != nil {
		return err
	}
	return s.Store.RegistryUpdate(sealed)
}

// Users.

func (s *encryptedStore) user(user *model.User, err error) (*model.User, error) {
	if err != nil {
		return nil, err
	}
	if user.AccessToken, err = s.open(user.AccessToken, user.EncryptionKeyID, accessTokenAD(user.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt access token of user %d: %w", user.ID, err)
	}
	if user.RefreshToken, err = s.open(user.RefreshToken, user.EncryptionKeyID, refreshTokenAD(user.ID)); err != nil {
		return nil, fmt.Errorf("could not decrypt refresh token of user %d: %w", user.ID, err)
	}
	user.EncryptionKeyID = ""
	return user, nil
}

func (s *encryptedStore) sealUser(user *model.User) (*model.User, error) {
	sealed := *user
	var err error
	if sealed.AccessToken, sealed.EncryptionKeyID, err = s.seal(user.AccessToken, accessTokenAD(user.ID)); err != nil {
		return nil, err
	}
	if sealed.RefreshToken, _, err = s.seal(user.RefreshToken, refreshTokenAD(user.ID)); err != nil {
		return nil, err
	}
	return &sealed, nil
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

// CreateUser stores the user without tokens first, because the tokens are
// bound to the user id, which only exists afterwards. If storing the tokens
// fails, the user stays without tokens until the next login stores them.
func (s *encryptedStore) CreateUser(user *model.User) error {
	if s.keyID == "" {
		return s.Store.CreateUser(user)
	}
	accessToken, refreshToken := user.AccessToken, user.RefreshToken
	user.AccessToken, user.RefreshToken = "", ""
	err := s.Store.CreateUser(user)
	user.AccessToken, user.RefreshToken = accessToken, refreshToken
	if err != nil {
		return err
	}
	return s.UpdateUser(user)
}

func (s *encryptedStore) UpdateUser(user *model.User) error {
	sealed, err := s.sealUser(user)
	if err != nil {
		return err
	}
	return s.Store.UpdateUser(sealed)
}
