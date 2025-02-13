// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestParameterFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	repo := &model.Repo{
		ID: 1,
	}
	parameter := &model.Parameter{
		RepoID:      repo.ID,
		Name:        "foo",
		Type:        model.ParameterTypeString,
		Description: "test parameter",
	}

	assert.NoError(t, store.ParameterCreate(repo, parameter))
	parameter, err := store.ParameterFind(repo, parameter.Name)
	assert.NoError(t, err)
	assert.Equal(t, "foo", parameter.Name)
}

func TestParameterList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	repo := &model.Repo{
		ID: 1,
	}
	parameters := []*model.Parameter{
		{
			RepoID:      repo.ID,
			Name:        "foo",
			Type:        model.ParameterTypeString,
			Description: "test parameter 1",
		},
		{
			RepoID:      repo.ID,
			Name:        "bar",
			Type:        model.ParameterTypeBoolean,
			Description: "test parameter 2",
		},
	}

	for _, parameter := range parameters {
		assert.NoError(t, store.ParameterCreate(repo, parameter))
	}

	list, err := store.ParameterList(repo)
	assert.NoError(t, err)
	assert.Len(t, list, len(parameters))
}

func TestParameterUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	repo := &model.Repo{
		ID: 1,
	}
	parameter := &model.Parameter{
		RepoID:      repo.ID,
		Name:        "foo",
		Type:        model.ParameterTypeString,
		Description: "test parameter",
	}

	assert.NoError(t, store.ParameterCreate(repo, parameter))
	parameter.Description = "updated description"
	assert.NoError(t, store.ParameterUpdate(repo, parameter))

	updated, err := store.ParameterFind(repo, parameter.Name)
	assert.NoError(t, err)
	assert.Equal(t, "updated description", updated.Description)
}

func TestParameterDelete(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	repo := &model.Repo{
		ID: 1,
	}
	parameter := &model.Parameter{
		RepoID:      repo.ID,
		Name:        "foo",
		Type:        model.ParameterTypeString,
		Description: "test parameter",
	}

	assert.NoError(t, store.ParameterCreate(repo, parameter))
	assert.NoError(t, store.ParameterDelete(repo, parameter.Name))

	_, err := store.ParameterFind(repo, parameter.Name)
	assert.Error(t, err)
}
