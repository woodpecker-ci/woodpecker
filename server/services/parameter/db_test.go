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

package parameter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/parameter"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestParameterList(t *testing.T) {
	mockStore := mocks_store.NewStore(t)

	testParam := &model.Parameter{
		ID:          1,
		RepoID:      1,
		Name:        "test",
		Type:        "string",
		Description: "test parameter",
	}

	mockStore.On("ParameterList", mock.Anything).Return([]*model.Parameter{testParam}, nil)

	s, err := parameter.NewDB(mockStore).ParameterList(&model.Repo{})
	assert.NoError(t, err)
	assert.Len(t, s, 1)
	assert.Equal(t, "test", s[0].Name)
}
