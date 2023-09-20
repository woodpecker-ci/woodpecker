// Copyright 2023 Woodpecker Authors
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

package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
	yaml_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
	yaml_base_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

func TestSecretAvailable(t *testing.T) {
	secret := Secret{
		Match:      []string{"golang"},
		PluginOnly: false,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	// secret only available for "golang" plugin
	secret = Secret{
		Match:      []string{"golang"},
		PluginOnly: true,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
}

func TestCompilerCompile(t *testing.T) {
	compiler := New(
		WithMetadata(metadata.Metadata{
			Repo: metadata.Repo{
				Owner:    "octacat",
				Name:     "hello-world",
				Private:  true,
				Link:     "https://github.com/octocat/hello-world",
				CloneURL: "https://github.com/octocat/hello-world.git",
			},
		}),
		WithEnviron(map[string]string{
			"VERBOSE": "true",
			"COLORED": "true",
		}),
		WithPrefix("test"),
	)

	defaultNetworks := []*backend_types.Network{{
		Name: "test_default",
	}}
	defaultVolumes := []*backend_types.Volume{{
		Name: "test_default",
	}}

	defaultCloneStage := &backend_types.Stage{
		Name:  "test_clone",
		Alias: "clone",
		Steps: []*backend_types.Step{{
			Name:      "test_clone",
			Alias:     "clone",
			Type:      backend_types.StepTypeClone,
			Image:     constant.DefaultCloneImage,
			OnSuccess: true,
			Failure:   "fail",
			Volumes:   []string{defaultVolumes[0].Name + ":"},
			Networks:  []backend_types.Conn{{Name: "test_default", Aliases: []string{"clone"}}},
		}},
	}

	tests := []struct {
		name     string
		fronConf *yaml_types.Workflow
		backConf *backend_types.Config
		expErr   bool
	}{{
		name:     "empty workflow, no clone",
		fronConf: &yaml_types.Workflow{SkipClone: true},
		backConf: &backend_types.Config{
			Networks: defaultNetworks,
			Volumes:  defaultVolumes,
		},
	}, {
		name:     "empty workflow, default clone",
		fronConf: &yaml_types.Workflow{},
		backConf: &backend_types.Config{
			Networks: defaultNetworks,
			Volumes:  defaultVolumes,
			Stages:   []*backend_types.Stage{defaultCloneStage},
		},
	}, {
		name: "workflow with one dummy step",
		fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
			Name:  "dummy",
			Image: "dummy_img",
		}}}},
		backConf: &backend_types.Config{
			Networks: defaultNetworks,
			Volumes:  defaultVolumes,
			Stages: []*backend_types.Stage{defaultCloneStage, {
				Name:  "test_stage_0",
				Alias: "dummy",
				Steps: []*backend_types.Step{{
					Name:      "test_step_0",
					Alias:     "dummy",
					Type:      backend_types.StepTypePlugin,
					Image:     "dummy_img",
					OnSuccess: true,
					Failure:   "fail",
					Volumes:   []string{defaultVolumes[0].Name + ":"},
					Networks:  []backend_types.Conn{{Name: "test_default", Aliases: []string{"dummy"}}},
				}},
			}},
		},
	}, {
		name: "workflow with three steps and one group",
		fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
			Name:     "echo env",
			Image:    "bash",
			Commands: []string{"env"},
		}, {
			Name:     "parallel echo 1",
			Group:    "parallel",
			Image:    "bash",
			Commands: []string{"echo 1"},
		}, {
			Name:     "parallel echo 2",
			Group:    "parallel",
			Image:    "bash",
			Commands: []string{"echo 2"},
		}}}},
		backConf: &backend_types.Config{
			Networks: defaultNetworks,
			Volumes:  defaultVolumes,
			Stages: []*backend_types.Stage{defaultCloneStage, {
				Name:  "test_stage_0",
				Alias: "echo env",
				Steps: []*backend_types.Step{{
					Name:      "test_step_0",
					Alias:     "echo env",
					Type:      backend_types.StepTypeCommands,
					Image:     "bash",
					Commands:  []string{"env"},
					OnSuccess: true,
					Failure:   "fail",
					Volumes:   []string{defaultVolumes[0].Name + ":"},
					Networks:  []backend_types.Conn{{Name: "test_default", Aliases: []string{"echo env"}}},
				}},
			}, {
				Name:  "test_stage_1",
				Alias: "parallel echo 1",
				Steps: []*backend_types.Step{{
					Name:      "test_step_1",
					Alias:     "parallel echo 1",
					Type:      backend_types.StepTypeCommands,
					Image:     "bash",
					Commands:  []string{"echo 1"},
					OnSuccess: true,
					Failure:   "fail",
					Volumes:   []string{defaultVolumes[0].Name + ":"},
					Networks:  []backend_types.Conn{{Name: "test_default", Aliases: []string{"parallel echo 1"}}},
				}, {
					Name:      "test_step_2",
					Alias:     "parallel echo 2",
					Type:      backend_types.StepTypeCommands,
					Image:     "bash",
					Commands:  []string{"echo 2"},
					OnSuccess: true,
					Failure:   "fail",
					Volumes:   []string{defaultVolumes[0].Name + ":"},
					Networks:  []backend_types.Conn{{Name: "test_default", Aliases: []string{"parallel echo 2"}}},
				}},
			}},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backConf, err := compiler.Compile(test.fronConf)
			if test.expErr {
				assert.Error(t, err)
			} else {
				// we ignore uuids in steps and only check if global env got set ...
				for _, st := range backConf.Stages {
					for _, s := range st.Steps {
						s.UUID = ""
						assert.Truef(t, s.Environment["VERBOSE"] == "true", "expect to get value of global set environment")
						assert.Truef(t, len(s.Environment) > 50, "expect to have a lot of build in variables")
						s.Environment = nil
					}
				}
				// check if we get an expected backend config based on a frontend config
				assert.EqualValues(t, *test.backConf, *backConf)
			}
		})
	}
}
