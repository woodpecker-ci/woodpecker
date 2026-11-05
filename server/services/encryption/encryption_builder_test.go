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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	store_types "go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// fakeClient is a minimal EncryptionClient that records whether the service was
// set and can be made to fail on demand.
type fakeClient struct {
	setErr  error
	service types.EncryptionService
}

func (c *fakeClient) SetEncryptionService(svc types.EncryptionService) error {
	c.service = svc
	return c.setErr
}
func (c *fakeClient) EnableEncryption() error                           { return nil }
func (c *fakeClient) MigrateEncryption(_ types.EncryptionService) error { return nil }

func TestNoEncryptionService(t *testing.T) {
	t.Parallel()

	svc := &noEncryption{}

	t.Run("encrypt and decrypt are passthrough", func(t *testing.T) {
		t.Parallel()
		ct, err := svc.Encrypt("plaintext", "associated")
		require.NoError(t, err)
		assert.Equal(t, "plaintext", ct)

		pt, err := svc.Decrypt("ciphertext", "associated")
		require.NoError(t, err)
		assert.Equal(t, "ciphertext", pt)
	})

	t.Run("disable is a no-op", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, svc.Disable())
	})
}

func TestNoEncryptionBuilder(t *testing.T) {
	t.Parallel()

	t.Run("wires service into all clients", func(t *testing.T) {
		t.Parallel()
		c1, c2 := &fakeClient{}, &fakeClient{}
		svc, err := noEncryptionBuilder{}.WithClients([]types.EncryptionClient{c1, c2}).Build()
		require.NoError(t, err)
		assert.NotNil(t, svc)
		assert.Same(t, svc, c1.service)
		assert.Same(t, svc, c2.service)
	})

	t.Run("propagates client error", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("set failed")
		_, err := noEncryptionBuilder{}.
			WithClients([]types.EncryptionClient{&fakeClient{setErr: wantErr}}).
			Build()
		assert.ErrorIs(t, err, wantErr)
	})
}

func TestDetectKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "no keys", args: []string{"woodpecker"}, want: keyTypeNone},
		{name: "raw key", args: []string{"woodpecker", "--encryption-raw-key", "secret"}, want: keyTypeRaw},
		{name: "tink keyset", args: []string{"woodpecker", "--encryption-tink-keyset", "/tmp/keyset"}, want: keyTypeTink},
		{
			name:    "both keys is an error",
			args:    []string{"woodpecker", "--encryption-raw-key", "secret", "--encryption-tink-keyset", "/tmp/keyset"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got string
			var gotErr error
			cmd := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: rawKeyConfigFlag},
					&cli.StringFlag{Name: tinkKeysetFilepathConfigFlag},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					b := builder{c: c}
					got, gotErr = b.detectKeyType()
					return nil
				},
			}
			require.NoError(t, cmd.Run(t.Context(), tt.args))

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServiceBuilder(t *testing.T) {
	t.Parallel()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: rawKeyConfigFlag},
			&cli.StringFlag{Name: tinkKeysetFilepathConfigFlag},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			b := builder{c: c, store: store_mocks.NewMockStore(t)}

			none, err := b.serviceBuilder(keyTypeNone)
			require.NoError(t, err)
			assert.IsType(t, &noEncryptionBuilder{}, none)

			raw, err := b.serviceBuilder(keyTypeRaw)
			require.NoError(t, err)
			assert.NotNil(t, raw)

			tink, err := b.serviceBuilder(keyTypeTink)
			require.NoError(t, err)
			assert.NotNil(t, tink)

			_, err = b.serviceBuilder("bogus")
			assert.Error(t, err)
			return nil
		},
	}
	require.NoError(t, cmd.Run(t.Context(), []string{"woodpecker"}))
}

func TestGetServiceNoKeysIsError(t *testing.T) {
	t.Parallel()
	b := builder{}
	_, err := b.getService(keyTypeNone)
	assert.Error(t, err)
}

func TestBuilderIsEnabled(t *testing.T) {
	t.Parallel()

	t.Run("enabled when sample present", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return("sample", nil)
		enabled, err := builder{store: s}.isEnabled()
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("disabled when record missing", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return("", store_types.ErrRecordNotExist)
		enabled, err := builder{store: s}.isEnabled()
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("propagates store error", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return("", errors.New("db down"))
		_, err := builder{store: s}.isEnabled()
		assert.Error(t, err)
	})
}

func TestValidateKey(t *testing.T) {
	t.Parallel()

	t.Run("not enabled when sample missing", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return("", store_types.ErrRecordNotExist)
		svc := &aesEncryptionService{store: s}
		require.NoError(t, svc.loadCipher(string(random.GetRandomBytes(32))))

		err := svc.validateKey()
		assert.ErrorIs(t, err, errEncryptionNotEnabled)
	})

	t.Run("invalid key when sample does not decrypt to keyID", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		svc := &aesEncryptionService{store: s}
		// loadCipher derives and sets svc.keyID from the password
		require.NoError(t, svc.loadCipher(string(random.GetRandomBytes(32))))

		// store a sample that encrypts something other than the derived keyID
		sample, err := svc.Encrypt("not-the-key-id", keyIDAssociatedData)
		require.NoError(t, err)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

		err = svc.validateKey()
		assert.ErrorIs(t, err, errEncryptionKeyInvalid)
	})

	t.Run("valid key round-trips", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		svc := &aesEncryptionService{store: s}
		require.NoError(t, svc.loadCipher(string(random.GetRandomBytes(32))))

		// the sample must decrypt back to the derived keyID
		sample, err := svc.Encrypt(svc.keyID, keyIDAssociatedData)
		require.NoError(t, err)
		s.On("ServerConfigGet", ciphertextSampleConfigKey).Return(sample, nil)

		assert.NoError(t, svc.validateKey())
	})
}
