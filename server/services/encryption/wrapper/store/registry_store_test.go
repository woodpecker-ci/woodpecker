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

func TestRegistryPasswordEncryption(t *testing.T) {
	t.Parallel()

	cipher := &fakeCipher{prefix: "k1:"}

	t.Run("create stores encrypted password and keeps plaintext in caller", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		var stored string
		s.On("RegistryCreate", mock.AnythingOfType("*model.Registry")).Return(nil)
		s.On("RegistryUpdate", mock.AnythingOfType("*model.Registry")).
			Run(func(args mock.Arguments) { stored = args.Get(0).(*model.Registry).Password }).
			Return(nil)

		wrapper := NewRegistryStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		registry := &model.Registry{Address: "docker.io", Password: "hunter2"}
		require.NoError(t, wrapper.RegistryCreate(registry))
		assert.Equal(t, cipher.fullPrefix()+"hunter2", stored)
		assert.Equal(t, "hunter2", registry.Password)
	})

	t.Run("find decrypts the password", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("RegistryFind", mock.Anything, "docker.io").
			Return(&model.Registry{ID: 3, Password: cipher.fullPrefix() + "hunter2"}, nil)

		wrapper := NewRegistryStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		registry, err := wrapper.RegistryFind(&model.Repo{}, "docker.io")
		require.NoError(t, err)
		assert.Equal(t, "hunter2", registry.Password)
	})

	t.Run("update failure keeps plaintext in caller", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("RegistryUpdate", mock.AnythingOfType("*model.Registry")).Return(errors.New("db gone"))

		wrapper := NewRegistryStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		registry := &model.Registry{ID: 3, Password: "hunter2"}
		assert.Error(t, wrapper.RegistryUpdate(registry))
		assert.Equal(t, "hunter2", registry.Password)
	})

	t.Run("enable skips already encrypted rows", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		encrypted := &model.Registry{ID: 1, Password: cipher.fullPrefix() + "done"}
		plain := &model.Registry{ID: 2, Password: "pending"}
		s.On("RegistryListAll").Return([]*model.Registry{encrypted, plain}, nil)
		s.On("RegistryUpdate", plain).Return(nil)

		wrapper := NewRegistryStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		require.NoError(t, wrapper.EnableEncryption())
		assert.Equal(t, cipher.fullPrefix()+"done", encrypted.Password)
		assert.Equal(t, cipher.fullPrefix()+"pending", plain.Password)
	})

	t.Run("migration in mixed state succeeds and keeps old service on failure", func(t *testing.T) {
		t.Parallel()
		newCipher := &fakeCipher{prefix: "k2:"}

		s := store_mocks.NewMockStore(t)
		migrated := &model.Registry{ID: 1, Password: cipher.fullPrefix() + "p1"}
		plain := &model.Registry{ID: 2, Password: "p2"}
		s.On("RegistryListAll").Return([]*model.Registry{migrated, plain}, nil)
		s.On("RegistryUpdate", migrated).Return(nil)
		s.On("RegistryUpdate", plain).Return(errors.New("db gone"))

		wrapper := NewRegistryStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		assert.Error(t, wrapper.MigrateEncryption(newCipher))
		assert.Same(t, cipher, wrapper.encryption)
	})
}
