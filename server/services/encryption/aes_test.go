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

package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/subtle/random"

	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestKeyDerivationKnownVector(t *testing.T) {
	// Pins the password -> AES key derivation (SHAKE256, 32 bytes), the key
	// id and the ciphertext format (nonce, then sealed data) so that
	// refactorings do not silently break decryption of existing data. The
	// ciphertext was made with another AES-GCM implementation.
	aes := &aesEncryptionService{}
	require.NoError(t, aes.loadCipher("this-is-a-test-password"))
	assert.Equal(t, "da16b2a62af81502faa8906f134df2220dcaa1bd44ef8112810b150d4e7563a2", aes.keyID)

	plaintext, err := aes.Decrypt("AAECAwQFBgcICQoLf+8y1zKVbljfIM/6J1LKLDEEh4t5+hYlGw==", "associated-data")
	require.NoError(t, err)
	assert.Equal(t, "plaintext", plaintext)
}

func TestKeyIDDeterministic(t *testing.T) {
	// The key id must be stable across service instances (server restarts),
	// otherwise validateKey rejects the correct key after a restart.
	password := string(random.GetRandomBytes(32))

	first := &aesEncryptionService{}
	require.NoError(t, first.loadCipher(password))

	second := &aesEncryptionService{}
	require.NoError(t, second.loadCipher(password))

	assert.NotEmpty(t, first.keyID)
	assert.Equal(t, first.keyID, second.keyID)
}

func TestKeyIDDiffersPerPassword(t *testing.T) {
	first := &aesEncryptionService{}
	require.NoError(t, first.loadCipher("password-one"))

	second := &aesEncryptionService{}
	require.NoError(t, second.loadCipher("password-two"))

	assert.NotEqual(t, first.keyID, second.keyID)
}

func TestKeyIDDiffersFromKey(t *testing.T) {
	// The key id is stored (encrypted) in the database and must never leak
	// the raw AES key or the plain key derivation of the password.
	aes := &aesEncryptionService{}
	require.NoError(t, aes.loadCipher("this-is-a-test-password"))

	// the AES key derived from this password
	key := "fd0331e5103fcd88306554e97f1e25e1b7fa73622ed18dd8a396d194f9271f6a"

	assert.NotEqual(t, key, aes.keyID)
	assert.NotContains(t, aes.keyID, key)
}

func TestValidateKeyAcrossRestart(t *testing.T) {
	password := string(random.GetRandomBytes(32))

	// first service instance: enable encryption and store the sample
	first := &aesEncryptionService{}
	require.NoError(t, first.loadCipher(password))
	sample, err := first.Encrypt(first.keyID, keyIDAssociatedData)
	require.NoError(t, err)

	// second service instance ("after restart"): same password, stored sample
	s := store_mocks.NewMockStore(t)
	s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

	second := &aesEncryptionService{store: s}
	require.NoError(t, second.loadCipher(password))
	assert.NoError(t, second.validateKey())
}

func TestValidateKeyWrongPassword(t *testing.T) {
	// sample was created with a different password -> key must be rejected
	first := &aesEncryptionService{}
	require.NoError(t, first.loadCipher("correct-password"))
	sample, err := first.Encrypt(first.keyID, keyIDAssociatedData)
	require.NoError(t, err)

	s := store_mocks.NewMockStore(t)
	s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

	second := &aesEncryptionService{store: s}
	require.NoError(t, second.loadCipher("wrong-password"))
	assert.Error(t, second.validateKey())
}
