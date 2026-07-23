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

package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestParameterCreateFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	parameter := &model.Parameter{
		RepoID:      1,
		Name:        "DEPLOY_TARGET",
		Type:        model.ParameterTypeChoice,
		Description: "target environment",
		Default:     "staging",
		Options:     []string{"staging", "production"},
		Required:    true,
		Order:       1,
		Source:      model.ParameterSourceRepoConfig,
	}
	assert.NoError(t, store.ParameterCreate(parameter))
	assert.NotZero(t, parameter.ID)

	found, err := store.ParameterFind(&model.Repo{ID: 1}, parameter.ID)
	assert.NoError(t, err)
	assert.Equal(t, "DEPLOY_TARGET", found.Name)
	assert.Equal(t, model.ParameterTypeChoice, found.Type)
	assert.Equal(t, "staging", found.Default)
	assert.Equal(t, []string{"staging", "production"}, found.Options)
	assert.True(t, found.Required)
	assert.Equal(t, model.ParameterSourceRepoConfig, found.Source)

	// not found for other repos
	_, err = store.ParameterFind(&model.Repo{ID: 2}, parameter.ID)
	assert.Error(t, err)
}

func TestParameterCreateInvalid(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	assert.Error(t, store.ParameterCreate(&model.Parameter{
		RepoID: 1,
		Name:   "not a valid env var",
		Type:   model.ParameterTypeString,
	}))

	assert.Error(t, store.ParameterCreate(&model.Parameter{
		RepoID: 1,
		Name:   "SOME_CHOICE",
		Type:   model.ParameterTypeChoice,
		// choice without options is invalid
	}))
}

func TestParameterCreateDuplicate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	parameter := &model.Parameter{
		RepoID: 1,
		Name:   "SOME_VAR",
		Type:   model.ParameterTypeString,
	}
	assert.NoError(t, store.ParameterCreate(parameter))

	assert.Error(t, store.ParameterCreate(&model.Parameter{
		RepoID: 1,
		Name:   "SOME_VAR",
		Type:   model.ParameterTypeString,
	}))

	// same name on another repo is fine
	assert.NoError(t, store.ParameterCreate(&model.Parameter{
		RepoID: 2,
		Name:   "SOME_VAR",
		Type:   model.ParameterTypeString,
	}))
}

func TestParameterList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	assert.NoError(t, store.ParameterCreate(&model.Parameter{
		RepoID: 1,
		Name:   "B_SECOND",
		Type:   model.ParameterTypeString,
		Order:  2,
	}))
	assert.NoError(t, store.ParameterCreate(&model.Parameter{
		RepoID: 1,
		Name:   "A_FIRST",
		Type:   model.ParameterTypeBoolean,
		Order:  1,
	}))
	assert.NoError(t, store.ParameterCreate(&model.Parameter{
		RepoID: 2,
		Name:   "OTHER_REPO",
		Type:   model.ParameterTypeString,
	}))

	list, err := store.ParameterList(&model.Repo{ID: 1}, &model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "A_FIRST", list[0].Name)
	assert.Equal(t, "B_SECOND", list[1].Name)
}

func TestParameterUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	parameter := &model.Parameter{
		RepoID:  1,
		Name:    "SOME_VAR",
		Type:    model.ParameterTypeString,
		Default: "foo",
	}
	assert.NoError(t, store.ParameterCreate(parameter))

	parameter.Default = "bar"
	parameter.Required = true
	assert.NoError(t, store.ParameterUpdate(parameter))

	updated, err := store.ParameterFind(&model.Repo{ID: 1}, parameter.ID)
	assert.NoError(t, err)
	assert.Equal(t, "bar", updated.Default)
	assert.True(t, updated.Required)

	updated.Type = "invalid-type"
	assert.Error(t, store.ParameterUpdate(updated))
}

func TestParameterDelete(t *testing.T) {
	store, closer := newTestStore(t, new(model.Parameter))
	defer closer()

	parameter := &model.Parameter{
		RepoID: 1,
		Name:   "SOME_VAR",
		Type:   model.ParameterTypeString,
	}
	assert.NoError(t, store.ParameterCreate(parameter))

	// deleting from the wrong repo must fail
	assert.Error(t, store.ParameterDelete(&model.Repo{ID: 2}, parameter.ID))

	assert.NoError(t, store.ParameterDelete(&model.Repo{ID: 1}, parameter.ID))
	_, err := store.ParameterFind(&model.Repo{ID: 1}, parameter.ID)
	assert.Error(t, err)
}
