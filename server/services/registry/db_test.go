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

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// Deleting looks the registry up in its scope first, so only a registry of
// that repo, org or the global scope can be deleted.
func TestDBDelete(t *testing.T) {
	repo := &model.Repo{ID: 1}

	t.Run("repo", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		reg := &model.Registry{ID: 2, RepoID: 1, Address: "docker.io"}
		s.On("RegistryFind", repo, "docker.io").Return(reg, nil)
		s.On("RegistryDelete", reg).Return(nil)
		s.On("RegistryFind", repo, "quay.io").Return(nil, types.ErrRecordNotExist)

		d := NewDB(s)
		assert.NoError(t, d.RegistryDelete(repo, "docker.io"))
		assert.ErrorIs(t, d.RegistryDelete(repo, "quay.io"), types.ErrRecordNotExist)
	})

	t.Run("org", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		reg := &model.Registry{ID: 3, OrgID: 5, Address: "docker.io"}
		s.On("OrgRegistryFind", int64(5), "docker.io").Return(reg, nil)
		s.On("RegistryDelete", reg).Return(nil)
		s.On("OrgRegistryFind", int64(5), "quay.io").Return(nil, types.ErrRecordNotExist)

		d := NewDB(s)
		assert.NoError(t, d.OrgRegistryDelete(5, "docker.io"))
		assert.ErrorIs(t, d.OrgRegistryDelete(5, "quay.io"), types.ErrRecordNotExist)
	})

	t.Run("global", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		reg := &model.Registry{ID: 4, Address: "docker.io"}
		s.On("GlobalRegistryFind", "docker.io").Return(reg, nil)
		s.On("RegistryDelete", reg).Return(nil)
		s.On("GlobalRegistryFind", "quay.io").Return(nil, types.ErrRecordNotExist)

		d := NewDB(s)
		assert.NoError(t, d.GlobalRegistryDelete("docker.io"))
		assert.ErrorIs(t, d.GlobalRegistryDelete("quay.io"), types.ErrRecordNotExist)
	})

	t.Run("store error on delete", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("GlobalRegistryFind", "docker.io").Return(&model.Registry{ID: 4}, nil)
		s.On("RegistryDelete", mock.Anything).Return(assert.AnError)

		assert.ErrorIs(t, NewDB(s).GlobalRegistryDelete("docker.io"), assert.AnError)
	})
}

// Registries of a pipeline are unique per address, the most specific scope wins.
func TestDBRegistryListPipelinePriority(t *testing.T) {
	s := store_mocks.NewMockStore(t)
	s.On("RegistryList", mock.Anything, true, mock.Anything).Return([]*model.Registry{
		{ID: 1, Address: "docker.io", Username: "global"},
		{ID: 2, OrgID: 5, Address: "docker.io", Username: "org"},
		{ID: 3, OrgID: 5, Address: "ghcr.io", Username: "org"},
		{ID: 4, RepoID: 1, Address: "docker.io", Username: "repo"},
		{ID: 5, Address: "quay.io", Username: "global"},
	}, nil)

	regs, err := NewDB(s).RegistryListPipeline(t.Context(), &model.Repo{ID: 1}, &model.Pipeline{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, []*model.Registry{
		{ID: 4, RepoID: 1, Address: "docker.io", Username: "repo"},
		{ID: 3, OrgID: 5, Address: "ghcr.io", Username: "org"},
		{ID: 5, Address: "quay.io", Username: "global"},
	}, regs)
}
