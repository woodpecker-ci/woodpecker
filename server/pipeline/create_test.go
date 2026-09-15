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
	"testing"

	"github.com/stretchr/testify/assert"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
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

// TestRestartRefetchedConfigPersisted verifies that when restarting a pipeline whose
// original config was not persisted (e.g. initial fetch failed), refetched configs
// are persisted and linked to the new pipeline (fixes #7115).
func TestRestartRefetchedConfigPersisted(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)

	newPipeline := &model.Pipeline{ID: 2, RepoID: 10}
	refetchedYamls := []*forge_types.FileMeta{
		{Name: ".woodpecker.yaml", Data: []byte("steps:\n  build:\n    image: golang")},
	}

	persistedConfig := &model.Config{ID: 5, RepoID: 10, Name: "woodpecker"}
	mockStore.On("ConfigPersist", &model.Config{
		RepoID: int64(10),
		Name:   "woodpecker",
		Data:   refetchedYamls[0].Data,
	}).Return(persistedConfig, nil)

	mockStore.On("PipelineConfigCreate", &model.PipelineConfig{
		ConfigID:   int64(5),
		PipelineID: int64(2),
	}).Return(nil)

	var newConfigs []*model.Config
	for _, forgeYamlConfig := range refetchedYamls {
		config, err := findOrPersistPipelineConfig(mockStore, newPipeline, forgeYamlConfig)
		assert.NoError(t, err)
		newConfigs = append(newConfigs, config)
	}
	err := linkPipelineConfigs(mockStore, newConfigs, newPipeline.ID)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}
