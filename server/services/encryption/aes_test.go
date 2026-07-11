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
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/subtle/random"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestShortMessageLongKey(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	assert.NoError(t, err)

	input := string(random.GetRandomBytes(4))
	cipher, err := aes.Encrypt(input, "")
	assert.NoError(t, err)

	output, err := aes.Decrypt(cipher, "")
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}

func TestLongMessageShortKey(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(12)))
	assert.NoError(t, err)

	input := string(random.GetRandomBytes(1024))
	cipher, err := aes.Encrypt(input, "")
	assert.NoError(t, err)

	output, err := aes.Decrypt(cipher, "")
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}

func TestEncryptDecryptWithAssociatedData(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	require.NoError(t, err)

	plaintext := "secret-value-12345"
	associatedData := "repo:123"

	ciphertext, err := aes.Encrypt(plaintext, associatedData)
	require.NoError(t, err)

	// Decrypt with correct associated data should succeed
	decrypted, err := aes.Decrypt(ciphertext, associatedData)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Decrypt with wrong associated data should fail
	_, err = aes.Decrypt(ciphertext, "repo:456")
	assert.Error(t, err, "decryption should fail with wrong associated data")

	// Decrypt with empty associated data should fail
	_, err = aes.Decrypt(ciphertext, "")
	assert.Error(t, err, "decryption should fail with missing associated data")
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	require.NoError(t, err)

	plaintext := "same-message"
	ciphertexts := make(map[string]bool)

	// Encrypt the same message multiple times
	for range 100 {
		ct, err := aes.Encrypt(plaintext, "")
		require.NoError(t, err)
		assert.False(t, ciphertexts[ct], "ciphertext should be unique due to random nonce")
		ciphertexts[ct] = true
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	require.NoError(t, err)

	plaintext := "sensitive-data"
	ciphertext, err := aes.Encrypt(plaintext, "")
	require.NoError(t, err)

	// Decode, tamper, re-encode
	decoded, err := base64.StdEncoding.DecodeString(unmark(ciphertext))
	require.NoError(t, err)

	// Tamper with the ciphertext (flip a bit in the middle)
	if len(decoded) > AES_GCM_SIV_NonceSize+1 {
		decoded[AES_GCM_SIV_NonceSize+1] ^= 0xFF
	}
	tampered := markEncrypted(base64.StdEncoding.EncodeToString(decoded))

	_, err = aes.Decrypt(tampered, "")
	assert.Error(t, err, "decryption of tampered ciphertext should fail")
}

func TestDecryptInvalidBase64(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	require.NoError(t, err)

	_, err = aes.Decrypt(markEncrypted("not-valid-base64!!!"), "")
	assert.Error(t, err, "decryption of invalid base64 should fail")
}

func TestDecryptTruncatedCiphertext(t *testing.T) {
	aes := &aesEncryptionService{}
	err := aes.loadCipher(string(random.GetRandomBytes(32)))
	require.NoError(t, err)

	plaintext := "test-message"
	ciphertext, err := aes.Encrypt(plaintext, "")
	require.NoError(t, err)

	// Truncate the ciphertext
	decoded, err := base64.StdEncoding.DecodeString(unmark(ciphertext))
	require.NoError(t, err)

	truncated := markEncrypted(base64.StdEncoding.EncodeToString(decoded[:len(decoded)/2]))
	_, err = aes.Decrypt(truncated, "")
	assert.Error(t, err, "decryption of truncated ciphertext should fail")
}

func TestRandomBytesUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for range 1000 {
		bytes := random.GetRandomBytes(32)
		key := string(bytes)
		assert.False(t, seen[key], "random bytes should be unique")
		seen[key] = true
	}
}

func TestRandomBytesLength(t *testing.T) {
	tests := []uint32{1, 12, 16, 32, 64, 128, 256}

	for _, length := range tests {
		bytes := random.GetRandomBytes(length)
		assert.Len(t, bytes, int(length), "random bytes should have requested length")
	}
}

func TestKeyDerivationKnownVector(t *testing.T) {
	// Pins the password -> AES key derivation (SHAKE256, 32 bytes) so that
	// refactorings do not silently break decryption of existing data.
	aes := &aesEncryptionService{}
	key, err := aes.hash([]byte("this-is-a-test-password"))
	require.NoError(t, err)
	assert.Equal(t, "fd0331e5103fcd88306554e97f1e25e1b7fa73622ed18dd8a396d194f9271f6a", hex.EncodeToString(key))
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
	require.NoError(t, aes.loadCipher("some-password"))

	key, err := aes.hash([]byte("some-password"))
	require.NoError(t, err)

	assert.NotEqual(t, hex.EncodeToString(key), aes.keyID)
	assert.NotContains(t, aes.keyID, hex.EncodeToString(key))
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

func TestAesDecryptMalformedCiphertext(t *testing.T) {
	t.Parallel()

	svc := &aesEncryptionService{}
	require.NoError(t, svc.loadCipher("password"))

	tests := []struct {
		name       string
		ciphertext string
	}{
		{name: "marked empty input", ciphertext: markEncrypted("")},
		{name: "marked shorter than nonce", ciphertext: markEncrypted(base64.StdEncoding.EncodeToString([]byte("short")))},
		{name: "marked exactly nonce size but no ciphertext", ciphertext: markEncrypted(base64.StdEncoding.EncodeToString(make([]byte, AES_GCM_SIV_NonceSize)))},
		{name: "marked not base64 at all", ciphertext: markEncrypted("%%% not base64 %%%")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				_, err := svc.Decrypt(tt.ciphertext, "aad")
				assert.Error(t, err)
			})
		})
	}
}

func TestAesEncryptedValueMarker(t *testing.T) {
	t.Parallel()

	svc := &aesEncryptionService{}
	require.NoError(t, svc.loadCipher("password"))

	t.Run("encrypt marks the value as encrypted", func(t *testing.T) {
		t.Parallel()
		ciphertext, err := svc.Encrypt("plain", "aad")
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(ciphertext, types.EncryptedValuePrefix))
	})

	t.Run("marked value round-trips", func(t *testing.T) {
		t.Parallel()
		ciphertext, err := svc.Encrypt("plain", "aad")
		require.NoError(t, err)
		plaintext, err := svc.Decrypt(ciphertext, "aad")
		require.NoError(t, err)
		assert.Equal(t, "plain", plaintext)
	})

	t.Run("unmarked value is passed through as plaintext", func(t *testing.T) {
		t.Parallel()
		plaintext, err := svc.Decrypt("legacy plaintext row", "aad")
		require.NoError(t, err)
		assert.Equal(t, "legacy plaintext row", plaintext)
	})

	t.Run("marked but corrupt value is an error", func(t *testing.T) {
		t.Parallel()
		_, err := svc.Decrypt(types.EncryptedValuePrefix+"@@ not base64 @@", "aad")
		assert.Error(t, err)
	})
}
