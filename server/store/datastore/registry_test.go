// Copyright 2018 Drone.IO Inc.
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
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func TestRegistryFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	err := store.RegistryCreate(&model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "foo",
		Password: "bar",
	})
	assert.NoError(t, err)

	registry, err := store.RegistryFind(&model.Repo{ID: 1}, "index.docker.io")
	assert.NoError(t, err)
	assert.EqualValues(t, 1, registry.RepoID)
	assert.Equal(t, "index.docker.io", registry.Address)
	assert.Equal(t, "foo", registry.Username)
	assert.Equal(t, "bar", registry.Password)
}

func TestRegistryList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	assert.NoError(t, store.RegistryCreate(&model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "foo",
		Password: "bar",
	}))
	assert.NoError(t, store.RegistryCreate(&model.Registry{
		RepoID:   1,
		Address:  "foo.docker.io",
		Username: "foo",
		Password: "bar",
	}))

	list, err := store.RegistryList(&model.Repo{ID: 1}, false, &model.ListOptions{Page: 1, PerPage: 50})
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRegistryUpdate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	registry := &model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "foo",
		Password: "bar",
	}
	assert.NoError(t, store.RegistryCreate(registry))
	registry.Password = "qux"
	assert.NoError(t, store.RegistryUpdate(registry))
	updated, err := store.RegistryFind(&model.Repo{ID: 1}, "index.docker.io")
	assert.NoError(t, err)
	assert.Equal(t, "qux", updated.Password)
}

func TestRegistryIndexes(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	assert.NoError(t, store.RegistryCreate(&model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "foo",
		Password: "bar",
	}))

	// fail due to duplicate addr
	assert.Error(t, store.RegistryCreate(&model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "baz",
		Password: "qux",
	}))
}

func TestRegistryDelete(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry), new(model.Repo))
	defer closer()

	reg1 := &model.Registry{
		RepoID:   1,
		Address:  "index.docker.io",
		Username: "foo",
		Password: "bar",
	}
	if !assert.NoError(t, store.RegistryCreate(reg1)) {
		return
	}

	assert.NoError(t, store.RegistryDelete(reg1))
	assert.ErrorIs(t, store.RegistryDelete(reg1), types.RecordNotExist)
}

func createTestRegistries(t *testing.T, store *storage) {
	assert.NoError(t, store.RegistryCreate(&model.Registry{
		OrgID:   12,
		Address: "my.regsitry.local",
	}))
	assert.NoError(t, store.RegistryCreate(&model.Registry{
		RepoID:  1,
		Address: "private.registry.local",
	}))
	assert.NoError(t, store.RegistryCreate(&model.Registry{
		RepoID:  1,
		Address: "very-private.registry.local",
	}))
	assert.NoError(t, store.RegistryCreate(&model.Registry{
		Address: "index.docker.io",
	}))
}

func TestOrgRegistryFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	err := store.RegistryCreate(&model.Registry{
		OrgID:    12,
		Address:  "my.regsitry.local",
		Username: "username",
		Password: "password",
	})
	assert.NoError(t, err)

	registry, err := store.OrgRegistryFind(12, "my.regsitry.local")
	assert.NoError(t, err)
	assert.EqualValues(t, 12, registry.OrgID)
	assert.Equal(t, "my.regsitry.local", registry.Address)
	assert.Equal(t, "username", registry.Username)
	assert.Equal(t, "password", registry.Password)
}

func TestOrgRegistryList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	createTestRegistries(t, store)

	list, err := store.OrgRegistryList(12, &model.ListOptions{All: true})
	assert.NoError(t, err)
	require.Len(t, list, 1)

	assert.True(t, list[0].IsOrganization())
}

func TestGlobalRegistryFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	err := store.RegistryCreate(&model.Registry{
		Address:  "my.regsitry.local",
		Username: "username",
		Password: "password",
	})
	assert.NoError(t, err)

	registry, err := store.GlobalRegistryFind("my.regsitry.local")
	assert.NoError(t, err)
	assert.Equal(t, "my.regsitry.local", registry.Address)
	assert.Equal(t, "username", registry.Username)
	assert.Equal(t, "password", registry.Password)
}

func TestGlobalRegistryList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Registry))
	defer closer()

	createTestRegistries(t, store)

	list, err := store.GlobalRegistryList(&model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	assert.True(t, list[0].IsGlobal())
}
