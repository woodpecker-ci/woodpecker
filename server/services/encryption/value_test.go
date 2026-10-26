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
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/tink"
)

func testKey(t *testing.T, rawKey string) tink.AEAD {
	t.Helper()
	key, _, err := rawAESKey(rawKey)
	require.NoError(t, err)
	return key
}

func TestEncryptDecrypt(t *testing.T) {
	key := testKey(t, "key")

	t.Run("a value round trips and is marked", func(t *testing.T) {
		encrypted, err := encrypt(key, "value", "secret/1")
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(encrypted, encryptedPrefix))
		assert.NotContains(t, encrypted, "value")

		again, err := encrypt(key, "value", "secret/1")
		require.NoError(t, err)
		assert.NotEqual(t, encrypted, again, "nonce must be random")

		decrypted, err := decrypt(key, encrypted, "secret/1")
		require.NoError(t, err)
		assert.Equal(t, "value", decrypted)
	})

	t.Run("a ciphertext is bound to its associated data and key", func(t *testing.T) {
		encrypted, err := encrypt(key, "value", "secret/1")
		require.NoError(t, err)

		_, err = decrypt(key, encrypted, "secret/2")
		assert.Error(t, err)
		_, err = decrypt(testKey(t, "other"), encrypted, "secret/1")
		assert.Error(t, err)
	})

	t.Run("plaintext passes through", func(t *testing.T) {
		decrypted, err := decrypt(key, "value", "secret/1")
		require.NoError(t, err)
		assert.Equal(t, "value", decrypted)
	})

	t.Run("without a key values stay plaintext and ciphertext is refused", func(t *testing.T) {
		plaintext, err := encrypt(nil, "value", "secret/1")
		require.NoError(t, err)
		assert.Equal(t, "value", plaintext)

		encrypted, err := encrypt(key, "value", "secret/1")
		require.NoError(t, err)
		_, err = decrypt(nil, encrypted, "secret/1")
		assert.ErrorIs(t, err, errNoKey)
	})

	t.Run("broken ciphertext is an error", func(t *testing.T) {
		_, err := decrypt(key, encryptedPrefix+"not base64!", "secret/1")
		assert.Error(t, err)

		tooShort := encryptedPrefix + base64.StdEncoding.EncodeToString([]byte("short"))
		_, err = decrypt(key, tooShort, "secret/1")
		assert.Error(t, err)
	})
}

func TestConvert(t *testing.T) {
	key := testKey(t, "key")
	encrypted, err := encrypt(key, "value", "ad")
	require.NoError(t, err)

	t.Run("encrypts plaintext", func(t *testing.T) {
		converted, err := convert("value", "ad", key, key, false)
		require.NoError(t, err)
		decrypted, err := decrypt(key, converted, "ad")
		require.NoError(t, err)
		assert.Equal(t, "value", decrypted)
	})

	t.Run("keeps encrypted values unless all are converted", func(t *testing.T) {
		converted, err := convert(encrypted, "ad", key, key, false)
		require.NoError(t, err)
		assert.Equal(t, encrypted, converted)

		converted, err = convert(encrypted, "ad", key, key, true)
		require.NoError(t, err)
		assert.NotEqual(t, encrypted, converted)
		decrypted, err := decrypt(key, converted, "ad")
		require.NoError(t, err)
		assert.Equal(t, "value", decrypted)
	})

	t.Run("decrypts to plaintext and keeps plaintext", func(t *testing.T) {
		converted, err := convert(encrypted, "ad", key, nil, false)
		require.NoError(t, err)
		assert.Equal(t, "value", converted)

		converted, err = convert("value", "ad", key, nil, false)
		require.NoError(t, err)
		assert.Equal(t, "value", converted)
	})

	t.Run("fails on values the old key cannot decrypt", func(t *testing.T) {
		_, err := convert(encrypted, "ad", testKey(t, "other"), key, true)
		assert.Error(t, err)
	})
}
