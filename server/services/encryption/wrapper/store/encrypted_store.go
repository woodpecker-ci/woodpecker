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
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// EncryptedStore is a store.Store decorator that routes secrets, registries
// and users through their encrypting wrappers. All other methods pass
// through to the wrapped store unchanged.
type EncryptedStore struct {
	store.Store
	secrets    *EncryptedSecretStore
	registries *EncryptedRegistryStore
	users      *EncryptedUserStore
}

// Ensure wrapper match interface.
var _ store.Store = new(EncryptedStore)

func NewEncryptedStore(s store.Store) *EncryptedStore {
	return &EncryptedStore{
		Store:      s,
		secrets:    NewSecretStore(s),
		registries: NewRegistryStore(s),
		users:      NewUserStore(s),
	}
}

// Clients returns the encryption clients of the wrapped domains. All of them
// have to be registered on a single encryption builder run, so that enabling,
// migrating and disabling covers every encrypted domain.
func (s *EncryptedStore) Clients() []types.EncryptionClient {
	return []types.EncryptionClient{s.secrets, s.registries, s.users}
}

// Secrets.

func (s *EncryptedStore) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	return s.secrets.SecretFind(repo, name)
}

func (s *EncryptedStore) SecretList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptions) ([]*model.Secret, error) {
	return s.secrets.SecretList(repo, includeGlobalAndOrg, p)
}

func (s *EncryptedStore) SecretListAll() ([]*model.Secret, error) {
	return s.secrets.SecretListAll()
}

func (s *EncryptedStore) SecretCreate(secret *model.Secret) error {
	return s.secrets.SecretCreate(secret)
}

func (s *EncryptedStore) SecretUpdate(secret *model.Secret) error {
	return s.secrets.SecretUpdate(secret)
}

func (s *EncryptedStore) SecretDelete(secret *model.Secret) error {
	return s.secrets.SecretDelete(secret)
}

func (s *EncryptedStore) OrgSecretFind(orgID int64, name string) (*model.Secret, error) {
	return s.secrets.OrgSecretFind(orgID, name)
}

func (s *EncryptedStore) OrgSecretList(orgID int64, p *model.ListOptions) ([]*model.Secret, error) {
	return s.secrets.OrgSecretList(orgID, p)
}

func (s *EncryptedStore) GlobalSecretFind(name string) (*model.Secret, error) {
	return s.secrets.GlobalSecretFind(name)
}

func (s *EncryptedStore) GlobalSecretList(p *model.ListOptions) ([]*model.Secret, error) {
	return s.secrets.GlobalSecretList(p)
}

// Registries.

func (s *EncryptedStore) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	return s.registries.RegistryFind(repo, addr)
}

func (s *EncryptedStore) RegistryList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptions) ([]*model.Registry, error) {
	return s.registries.RegistryList(repo, includeGlobalAndOrg, p)
}

func (s *EncryptedStore) RegistryListAll() ([]*model.Registry, error) {
	return s.registries.RegistryListAll()
}

func (s *EncryptedStore) RegistryCreate(registry *model.Registry) error {
	return s.registries.RegistryCreate(registry)
}

func (s *EncryptedStore) RegistryUpdate(registry *model.Registry) error {
	return s.registries.RegistryUpdate(registry)
}

func (s *EncryptedStore) RegistryDelete(registry *model.Registry) error {
	return s.registries.RegistryDelete(registry)
}

func (s *EncryptedStore) OrgRegistryFind(orgID int64, addr string) (*model.Registry, error) {
	return s.registries.OrgRegistryFind(orgID, addr)
}

func (s *EncryptedStore) OrgRegistryList(orgID int64, p *model.ListOptions) ([]*model.Registry, error) {
	return s.registries.OrgRegistryList(orgID, p)
}

func (s *EncryptedStore) GlobalRegistryFind(addr string) (*model.Registry, error) {
	return s.registries.GlobalRegistryFind(addr)
}

func (s *EncryptedStore) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	return s.registries.GlobalRegistryList(p)
}

// Users.

func (s *EncryptedStore) GetUser(id int64) (*model.User, error) {
	return s.users.GetUser(id)
}

func (s *EncryptedStore) GetUserByRemoteID(forgeID int64, remoteID model.ForgeRemoteID) (*model.User, error) {
	return s.users.GetUserByRemoteID(forgeID, remoteID)
}

func (s *EncryptedStore) GetUserByLogin(forgeID int64, login string) (*model.User, error) {
	return s.users.GetUserByLogin(forgeID, login)
}

func (s *EncryptedStore) GetUserList(p *model.ListOptions) ([]*model.User, error) {
	return s.users.GetUserList(p)
}

func (s *EncryptedStore) CreateUser(user *model.User) error {
	return s.users.CreateUser(user)
}

func (s *EncryptedStore) UpdateUser(user *model.User) error {
	return s.users.UpdateUser(user)
}
