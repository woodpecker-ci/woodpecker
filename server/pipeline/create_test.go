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

package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	config_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/config/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// TestConfigPersistedBeforeParsing verifies that pipeline configs are persisted
// to the store BEFORE createPipelineItems is called. This ensures that retries
// of failed pipelines can still access the config (fixes #2982).
func TestConfigPersistedBeforeParsing(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)

	pipeline := &model.Pipeline{ID: 1, RepoID: 10}
	forgeYamls := []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: []byte("steps:\n  test:\n    image: alpine")},
	}

	persistedConfig := &model.Config{ID: 1, RepoID: 10, Name: "woodpecker"}
	mockStore.On("ConfigPersist", &model.Config{
		RepoID: int64(10),
		Name:   "woodpecker",
		Data:   forgeYamls[0].Data,
	}).Return(persistedConfig, nil)

	mockStore.On("PipelineConfigCreate", &model.PipelineConfig{
		ConfigID:   int64(1),
		PipelineID: int64(1),
	}).Return(nil)

	// Call the persist and link functions directly to verify they work
	// (the actual Create function requires too many dependencies to mock fully)
	var configs []*model.Config
	for _, forgeYamlConfig := range forgeYamls {
		config, err := findOrPersistPipelineConfig(mockStore, pipeline, forgeYamlConfig)
		assert.NoError(t, err)
		configs = append(configs, config)
	}
	err := linkPipelineConfigs(mockStore, configs, pipeline.ID)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}

// A failure to persist the workflows (e.g. a locked database) must be returned
// as an error. Create runs in a bare goroutine for webhooks and crons, so a
// panic here would crash the whole server.
func TestCreateWorkflowSaveErrorReturnsError(t *testing.T) {
	repo := &model.Repo{ID: 1, UserID: 1, FullName: "octocat/hello-world", Config: ".woodpecker.yaml"}
	yaml := []byte("when:\n  - event: push\n\nsteps:\n  - name: test\n    image: alpine\n    commands:\n      - echo hi\n")

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", int64(1)).Return(&model.User{ID: 1, Login: "octocat"}, nil)
	mockStore.On("CreatePipeline", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("GetPipelineLastBefore", mock.Anything, mock.Anything, mock.Anything).Return(&model.Pipeline{}, nil)
	mockStore.On("ConfigPersist", mock.Anything).Return(&model.Config{ID: 1, RepoID: 1}, nil)
	mockStore.On("PipelineConfigCreate", mock.Anything).Return(nil)
	mockStore.On("WorkflowsCreate", mock.Anything).Return(errors.New("database is locked"))

	mockForge := forge_mocks.NewMockForge(t)
	mockForge.On("Config").Return(metadata.Forge{Type: "github", URL: "https://github.com", Host: "github.com"})
	mockForge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{}, nil)

	mockManager := manager_mocks.NewMockManager(t)
	mockManager.On("ForgeFromRepo", mock.Anything).Return(mockForge, nil)

	configService := config_service_mocks.NewMockService(t)
	configService.On("Fetch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: yaml}}, nil)
	mockManager.On("ConfigServiceFromRepo", mock.Anything).Return(configService)

	secretService := secret_service_mocks.NewMockService(t)
	secretService.On("SecretListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Secret{}, nil)
	mockManager.On("SecretServiceFromRepo", mock.Anything).Return(secretService, nil)

	registryService := registry_service_mocks.NewMockService(t)
	registryService.On("RegistryListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Registry{}, nil)
	mockManager.On("RegistryServiceFromRepo", mock.Anything).Return(registryService, nil)

	mockManager.On("EnvironmentService").Return(nil, nil)

	server.Config.Services.Manager = mockManager

	var created *model.Pipeline
	var err error
	require.NotPanics(t, func() {
		created, err = Create(t.Context(), mockStore, repo, &model.Pipeline{
			Event: model.EventPush,
			Ref:   "refs/heads/main",
		})
	})

	assert.Nil(t, created)
	assert.ErrorContains(t, err, "database is locked")
}

// Create must return ErrFiltered unwrapped so callers can use errors.Is.
func TestCreateSkipCommitMessageReturnsErrFiltered(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	repo := &model.Repo{ID: 10, UserID: 1, FullName: "octocat/hello-world"}
	mockStore.On("GetUser", int64(1)).Return(&model.User{ID: 1, Login: "octocat"}, nil)

	// Filtered before the pipeline is persisted, so the store is untouched.
	created, err := Create(t.Context(), mockStore, repo, &model.Pipeline{
		Event:   model.EventPush,
		Message: "chore: tidy up [skip ci]",
	})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, ErrFiltered)
}

// The `when` filter path persists the pipeline first, then deletes it.
func TestCreateAllWorkflowsFilteredReturnsErrFiltered(t *testing.T) {
	repo := &model.Repo{ID: 1, UserID: 1, FullName: "octocat/hello-world", Config: ".woodpecker.yaml"}

	// A push webhook against a tag-only workflow leaves nothing to run.
	yaml := []byte("when:\n  - event: tag\n\nsteps:\n  - name: test\n    image: alpine\n    commands:\n      - echo hi\n")

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetUser", int64(1)).Return(&model.User{ID: 1, Login: "octocat"}, nil)
	mockStore.On("CreatePipeline", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("GetPipelineLastBefore", mock.Anything, mock.Anything, mock.Anything).Return(&model.Pipeline{}, nil)
	mockStore.On("ConfigPersist", mock.Anything).Return(&model.Config{ID: 1, RepoID: 1}, nil)
	mockStore.On("PipelineConfigCreate", mock.Anything).Return(nil)
	// The row is created and then removed again.
	mockStore.On("DeletePipeline", mock.Anything).Return(nil).Once()

	mockForge := forge_mocks.NewMockForge(t)
	mockForge.On("Config").Return(metadata.Forge{Type: "github", URL: "https://github.com", Host: "github.com"})
	mockForge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{}, nil)

	mockManager := manager_mocks.NewMockManager(t)
	mockManager.On("ForgeFromRepo", mock.Anything).Return(mockForge, nil)

	configService := config_service_mocks.NewMockService(t)
	configService.On("Fetch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: yaml}}, nil)
	mockManager.On("ConfigServiceFromRepo", mock.Anything).Return(configService)

	secretService := secret_service_mocks.NewMockService(t)
	secretService.On("SecretListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Secret{}, nil)
	mockManager.On("SecretServiceFromRepo", mock.Anything).Return(secretService, nil)

	registryService := registry_service_mocks.NewMockService(t)
	registryService.On("RegistryListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Registry{}, nil)
	mockManager.On("RegistryServiceFromRepo", mock.Anything).Return(registryService, nil)

	mockManager.On("EnvironmentService").Return(nil, nil)

	server.Config.Services.Manager = mockManager

	created, err := Create(t.Context(), mockStore, repo, &model.Pipeline{
		Event: model.EventPush,
		Ref:   "refs/heads/main",
	})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, ErrFiltered)
}
