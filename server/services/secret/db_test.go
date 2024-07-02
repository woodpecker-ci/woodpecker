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

	"github.com/franela/goblin"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/secret"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func TestSecretListPipeline(t *testing.T) {
	g := goblin.Goblin(t)
	mockStore := mocks_store.NewStore(t)

	// global secret
	globalSecret := &model.Secret{
		ID:     1,
		OrgID:  0,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-global",
	}

	// org secret
	orgSecret := &model.Secret{
		ID:     2,
		OrgID:  1,
		RepoID: 0,
		Name:   "secret",
		Value:  "value-org",
	}

	// repo secret
	repoSecret := &model.Secret{
		ID:     3,
		OrgID:  0,
		RepoID: 1,
		Name:   "secret",
		Value:  "value-repo",
	}

	g.Describe("Priority of secrets", func() {
		g.It("should get the repo secret", func() {
			mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
				globalSecret,
				orgSecret,
				repoSecret,
			}, nil)

			s, err := secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-repo")
		})

		g.It("should get the org secret", func() {
			mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
				globalSecret,
				orgSecret,
			}, nil)

			s, err := secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-org")
		})

		g.It("should get the global secret", func() {
			mockStore.On("SecretList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Secret{
				globalSecret,
			}, nil)

			s, err := secret.NewDB(mockStore).SecretListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-global")
		})
	})
}
