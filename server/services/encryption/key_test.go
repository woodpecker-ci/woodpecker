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
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/aead"
	insecure_clear_text_keyset "github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
	"github.com/urfave/cli/v3"
)

func loadKeyWith(t *testing.T, args ...string) (tink.AEAD, string, error) {
	t.Helper()
	var (
		key   tink.AEAD
		keyID string
		err   error
	)
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: rawKeyFlag},
			&cli.StringFlag{Name: tinkKeysetFlag},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			key, keyID, err = loadKey(c)
			return nil
		},
	}
	require.NoError(t, cmd.Run(t.Context(), append([]string{"woodpecker-server"}, args...)))
	return key, keyID, err
}

func TestLoadKey(t *testing.T) {
	t.Run("no key configured", func(t *testing.T) {
		key, _, err := loadKeyWith(t)
		require.NoError(t, err)
		assert.Nil(t, key)

		// a blank WOODPECKER_ENCRYPTION_KEY is trimmed to an empty value
		key, _, err = loadKeyWith(t, "--encryption-raw-key", "")
		require.NoError(t, err)
		assert.Nil(t, key)
	})

	t.Run("raw key and keyset at once", func(t *testing.T) {
		_, _, err := loadKeyWith(t, "--encryption-raw-key", "key", "--encryption-tink-keyset", "keyset.json")
		assert.Error(t, err)
	})

	t.Run("the raw key id identifies the key", func(t *testing.T) {
		key, id, err := loadKeyWith(t, "--encryption-raw-key", "key")
		require.NoError(t, err)
		require.NotNil(t, key)
		_, sameID, err := loadKeyWith(t, "--encryption-raw-key", "key")
		require.NoError(t, err)
		_, otherID, err := loadKeyWith(t, "--encryption-raw-key", "other")
		require.NoError(t, err)

		assert.Equal(t, id, sameID)
		assert.NotEqual(t, id, otherID)
		assert.Len(t, id, 64)
	})

	t.Run("the tink key id is the primary key id", func(t *testing.T) {
		handle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
		require.NoError(t, err)
		path := filepath.Join(t.TempDir(), "keyset.json")
		file, err := os.Create(path)
		require.NoError(t, err)
		require.NoError(t, insecure_clear_text_keyset.Write(handle, keyset.NewJSONWriter(file)))
		require.NoError(t, file.Close())

		key, id, err := loadKeyWith(t, "--encryption-tink-keyset", path)
		require.NoError(t, err)
		require.NotNil(t, key)
		assert.Equal(t, strconv.FormatUint(uint64(handle.KeysetInfo().GetPrimaryKeyId()), 10), id)
	})

	t.Run("a missing or broken keyset file", func(t *testing.T) {
		_, _, err := loadKeyWith(t, "--encryption-tink-keyset", filepath.Join(t.TempDir(), "missing.json"))
		assert.Error(t, err)

		path := filepath.Join(t.TempDir(), "broken.json")
		require.NoError(t, os.WriteFile(path, []byte("{"), 0o600))
		_, _, err = loadKeyWith(t, "--encryption-tink-keyset", path)
		assert.Error(t, err)
	})
}
