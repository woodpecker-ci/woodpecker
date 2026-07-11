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

// fakeCipher is a reversible toy cipher: Encrypt prefixes the value, Decrypt
// strips the prefix and fails on values that do not carry it.
type fakeCipher struct {
	prefix string
}

func (c *fakeCipher) Encrypt(plaintext, _ string) (string, error) {
	return c.prefix + plaintext, nil
}

func (c *fakeCipher) Decrypt(ciphertext, _ string) (string, error) {
	if len(ciphertext) < len(c.prefix) || ciphertext[:len(c.prefix)] != c.prefix {
		return "", errors.New("wrong cipher")
	}
	return ciphertext[len(c.prefix):], nil
}

func (c *fakeCipher) Disable() error { return nil }

func TestMigrateEncryption(t *testing.T) {
	t.Parallel()

	oldCipher := &fakeCipher{prefix: "old:"}
	newCipher := &fakeCipher{prefix: "new:"}

	t.Run("successful migration re-encrypts and swaps the service", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		secrets := []*model.Secret{{ID: 1, Value: "old:s1"}, {ID: 2, Value: "old:s2"}}
		s.On("SecretListAll").Return(secrets, nil)
		s.On("SecretUpdate", secrets[0]).Return(nil)
		s.On("SecretUpdate", secrets[1]).Return(nil)

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(oldCipher))

		require.NoError(t, wrapper.MigrateEncryption(newCipher))
		assert.Equal(t, "new:s1", secrets[0].Value)
		assert.Equal(t, "new:s2", secrets[1].Value)
		assert.Same(t, newCipher, wrapper.encryption)
	})

	t.Run("failed migration keeps the previous service active", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		secrets := []*model.Secret{{ID: 1, Value: "old:s1"}}
		s.On("SecretListAll").Return(secrets, nil)
		s.On("SecretUpdate", secrets[0]).Return(errors.New("db gone"))

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(oldCipher))

		assert.Error(t, wrapper.MigrateEncryption(newCipher))
		assert.Same(t, oldCipher, wrapper.encryption,
			"wrapper must not switch to the new service when migration failed")
	})
}

func TestSecretWritesRestorePlaintextValue(t *testing.T) {
	t.Parallel()

	cipher := &fakeCipher{prefix: "enc:"}

	t.Run("update success leaves plaintext in caller object", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		var stored string
		s.On("SecretUpdate", mock.AnythingOfType("*model.Secret")).
			Run(func(args mock.Arguments) { stored = args.Get(0).(*model.Secret).Value }).
			Return(nil)

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		secret := &model.Secret{ID: 7, Value: "plain"}
		require.NoError(t, wrapper.SecretUpdate(secret))
		assert.Equal(t, "enc:plain", stored, "value must be stored encrypted")
		assert.Equal(t, "plain", secret.Value, "caller must see plaintext")
	})

	t.Run("update failure leaves plaintext in caller object", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("SecretUpdate", mock.AnythingOfType("*model.Secret")).Return(errors.New("db gone"))

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		secret := &model.Secret{ID: 7, Value: "plain"}
		assert.Error(t, wrapper.SecretUpdate(secret))
		assert.Equal(t, "plain", secret.Value,
			"failed update must not leak ciphertext into the caller object")
	})

	t.Run("create failure on update step leaves plaintext in caller object", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("SecretCreate", mock.AnythingOfType("*model.Secret")).Return(nil)
		s.On("SecretUpdate", mock.AnythingOfType("*model.Secret")).Return(errors.New("db gone"))
		s.On("SecretDelete", mock.AnythingOfType("*model.Secret")).Return(nil)

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(cipher))

		secret := &model.Secret{Value: "plain"}
		assert.Error(t, wrapper.SecretCreate(secret))
		assert.Equal(t, "plain", secret.Value,
			"failed create must not leak ciphertext into the caller object")
	})
}
