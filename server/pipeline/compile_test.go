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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// compileWorkflowFixture wires up every store and service dependency
// CompileWorkflow needs for a pipeline with a single one-step workflow.
func compileWorkflowFixture(t *testing.T, persistedSteps []*model.Step) *store_mocks.MockStore {
	t.Helper()

	workflow := &model.Workflow{ID: 20, PID: 1, Name: "woodpecker", PipelineID: 1}
	pipe := &model.Pipeline{ID: 1, RepoID: 1, Event: model.EventPush}
	repo := &model.Repo{ID: 1, UserID: 1, Timeout: 42}
	user := &model.User{ID: 1}

	forge := forge_mocks.NewMockForge(t)
	forge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{Login: "user", Password: "fresh-token"}, nil)
	forge.On("Name").Return("github").Maybe()
	forge.On("URL").Return("https://github.com").Maybe()

	_store := store_mocks.NewMockStore(t)
	_store.On("WorkflowLoad", int64(20)).Return(workflow, nil)
	_store.On("GetPipeline", int64(1)).Return(pipe, nil)
	_store.On("GetRepo", int64(1)).Return(repo, nil)
	_store.On("GetUser", int64(1)).Return(user, nil)
	_store.On("ConfigsForPipeline", int64(1)).Return([]*model.Config{{
		Name: "woodpecker.yml",
		Data: []byte(`
when:
  - event: push

steps:
  - name: test
    image: alpine
    commands:
      - echo "hello world"
`),
	}}, nil)
	_store.On("GetPipelineLastBefore", mock.Anything, mock.Anything, pipe.ID).Return(&model.Pipeline{}, nil)
	_store.On("StepListFromWorkflowFind", workflow).Return(persistedSteps, nil)

	mockManager := manager_mocks.NewMockManager(t)
	mockManager.On("ForgeFromRepo", repo).Return(forge, nil)

	secretService := secret_service_mocks.NewMockService(t)
	secretService.On("SecretListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Secret{}, nil)
	mockManager.On("SecretServiceFromRepo", mock.Anything).Return(secretService, nil)

	registryService := registry_service_mocks.NewMockService(t)
	registryService.On("RegistryListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Registry{}, nil)
	mockManager.On("RegistryServiceFromRepo", mock.Anything).Return(registryService, nil)

	mockManager.On("EnvironmentService").Return(nil, nil)

	server.Config.Services.Manager = mockManager

	return _store
}

func TestCompileWorkflow(t *testing.T) {
	_store := compileWorkflowFixture(t, []*model.Step{
		{Name: "test", UUID: "persisted-uuid-1", PID: 2, PPID: 1},
	})

	rpcWorkflow, err := CompileWorkflow(t.Context(), _store, 20)
	require.NoError(t, err)
	require.NotNil(t, rpcWorkflow)

	// the payload identifies the persisted workflow
	assert.Equal(t, "20", rpcWorkflow.ID)
	assert.EqualValues(t, 42, rpcWorkflow.Timeout)

	// the compiled config carries the persisted step identity so the agent
	// reports state against the step rows that already exist
	require.Len(t, rpcWorkflow.Config.Stages, 1)
	require.Len(t, rpcWorkflow.Config.Stages[0].Steps, 1)
	step := rpcWorkflow.Config.Stages[0].Steps[0]
	assert.Equal(t, "test", step.Name)
	assert.Equal(t, "persisted-uuid-1", step.UUID)

	// fresh credentials are gathered at fetch time: the strict forge mock
	// asserts Netrc is called for every compilation

	// compiling again (workflow re-scheduled) yields the same identity
	again, err := CompileWorkflow(t.Context(), _store, 20)
	require.NoError(t, err)
	assert.Equal(t, rpcWorkflow.ID, again.ID)
	assert.Equal(t, step.UUID, again.Config.Stages[0].Steps[0].UUID)
}

func TestCompileWorkflowStepMismatch(t *testing.T) {
	// persisted steps that no longer line up with the compiled config must
	// surface as an error instead of silently mis-assigning identities
	_store := compileWorkflowFixture(t, []*model.Step{
		{Name: "test", UUID: "uuid-1", PID: 2, PPID: 1},
		{Name: "gone", UUID: "uuid-2", PID: 3, PPID: 1},
	})

	_, err := CompileWorkflow(t.Context(), _store, 20)
	assert.Error(t, err)
}
