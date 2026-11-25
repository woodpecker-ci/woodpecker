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

package secret_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/secret"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

var (
	globalSecret = &model.Secret{
		ID:     1,
		OrgID:  0,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-global",
	}

	orgSecret = &model.Secret{
		ID:     2,
		OrgID:  1,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-org",
	}

	repoSecret = &model.Secret{
		ID:     3,
		OrgID:  0,
		RepoID: 1,
		Name:   "secret",
		Value:  "value-repo",
	}
)

func TestSecretListPipeline(t *testing.T) {
	mockStore := store_mocks.NewMockStore(t)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
		orgSecret,
		repoSecret,
	}, nil)

	s, err := secret.NewDB(mockStore).SecretListPipeline(t.Context(), &model.Repo{}, &model.Pipeline{}, nil)
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-repo", s[0].Value)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
		orgSecret,
	}, nil)

	s, err = secret.NewDB(mockStore).SecretListPipeline(t.Context(), &model.Repo{}, &model.Pipeline{}, nil)
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-org", s[0].Value)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
	}, nil)

	s, err = secret.NewDB(mockStore).SecretListPipeline(t.Context(), &model.Repo{}, &model.Pipeline{}, nil)
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-global", s[0].Value)
}

// Deleting looks the secret up in its scope first, so only a secret of that
// repo, org or the global scope can be deleted.
func TestSecretDelete(t *testing.T) {
	repo := &model.Repo{ID: 1}

	t.Run("repo", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("SecretFind", repo, "secret").Return(repoSecret, nil)
		s.On("SecretDelete", repoSecret).Return(nil)
		s.On("SecretFind", repo, "missing").Return(nil, types.ErrRecordNotExist)

		db := secret.NewDB(s)
		assert.NoError(t, db.SecretDelete(repo, "secret"))
		assert.ErrorIs(t, db.SecretDelete(repo, "missing"), types.ErrRecordNotExist)
	})

	t.Run("org", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("OrgSecretFind", int64(1), "secret").Return(orgSecret, nil)
		s.On("SecretDelete", orgSecret).Return(nil)
		s.On("OrgSecretFind", int64(1), "missing").Return(nil, types.ErrRecordNotExist)

		db := secret.NewDB(s)
		assert.NoError(t, db.OrgSecretDelete(1, "secret"))
		assert.ErrorIs(t, db.OrgSecretDelete(1, "missing"), types.ErrRecordNotExist)
	})

	t.Run("global", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("GlobalSecretFind", "secret").Return(globalSecret, nil)
		s.On("SecretDelete", globalSecret).Return(nil)
		s.On("GlobalSecretFind", "missing").Return(nil, types.ErrRecordNotExist)

		db := secret.NewDB(s)
		assert.NoError(t, db.GlobalSecretDelete("secret"))
		assert.ErrorIs(t, db.GlobalSecretDelete("missing"), types.ErrRecordNotExist)
	})

	t.Run("store error on delete", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("GlobalSecretFind", "secret").Return(globalSecret, nil)
		s.On("SecretDelete", mock.Anything).Return(assert.AnError)

		assert.ErrorIs(t, secret.NewDB(s).GlobalSecretDelete("secret"), assert.AnError)
	})
}

func TestSecretListPipelineStoreError(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	s.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, err := secret.NewDB(s).SecretListPipeline(t.Context(), &model.Repo{}, &model.Pipeline{}, nil)
	assert.ErrorIs(t, err, assert.AnError)
}
