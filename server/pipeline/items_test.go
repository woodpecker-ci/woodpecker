package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/stepbuilder"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestSetPipelineStepsOnPipeline(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		ID:    1,
		Event: model.EventPush,
	}

	workflow := &model.Workflow{
		ID: 1,
	}

	pipelineItems := []*stepbuilder.Item{{
		Workflow: &stepbuilder.Workflow{
			ID:  1,
			PID: 1,
		},
		Config: &backend_types.Config{
			Stages: []*backend_types.Stage{
				{
					Steps: []*backend_types.Step{
						{
							Name: "clone",
						},
					},
				},
				{
					Steps: []*backend_types.Step{
						{
							Name: "step",
						},
					},
				},
			},
		},
	}}

	s := store_mocks.NewMockStore(t)
	s.On("WorkflowLoad", mock.Anything).Return(workflow, nil)

	pipeline = applyWorkflowsFromStepBuilder(s, pipeline, pipelineItems)
	if len(pipeline.Workflows) != 1 {
		t.Fatal("Should generate three in total")
	}
	if pipeline.Workflows[0].PipelineID != 1 {
		t.Fatal("Should set workflow's pipeline ID")
	}
	if pipeline.Workflows[0].Children[0].PPID != 1 {
		t.Fatal("Should set step PPID")
	}
}

func TestParsePipeline(t *testing.T) {
	t.Parallel()

	pipeline := &model.Pipeline{
		ID:    1,
		Event: model.EventPush,
		AdditionalVariables: map[string]string{
			"ADDITIONAL": "value",
		},
	}

	user := &model.User{
		ID: 1,
	}

	repo := &model.Repo{
		ID: 1,
	}

	yamls := []*forge_types.FileMeta{
		{
			Name: "woodpecker.yml",
			Data: []byte(`
when:
  - event: push

steps:
  - name: test
    image: alpine
    environment:
      HELLO:
        from_secret: hello
    commands:
      - echo "hello world"
`),
		},
	}

	envs := map[string]string{
		"FOO": "bar",
	}

	forge := forge_mocks.NewMockForge(t)
	forge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{
		Login:    "user",
		Password: "password",
	}, nil)
	forge.On("Name").Return("github")
	forge.On("URL").Return("https://github.com")

	store := store_mocks.NewMockStore(t)
	store.On("GetPipelineLastBefore", mock.Anything, mock.Anything, pipeline.ID).Return(&model.Pipeline{}, nil)

	mockManager := manager_mocks.NewMockManager(t)
	server.Config.Services.Manager = mockManager

	secretService := secret_service_mocks.NewMockService(t)
	secretService.On("SecretListPipeline", mock.Anything, mock.Anything).Return([]*model.Secret{
		{
			Name:  "hello",
			Value: "secret world",
		},
	}, nil)
	mockManager.On("SecretServiceFromRepo", mock.Anything).Return(secretService, nil)

	registryService := registry_service_mocks.NewMockService(t)
	registryService.On("RegistryListPipeline", mock.Anything, mock.Anything).Return([]*model.Registry{
		{
			Address:  "docker.io",
			Username: "user",
			Password: "password",
		},
	}, nil)
	mockManager.On("RegistryServiceFromRepo", mock.Anything).Return(registryService, nil)

	mockManager.On("EnvironmentService").Return(nil, nil)

	pipelineItems, err := parsePipeline(forge, store, pipeline, user, repo, yamls, envs)
	assert.NoError(t, err)

	assert.Len(t, pipelineItems, 1)
	assert.Equal(t, "test", pipelineItems[0].Config.Stages[0].Steps[0].Name)
	assert.Equal(t, "alpine", pipelineItems[0].Config.Stages[0].Steps[0].Image)
	step := pipelineItems[0].Config.Stages[0].Steps[0]
	assert.Equal(t, []string{`echo "hello world"`}, step.Commands)
	assert.Equal(t, "value", step.Environment["ADDITIONAL"])
	assert.Equal(t, "bar", step.Environment["FOO"])
	assert.Equal(t, "secret world", step.Environment["HELLO"])
}
