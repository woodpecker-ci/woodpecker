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

package variable_test

import (
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/variable"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func TestVariableListPipeline(t *testing.T) {
	g := goblin.Goblin(t)
	mockStore := mocks_store.NewStore(t)

	// global variable
	globalVariable := &model.Variable{
		ID:     1,
		OrgID:  0,
		RepoID: 0,
		Name:   "variable",
		Value:  "value-global",
	}

	// org variable
	orgVariable := &model.Variable{
		ID:     2,
		OrgID:  1,
		RepoID: 0,
		Name:   "variable",
		Value:  "value-org",
	}

	// repo variable
	repoVariable := &model.Variable{
		ID:     3,
		OrgID:  0,
		RepoID: 1,
		Name:   "variable",
		Value:  "value-repo",
	}

	g.Describe("Priority of variables", func() {
		g.It("should get the repo variable", func() {
			mockStore.On("VariableList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Variable{
				globalVariable,
				orgVariable,
				repoVariable,
			}, nil)

			s, err := variable.NewDB(mockStore).VariableListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-repo")
		})

		g.It("should get the org variable", func() {
			mockStore.On("VariableList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Variable{
				globalVariable,
				orgVariable,
			}, nil)

			s, err := variable.NewDB(mockStore).VariableListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-org")
		})

		g.It("should get the global variable", func() {
			mockStore.On("VariableList", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*model.Variable{
				globalVariable,
			}, nil)

			s, err := variable.NewDB(mockStore).VariableListPipeline(&model.Repo{}, &model.Pipeline{})
			g.Assert(err).IsNil()

			g.Assert(len(s)).Equal(1)
			g.Assert(s[0].Value).Equal("value-global")
		})
	})
}
