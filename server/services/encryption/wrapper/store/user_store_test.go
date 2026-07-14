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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestUserTokenEncryption(t *testing.T) {
	t.Parallel()

	cipher := &fakeCipher{prefix: "k1:"}

	t.Run("update stores encrypted tokens and keeps plaintext in caller", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		var storedAccess, storedRefresh string
		s.On("UpdateUser", mock.AnythingOfType("*model.User")).
			Run(func(args mock.Arguments) {
				u := args.Get(0).(*model.User)
				storedAccess, storedRefresh = u.AccessToken, u.RefreshToken
			}).
			Return(nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		user := &model.User{ID: 5, AccessToken: "access", RefreshToken: "refresh"}
		require.NoError(t, wrapper.UpdateUser(user))
		assert.Equal(t, cipher.fullPrefix()+"access", storedAccess)
		assert.Equal(t, cipher.fullPrefix()+"refresh", storedRefresh)
		assert.Equal(t, "access", user.AccessToken)
		assert.Equal(t, "refresh", user.RefreshToken)
	})

	t.Run("empty tokens stay empty", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		var storedRefresh string
		s.On("UpdateUser", mock.AnythingOfType("*model.User")).
			Run(func(args mock.Arguments) { storedRefresh = args.Get(0).(*model.User).RefreshToken }).
			Return(nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		user := &model.User{ID: 5, AccessToken: "access"}
		require.NoError(t, wrapper.UpdateUser(user))
		assert.Empty(t, storedRefresh)
	})

	t.Run("reads decrypt tokens", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetUser", int64(5)).Return(&model.User{
			ID:           5,
			AccessToken:  cipher.fullPrefix() + "access",
			RefreshToken: cipher.fullPrefix() + "refresh",
		}, nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		user, err := wrapper.GetUser(5)
		require.NoError(t, err)
		assert.Equal(t, "access", user.AccessToken)
		assert.Equal(t, "refresh", user.RefreshToken)
	})

	t.Run("plaintext tokens pass through in mixed state", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetUser", int64(5)).Return(&model.User{ID: 5, AccessToken: "legacy-plain"}, nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		user, err := wrapper.GetUser(5)
		require.NoError(t, err)
		assert.Equal(t, "legacy-plain", user.AccessToken)
	})

	t.Run("create encrypts tokens after the row exists", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		var storedAccess string
		s.On("CreateUser", mock.AnythingOfType("*model.User")).
			Run(func(args mock.Arguments) { args.Get(0).(*model.User).ID = 9 }).
			Return(nil)
		s.On("UpdateUser", mock.AnythingOfType("*model.User")).
			Run(func(args mock.Arguments) { storedAccess = args.Get(0).(*model.User).AccessToken }).
			Return(nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		user := &model.User{AccessToken: "access"}
		require.NoError(t, wrapper.CreateUser(user))
		assert.Equal(t, int64(9), user.ID)
		assert.Equal(t, cipher.fullPrefix()+"access", storedAccess)
		assert.Equal(t, "access", user.AccessToken)
	})

	t.Run("enable skips already encrypted tokens", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		encrypted := &model.User{ID: 1, AccessToken: cipher.fullPrefix() + "done"}
		plain := &model.User{ID: 2, AccessToken: "pending"}
		s.On("GetUserList", &model.ListOptions{All: true}).Return([]*model.User{encrypted, plain}, nil)
		s.On("UpdateUser", encrypted).Return(nil)
		s.On("UpdateUser", plain).Return(nil)

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		require.NoError(t, wrapper.EnableEncryption())
		assert.Equal(t, cipher.fullPrefix()+"done", encrypted.AccessToken,
			"already encrypted token must not be encrypted twice")
		assert.Equal(t, cipher.fullPrefix()+"pending", plain.AccessToken)
	})

	t.Run("failed migration keeps the previous service active", func(t *testing.T) {
		t.Parallel()
		newCipher := &fakeCipher{prefix: "k2:"}
		s := store_mocks.NewMockStore(t)
		s.On("GetUserList", &model.ListOptions{All: true}).
			Return([]*model.User{{ID: 1, AccessToken: cipher.fullPrefix() + "a"}}, nil)
		s.On("UpdateUser", mock.AnythingOfType("*model.User")).Return(errors.New("db gone"))

		wrapper := NewUserStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		assert.Error(t, wrapper.MigrateEncryption(newCipher))
		assert.Same(t, cipher, wrapper.encryption)
	})
}

func TestEncryptedStoreRouting(t *testing.T) {
	t.Parallel()

	cipher := &fakeCipher{prefix: "k1:"}

	s := store_mocks.NewMockStore(t)
	s.On("GlobalSecretFind", "token").
		Return(&model.Secret{ID: 1, Value: cipher.fullPrefix() + "sv"}, nil)
	s.On("GlobalRegistryFind", "docker.io").
		Return(&model.Registry{ID: 2, Password: cipher.fullPrefix() + "rp"}, nil)
	s.On("GetUser", int64(3)).
		Return(&model.User{ID: 3, AccessToken: cipher.fullPrefix() + "at"}, nil)
	s.On("GetUserCount").Return(int64(7), nil)

	wrapped := NewEncryptedStore(s)
	for _, client := range wrapped.Clients() {
		require.NoError(t, client.SetEncryptionService(cipher))
	}

	secret, err := wrapped.GlobalSecretFind("token")
	require.NoError(t, err)
	assert.Equal(t, "sv", secret.Value)

	registry, err := wrapped.GlobalRegistryFind("docker.io")
	require.NoError(t, err)
	assert.Equal(t, "rp", registry.Password)

	user, err := wrapped.GetUser(3)
	require.NoError(t, err)
	assert.Equal(t, "at", user.AccessToken)

	// non-encrypted methods pass through via embedding
	count, err := wrapped.GetUserCount()
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}
