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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// fakeCipher is a reversible toy cipher mirroring the real services: Encrypt
// marks the value with the encrypted value marker plus a cipher-specific
// prefix, Decrypt passes unmarked values through as plaintext and fails on
// values marked by a different cipher.
type fakeCipher struct {
	prefix string
}

func (c *fakeCipher) fullPrefix() string {
	return types.EncryptedValuePrefix + c.prefix
}

func (c *fakeCipher) Encrypt(plaintext, _ string) (string, error) {
	return c.fullPrefix() + plaintext, nil
}

func (c *fakeCipher) Decrypt(ciphertext, _ string) (string, error) {
	if !strings.HasPrefix(ciphertext, types.EncryptedValuePrefix) {
		return ciphertext, nil
	}
	if !strings.HasPrefix(ciphertext, c.fullPrefix()) {
		return "", errors.New("wrong cipher")
	}
	return strings.TrimPrefix(ciphertext, c.fullPrefix()), nil
}

func (c *fakeCipher) Disable() error { return nil }

func TestMigrateEncryption(t *testing.T) {
	t.Parallel()

	oldCipher := &fakeCipher{prefix: "old:"}
	newCipher := &fakeCipher{prefix: "new:"}

	t.Run("successful migration re-encrypts and swaps the service", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		secrets := []*model.Secret{{ID: 1, Value: oldCipher.fullPrefix() + "s1"}, {ID: 2, Value: oldCipher.fullPrefix() + "s2"}}
		s.On("SecretListAll").Return(secrets, nil)
		s.On("SecretUpdate", secrets[0]).Return(nil)
		s.On("SecretUpdate", secrets[1]).Return(nil)

		wrapper := NewSecretStore(s)
		require.NoError(t, wrapper.SetEncryptionService(oldCipher))

		require.NoError(t, wrapper.MigrateEncryption(newCipher))
		assert.Equal(t, newCipher.fullPrefix()+"s1", secrets[0].Value)
		assert.Equal(t, newCipher.fullPrefix()+"s2", secrets[1].Value)
		assert.Same(t, newCipher, wrapper.encryption)
	})

	t.Run("failed migration keeps the previous service active", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		secrets := []*model.Secret{{ID: 1, Value: oldCipher.fullPrefix() + "s1"}}
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
		assert.Equal(t, cipher.fullPrefix()+"plain", stored, "value must be stored encrypted")
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

func TestDecryptListErrorWrappedOnce(t *testing.T) {
	t.Parallel()

	wrapper := NewSecretStore(store_mocks.NewMockStore(t))
	require.NoError(t, wrapper.SetEncryptionService(&fakeCipher{prefix: "enc:"}))

	err := wrapper.decryptList([]*model.Secret{{ID: 42, Value: types.EncryptedValuePrefix + "other:x"}})
	require.Error(t, err)
	assert.Equal(t, 1, strings.Count(err.Error(), "failed to decrypt secret id=42"),
		"decrypt failure must be wrapped exactly once")
}

func TestEnableEncryptionResume(t *testing.T) {
	t.Parallel()

	cipher := &fakeCipher{prefix: "k1:"}

	// one row was already encrypted by an interrupted earlier run, one is
	// still plaintext
	s := store_mocks.NewMockStore(t)
	encrypted := &model.Secret{ID: 1, Value: cipher.fullPrefix() + "done"}
	plain := &model.Secret{ID: 2, Value: "pending"}
	s.On("SecretListAll").Return([]*model.Secret{encrypted, plain}, nil)
	s.On("SecretUpdate", plain).Return(nil)
	// no SecretUpdate expectation for the encrypted row: re-encrypting it
	// would fail the test

	wrapper := NewSecretStore(s)
	require.NoError(t, wrapper.SetEncryptionService(cipher))

	require.NoError(t, wrapper.EnableEncryption())
	assert.Equal(t, cipher.fullPrefix()+"done", encrypted.Value,
		"already encrypted row must be skipped, not encrypted twice")
	assert.Equal(t, cipher.fullPrefix()+"pending", plain.Value)
}

func TestMigrateEncryptionMixedState(t *testing.T) {
	t.Parallel()

	oldCipher := &fakeCipher{prefix: "k1:"}
	newCipher := &fakeCipher{prefix: "k2:"}

	// interrupted migration: one row on the old cipher, one still plaintext
	s := store_mocks.NewMockStore(t)
	migrated := &model.Secret{ID: 1, Value: oldCipher.fullPrefix() + "s1"}
	plain := &model.Secret{ID: 2, Value: "s2"}
	s.On("SecretListAll").Return([]*model.Secret{migrated, plain}, nil)
	s.On("SecretUpdate", migrated).Return(nil)
	s.On("SecretUpdate", plain).Return(nil)

	wrapper := NewSecretStore(s)
	require.NoError(t, wrapper.SetEncryptionService(oldCipher))

	require.NoError(t, wrapper.MigrateEncryption(newCipher),
		"mixed state must not abort the migration")
	assert.Equal(t, newCipher.fullPrefix()+"s1", migrated.Value)
	assert.Equal(t, newCipher.fullPrefix()+"s2", plain.Value)
}
