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

	t.Run("a value round trips", func(t *testing.T) {
		encrypted, err := encrypt(key, "value", "secret/0/0/name")
		require.NoError(t, err)
		assert.NotContains(t, encrypted, "value")

		again, err := encrypt(key, "value", "secret/0/0/name")
		require.NoError(t, err)
		assert.NotEqual(t, encrypted, again, "nonce must be random")

		decrypted, err := decrypt(key, encrypted, "secret/0/0/name")
		require.NoError(t, err)
		assert.Equal(t, "value", decrypted)
	})

	t.Run("a ciphertext is bound to its associated data and key", func(t *testing.T) {
		encrypted, err := encrypt(key, "value", "secret/0/0/name")
		require.NoError(t, err)

		_, err = decrypt(key, encrypted, "secret/0/1/name")
		assert.Error(t, err)
		_, err = decrypt(testKey(t, "other"), encrypted, "secret/0/0/name")
		assert.Error(t, err)
	})

	t.Run("broken ciphertext is an error", func(t *testing.T) {
		_, err := decrypt(key, "not base64!", "secret/0/0/name")
		assert.Error(t, err)

		_, err = decrypt(key, base64.StdEncoding.EncodeToString([]byte("short")), "secret/0/0/name")
		assert.Error(t, err)
	})
}
