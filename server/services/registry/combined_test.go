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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestCombinedRegistryListPipeline(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		repoName string
		dbRegs   []*model.Registry
		expected []*model.Registry
		wantErr  bool
	}{
		{
			name:     "DB registries override file registry",
			repoName: "override-test",
			dbRegs: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "docker.io", Username: "shared", Password: "db-value"},
				{ID: 2, RepoID: 1, Address: "quay.io", Username: "db-only", Password: "only-in-db"},
			},
			expected: []*model.Registry{
				{Address: "example.com", Username: "user", Password: "password-encoded", ReadOnly: true},
				{ID: 1, RepoID: 1, Address: "docker.io", Username: "shared", Password: "db-value"},
				{ID: 2, RepoID: 1, Address: "quay.io", Username: "db-only", Password: "only-in-db"},
			},
			wantErr: false,
		},
		{
			name:     "No overriding, but merged",
			repoName: "no-content",
			dbRegs: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			expected: []*model.Registry{
				{Address: "docker.io", Username: "user", Password: "your-pw", ReadOnly: true},
				{Address: "example.com", Username: "user", Password: "password-encoded", ReadOnly: true},
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			wantErr: false,
		},
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "registry-test-combined-*.json")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(`{"auths": {"docker.io": {"username": "user", "password": "your-pw"}, "example.com": {"auth": "dXNlcjpwYXNzd29yZC1lbmNvZGVk"}}}`)
	require.NoError(t, err)

	fsService := NewFilesystem(tmpFile.Name())

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store_mocks.NewMockStore(t)
			mockStore.On("RegistryList", mock.Anything, true, mock.Anything).Return(tt.dbRegs, nil)
			mockStore.On("GlobalRegistryList", mock.Anything).Return(nil, nil)

			combined := NewCombined(NewDB(mockStore), fsService)

			registries, err := combined.RegistryListPipeline(
				t.Context(),
				&model.Repo{ID: 1, Name: tt.repoName},
				&model.Pipeline{},
				nil,
			)
			if tt.wantErr {
				require.Error(t, err, "expected an error")
			} else {
				require.NoError(t, err, "error fetching registries")
			}

			assert.ElementsMatch(t, tt.expected, registries, "expected some other registries")
		})
	}
}

func TestCombinedGlobalRegistryFind(t *testing.T) {
	fs := NewFilesystem(writeDockerConfig(t, `{"auths":{"docker.io":{"username":"file","password":"pw"},"ghcr.io":{"username":"file","password":"pw"}}}`))

	s := store_mocks.NewMockStore(t)
	s.On("GlobalRegistryFind", "docker.io").Return(&model.Registry{ID: 1, Address: "docker.io", Username: "db"}, nil)
	s.On("GlobalRegistryFind", "ghcr.io").Return(nil, types.ErrRecordNotExist)
	s.On("GlobalRegistryFind", "quay.io").Return(nil, types.ErrRecordNotExist)
	s.On("GlobalRegistryFind", "broken.io").Return(nil, assert.AnError)

	combined := NewCombined(NewDB(s), fs)

	// the database wins over the docker config
	reg, err := combined.GlobalRegistryFind("docker.io")
	require.NoError(t, err)
	assert.Equal(t, "db", reg.Username)

	reg, err = combined.GlobalRegistryFind("ghcr.io")
	require.NoError(t, err)
	assert.Equal(t, &model.Registry{Address: "ghcr.io", Username: "file", Password: "pw", ReadOnly: true}, reg)

	_, err = combined.GlobalRegistryFind("quay.io")
	assert.ErrorIs(t, err, types.ErrRecordNotExist)

	_, err = combined.GlobalRegistryFind("broken.io")
	assert.ErrorIs(t, err, assert.AnError)
}

func TestCombinedGlobalRegistryList(t *testing.T) {
	fs := NewFilesystem(writeDockerConfig(t, `{"auths":{"docker.io":{"username":"file"},"ghcr.io":{"username":"file"}}}`))
	dbRegs := []*model.Registry{
		{ID: 1, Address: "docker.io", Username: "db"},
		{ID: 2, Address: "quay.io", Username: "db"},
	}

	s := store_mocks.NewMockStore(t)
	s.On("GlobalRegistryList", mock.Anything).Return(dbRegs, nil)
	combined := NewCombined(NewDB(s), fs)

	regs, err := combined.GlobalRegistryList(&model.ListOptionsWithAll{All: true})
	require.NoError(t, err)
	// a database registry hides the file one with the same address
	assert.ElementsMatch(t, []*model.Registry{
		{Address: "ghcr.io", Username: "file", ReadOnly: true},
		{ID: 1, Address: "docker.io", Username: "db"},
		{ID: 2, Address: "quay.io", Username: "db"},
	}, regs)

	regs, err = combined.GlobalRegistryList(&model.ListOptionsWithAll{ListOptions: &model.ListOptions{Page: 1, PerPage: 2}})
	require.NoError(t, err)
	assert.Len(t, regs, 2)

	t.Run("errors", func(t *testing.T) {
		s := store_mocks.NewMockStore(t)
		s.On("GlobalRegistryList", mock.Anything).Return(nil, assert.AnError)
		_, err := NewCombined(NewDB(s), fs).GlobalRegistryList(&model.ListOptionsWithAll{All: true})
		assert.ErrorIs(t, err, assert.AnError)

		s = store_mocks.NewMockStore(t)
		s.On("GlobalRegistryList", mock.Anything).Return(nil, nil)
		_, err = NewCombined(NewDB(s), NewFilesystem(writeDockerConfig(t, `{`))).GlobalRegistryList(&model.ListOptionsWithAll{All: true})
		assert.Error(t, err)
	})
}
