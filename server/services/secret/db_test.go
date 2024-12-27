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
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

var (
	globalSecret = &model.Secret{
		ID:     1,
		OrgID:  0,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-global",
	}

	// org secret
	orgSecret = &model.Secret{
		ID:     2,
		OrgID:  1,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-org",
	}

	// repo secret
	repoSecret = &model.Secret{
		ID:     3,
		OrgID:  0,
		RepoID: 1,
		Name:   "secret",
		Value:  "value-repo",
	}
)

func TestSecretListPipeline(t *testing.T) {
	mockStore := mocks_store.NewStore(t)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
		orgSecret,
		repoSecret,
	}, nil)

	s, err := secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-repo", s[0].Value)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
		orgSecret,
	}, nil)

	s, err = secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-org", s[0].Value)

	mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
		globalSecret,
	}, nil)

	s, err = secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)

	assert.Len(t, s, 1)
	assert.Equal(t, "value-global", s[0].Value)
}
