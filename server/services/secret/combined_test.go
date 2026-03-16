// Copyright 2024 Woodpecker Authors
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

package secret_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/secret"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCombinedSecretListPipeline(t *testing.T) {
	t.Run("DB only when no external service", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("SecretList", mock.Anything, true, mock.Anything).Return([]*model.Secret{
			{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
		}, nil)

		dbService := secret.NewDB(mockStore)
		combined := secret.NewCombined(dbService)

		secrets, err := combined.SecretListPipeline(&model.Repo{ID: 1}, &model.Pipeline{}, nil)
		require.NoError(t, err)
		assert.Len(t, secrets, 1)
		assert.Equal(t, "db-secret", secrets[0].Name)
		assert.Equal(t, "db-value", secrets[0].Value)
	})

	t.Run("external overrides DB secret by name", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("SecretList", mock.Anything, true, mock.Anything).Return([]*model.Secret{
			{ID: 1, RepoID: 1, Name: "shared", Value: "db-value"},
			{ID: 2, RepoID: 1, Name: "db-only", Value: "only-in-db"},
		}, nil)

		dbService := secret.NewDB(mockStore)

		mockExternal := &mockSecretService{
			secrets: []*model.Secret{
				{Name: "shared", Value: "external-value"},
				{Name: "ext-only", Value: "only-in-ext"},
			},
		}

		combined := secret.NewCombined(dbService, mockExternal)
		secrets, err := combined.SecretListPipeline(&model.Repo{ID: 1}, &model.Pipeline{}, nil)
		require.NoError(t, err)

		// Should have 3 unique secrets: shared (external wins), db-only, ext-only
		assert.Len(t, secrets, 3)

		secretMap := make(map[string]string)
		for _, s := range secrets {
			secretMap[s.Name] = s.Value
		}

		assert.Equal(t, "external-value", secretMap["shared"], "external should override DB")
		assert.Equal(t, "only-in-db", secretMap["db-only"], "DB-only secret preserved")
		assert.Equal(t, "only-in-ext", secretMap["ext-only"], "external-only secret preserved")
	})

	t.Run("external returns 204 no secrets", func(t *testing.T) {
		mockStore := store_mocks.NewMockStore(t)
		mockStore.On("SecretList", mock.Anything, true, mock.Anything).Return([]*model.Secret{
			{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
		}, nil)

		dbService := secret.NewDB(mockStore)

		mockExternal := &mockSecretService{
			secrets: nil, // simulates 204 No Content
		}

		combined := secret.NewCombined(dbService, mockExternal)
		secrets, err := combined.SecretListPipeline(&model.Repo{ID: 1}, &model.Pipeline{}, nil)
		require.NoError(t, err)

		assert.Len(t, secrets, 1)
		assert.Equal(t, "db-secret", secrets[0].Name)
	})
}

// mockSecretService is a minimal mock implementing only SecretListPipeline.
type mockSecretService struct {
	secrets []*model.Secret
	err     error
}

func (m *mockSecretService) SecretListPipeline(_ *model.Repo, _ *model.Pipeline, _ *model.Netrc) ([]*model.Secret, error) {
	return m.secrets, m.err
}

func (m *mockSecretService) SecretFind(_ *model.Repo, _ string) (*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) SecretList(_ *model.Repo, _ *model.ListOptions) ([]*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) SecretCreate(_ *model.Repo, _ *model.Secret) error { return nil }
func (m *mockSecretService) SecretUpdate(_ *model.Repo, _ *model.Secret) error { return nil }
func (m *mockSecretService) SecretDelete(_ *model.Repo, _ string) error        { return nil }
func (m *mockSecretService) OrgSecretFind(_ int64, _ string) (*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) OrgSecretList(_ int64, _ *model.ListOptions) ([]*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) OrgSecretCreate(_ int64, _ *model.Secret) error { return nil }
func (m *mockSecretService) OrgSecretUpdate(_ int64, _ *model.Secret) error { return nil }
func (m *mockSecretService) OrgSecretDelete(_ int64, _ string) error        { return nil }
func (m *mockSecretService) GlobalSecretFind(_ string) (*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) GlobalSecretList(_ *model.ListOptions) ([]*model.Secret, error) {
	return nil, nil
}
func (m *mockSecretService) GlobalSecretCreate(_ *model.Secret) error { return nil }
func (m *mockSecretService) GlobalSecretUpdate(_ *model.Secret) error { return nil }
func (m *mockSecretService) GlobalSecretDelete(_ string) error        { return nil }
