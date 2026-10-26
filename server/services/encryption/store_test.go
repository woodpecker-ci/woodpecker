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

//go:build test

package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

func TestEncryptedStore(t *testing.T) {
	db := datastore.NewTestStore(t)
	s := &encryptedStore{Store: db, key: testKey(t, "key")}
	repo := &model.Repo{ID: 1}
	all := &model.ListOptionsWithAll{All: true}

	for _, secret := range []*model.Secret{
		{RepoID: repo.ID, Name: "repo", Value: "repo-value"},
		{OrgID: 2, Name: "org", Value: "org-value"},
		{Name: "global", Value: "global-value"},
	} {
		require.NoError(t, s.SecretCreate(secret))
	}
	for _, registry := range []*model.Registry{
		{RepoID: repo.ID, Address: "repo.example.com", Password: "repo-password"},
		{OrgID: 2, Address: "org.example.com", Password: "org-password"},
		{Address: "global.example.com", Password: "global-password"},
	} {
		require.NoError(t, s.RegistryCreate(registry))
	}
	user := &model.User{ForgeID: 1, ForgeRemoteID: "7", Login: "user", Hash: "hash", AccessToken: "access", RefreshToken: "refresh"}
	require.NoError(t, s.CreateUser(user))

	t.Run("secrets are decrypted by every read", func(t *testing.T) {
		secret, err := s.SecretFind(repo, "repo")
		require.NoError(t, err)
		assert.Equal(t, "repo-value", secret.Value)
		secret, err = s.OrgSecretFind(2, "org")
		require.NoError(t, err)
		assert.Equal(t, "org-value", secret.Value)
		secret, err = s.GlobalSecretFind("global")
		require.NoError(t, err)
		assert.Equal(t, "global-value", secret.Value)

		secrets, err := s.SecretList(repo, true, all)
		require.NoError(t, err)
		assert.Len(t, secrets, 2) // the org secret belongs to another org
		for _, secret := range secrets {
			assert.Equal(t, secret.Name+"-value", secret.Value)
		}
		secrets, err = s.OrgSecretList(2, all)
		require.NoError(t, err)
		require.Len(t, secrets, 1)
		assert.Equal(t, "org-value", secrets[0].Value)
		secrets, err = s.GlobalSecretList(all)
		require.NoError(t, err)
		require.Len(t, secrets, 1)
		assert.Equal(t, "global-value", secrets[0].Value)
	})

	t.Run("registry passwords are decrypted by every read", func(t *testing.T) {
		registry, err := s.RegistryFind(repo, "repo.example.com")
		require.NoError(t, err)
		assert.Equal(t, "repo-password", registry.Password)
		registry, err = s.OrgRegistryFind(2, "org.example.com")
		require.NoError(t, err)
		assert.Equal(t, "org-password", registry.Password)
		registry, err = s.GlobalRegistryFind("global.example.com")
		require.NoError(t, err)
		assert.Equal(t, "global-password", registry.Password)

		registries, err := s.RegistryList(repo, true, all)
		require.NoError(t, err)
		assert.Len(t, registries, 2)
		for _, registry := range registries {
			assert.Equal(t, registry.Address[:len(registry.Address)-len(".example.com")]+"-password", registry.Password)
		}
		registries, err = s.OrgRegistryList(2, all)
		require.NoError(t, err)
		require.Len(t, registries, 1)
		assert.Equal(t, "org-password", registries[0].Password)
		registries, err = s.GlobalRegistryList(all)
		require.NoError(t, err)
		require.Len(t, registries, 1)
		assert.Equal(t, "global-password", registries[0].Password)

		registry.Password = "updated-password"
		require.NoError(t, s.RegistryUpdate(registry))
		registry, err = s.GlobalRegistryFind("global.example.com")
		require.NoError(t, err)
		assert.Equal(t, "updated-password", registry.Password)
	})

	t.Run("user tokens are decrypted by every read", func(t *testing.T) {
		for _, get := range []func() (*model.User, error){
			func() (*model.User, error) { return s.GetUser(user.ID) },
			func() (*model.User, error) { return s.GetUserByRemoteID(1, "7") },
			func() (*model.User, error) { return s.GetUserByLogin(1, "user") },
		} {
			got, err := get()
			require.NoError(t, err)
			assert.Equal(t, "access", got.AccessToken)
			assert.Equal(t, "refresh", got.RefreshToken)
		}
	})

	t.Run("a duplicate create fails without leaving a record behind", func(t *testing.T) {
		assert.Error(t, s.SecretCreate(&model.Secret{Name: "global", Value: "duplicate"}))
		assert.Error(t, s.RegistryCreate(&model.Registry{Address: "global.example.com", Password: "duplicate"}))

		secrets, err := db.SecretListAll()
		require.NoError(t, err)
		assert.Len(t, secrets, 3)
		registries, err := db.RegistryListAll()
		require.NoError(t, err)
		assert.Len(t, registries, 3)
	})

	t.Run("a ciphertext copied to another record does not decrypt", func(t *testing.T) {
		global, err := db.GlobalSecretFind("global")
		require.NoError(t, err)
		repoSecret, err := db.SecretFind(repo, "repo")
		require.NoError(t, err)
		repoSecret.Value = global.Value
		require.NoError(t, db.SecretUpdate(repoSecret))

		_, err = s.SecretFind(repo, "repo")
		assert.ErrorContains(t, err, "secret 1")
		_, err = s.SecretList(repo, true, all)
		assert.Error(t, err)
	})
}
