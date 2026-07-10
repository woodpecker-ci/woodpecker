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

package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	store_types "go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func runSetupSecretService(t *testing.T, args []string, s *store_mocks.MockStore) error {
	t.Helper()
	var setupErr error
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "encryption-raw-key"},
			&cli.StringFlag{Name: "encryption-tink-keyset"},
			&cli.BoolFlag{Name: "encryption-disable-flag"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			service, err := setupSecretService(c, s, "", nil, false)
			if err == nil {
				assert.NotNil(t, service)
			}
			setupErr = err
			return nil
		},
	}
	require.NoError(t, cmd.Run(t.Context(), args))
	return setupErr
}

func TestSetupSecretService(t *testing.T) {
	t.Parallel()

	t.Run("unencrypted mode without keys succeeds", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", "encryption-ciphertext-sample").Return("", store_types.ErrRecordNotExist)

		assert.NoError(t, runSetupSecretService(t, []string{"woodpecker"}, s))
	})

	t.Run("raw key first boot enables encryption on secrets", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", "encryption-ciphertext-sample").Return("", store_types.ErrRecordNotExist)
		s.On("SecretListAll").Return([]*model.Secret{}, nil)
		s.On("ServerConfigSet", "encryption-ciphertext-sample", mock.AnythingOfType("string")).Return(nil)

		args := []string{"woodpecker", "--encryption-raw-key", "password"}
		assert.NoError(t, runSetupSecretService(t, args, s))
	})

	t.Run("conflicting key configuration is an error", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("ServerConfigGet", "encryption-ciphertext-sample").Return("", store_types.ErrRecordNotExist)

		args := []string{"woodpecker", "--encryption-raw-key", "password", "--encryption-tink-keyset", "/tmp/keyset"}
		assert.Error(t, runSetupSecretService(t, args, s))
	})
}
