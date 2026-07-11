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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/aead"
	insecure_clear_text_keyset "github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"

	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	store_types "go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// writeKeysetFile writes the given keyset handle as cleartext JSON to a new
// file in the test's temp dir and returns its path.
func writeKeysetFile(t *testing.T, handle *keyset.Handle) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "keyset.json")
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()
	require.NoError(t, insecure_clear_text_keyset.Write(handle, keyset.NewJSONWriter(file)))
	return path
}

// newKeysetHandle creates a fresh AES256-GCM keyset handle.
func newKeysetHandle(t *testing.T) *keyset.Handle {
	t.Helper()
	handle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)
	return handle
}

// rotatedKeysetHandle returns a copy of the given keyset with an additional
// key that is promoted to primary, mimicking a key rotation.
func rotatedKeysetHandle(t *testing.T, handle *keyset.Handle) *keyset.Handle {
	t.Helper()
	manager := keyset.NewManagerFromHandle(handle)
	keyID, err := manager.Add(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)
	require.NoError(t, manager.SetPrimary(keyID))
	rotated, err := manager.Handle()
	require.NoError(t, err)
	return rotated
}

// loadTinkService builds a tinkEncryptionService with the keyset loaded from
// the given file, without running validation or state transitions.
func loadTinkService(t *testing.T, keysetPath string, s *store_mocks.MockStore) *tinkEncryptionService {
	t.Helper()
	svc := &tinkEncryptionService{keysetFilePath: keysetPath, store: s}
	require.NoError(t, svc.loadKeyset())
	return svc
}

func TestTinkValidateKeyset(t *testing.T) {
	t.Parallel()

	t.Run("not enabled when sample missing", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return("", store_types.ErrRecordNotExist)

		svc := loadTinkService(t, writeKeysetFile(t, newKeysetHandle(t)), s)
		assert.ErrorIs(t, svc.validateKeyset(), errEncryptionNotEnabled)
	})

	t.Run("valid keyset round-trips", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		svc := loadTinkService(t, writeKeysetFile(t, newKeysetHandle(t)), s)

		sample, err := svc.Encrypt(svc.primaryKeyID, keyIDAssociatedData)
		require.NoError(t, err)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

		assert.NoError(t, svc.validateKeyset())
	})

	t.Run("rotated keyset detected when sample uses previous primary", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		oldHandle := newKeysetHandle(t)

		// sample written by the service running on the old keyset
		oldSvc := loadTinkService(t, writeKeysetFile(t, oldHandle), s)
		sample, err := oldSvc.Encrypt(oldSvc.primaryKeyID, keyIDAssociatedData)
		require.NoError(t, err)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

		// service now runs on the rotated keyset: old key still present,
		// new primary key id
		newSvc := loadTinkService(t, writeKeysetFile(t, rotatedKeysetHandle(t, oldHandle)), s)
		assert.ErrorIs(t, newSvc.validateKeyset(), errEncryptionKeyRotated)
	})

	t.Run("undecryptable sample is an invalid key, not a rotation", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)

		// sample written with a completely unrelated keyset
		otherSvc := loadTinkService(t, writeKeysetFile(t, newKeysetHandle(t)), s)
		sample, err := otherSvc.Encrypt(otherSvc.primaryKeyID, keyIDAssociatedData)
		require.NoError(t, err)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

		svc := loadTinkService(t, writeKeysetFile(t, newKeysetHandle(t)), s)
		err = svc.validateKeyset()
		assert.NotErrorIs(t, err, errEncryptionKeyRotated,
			"decryption failure must not be mistaken for a key rotation")
		assert.ErrorIs(t, err, errEncryptionKeyInvalid)
	})
}
