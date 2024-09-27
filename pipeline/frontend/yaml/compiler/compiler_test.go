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

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	yaml_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
	yaml_base_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types/base"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

func TestSecretAvailable(t *testing.T) {
	secret := Secret{
		AllowedPlugins: []string{},
		Events:         []string{"push"},
	}
	assert.NoError(t, secret.Available("push", &yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))

	// secret only available for "golang" plugin
	secret = Secret{
		Name:           "foo",
		AllowedPlugins: []string{"golang"},
		Events:         []string{"push"},
	}
	assert.NoError(t, secret.Available("push", &yaml_types.Container{
		Name:     "step",
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.ErrorContains(t, secret.Available("push", &yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}), "is only allowed to be used by plugins (a filter has been set on the secret). Note: Image filters do not work for normal steps")
	assert.ErrorContains(t, secret.Available("push", &yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{},
	}), "not allowed to be used with image ")
	assert.ErrorContains(t, secret.Available("pull_request", &yaml_types.Container{
		Image: "golang",
	}), "not allowed to be used with pipeline event ")
}

func TestCompilerCompile(t *testing.T) {
	repoURL := "https://github.com/octocat/hello-world"
	compiler := New(
		WithMetadata(metadata.Metadata{
			Repo: metadata.Repo{
				Owner:    "octacat",
				Name:     "hello-world",
				Private:  true,
				ForgeURL: repoURL,
				CloneURL: "https://github.com/octocat/hello-world.git",
			},
		}),
		WithEnviron(map[string]string{
			"VERBOSE": "true",
			"COLORED": "true",
		}),
		WithPrefix("test"),
		// we use "/test" as custom workspace base to ensure the enforcement of the pluginWorkspaceBase is applied
		WithWorkspaceFromURL("/test", repoURL),
	)

	defaultNetworks := []*backend_types.Network{{
		Name: "test_default",
	}}
	defaultVolumes := []*backend_types.Volume{{
		Name: "test_default",
	}}

	defaultCloneStage := &backend_types.Stage{
		Steps: []*backend_types.Step{{
			Name:       "clone",
			Type:       backend_types.StepTypeClone,
			Image:      constant.DefaultClonePlugin,
			OnSuccess:  true,
			Failure:    "fail",
			Volumes:    []string{defaultVolumes[0].Name + ":/woodpecker"},
			WorkingDir: "/woodpecker/src/github.com/octocat/hello-world",
			Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"clone"}}},
			ExtraHosts: []backend_types.HostAlias{},
		}},
	}

	tests := []struct {
		name        string
		fronConf    *yaml_types.Workflow
		backConf    *backend_types.Config
		expectedErr string
	}{
		{
			name:     "empty workflow, no clone",
			fronConf: &yaml_types.Workflow{SkipClone: true},
			backConf: &backend_types.Config{
				Networks: defaultNetworks,
				Volumes:  defaultVolumes,
			},
		},
		{
			name:     "empty workflow, default clone",
			fronConf: &yaml_types.Workflow{},
			backConf: &backend_types.Config{
				Networks: defaultNetworks,
				Volumes:  defaultVolumes,
				Stages:   []*backend_types.Stage{defaultCloneStage},
			},
		},
		{
			name: "workflow with one dummy step",
			fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
				Name:  "dummy",
				Image: "dummy_img",
			}}}},
			backConf: &backend_types.Config{
				Networks: defaultNetworks,
				Volumes:  defaultVolumes,
				Stages: []*backend_types.Stage{defaultCloneStage, {
					Steps: []*backend_types.Step{{
						Name:       "dummy",
						Type:       backend_types.StepTypePlugin,
						Image:      "dummy_img",
						OnSuccess:  true,
						Failure:    "fail",
						Volumes:    []string{defaultVolumes[0].Name + ":/woodpecker"},
						WorkingDir: "/woodpecker/src/github.com/octocat/hello-world",
						Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"dummy"}}},
						ExtraHosts: []backend_types.HostAlias{},
					}},
				}},
			},
		},
		{
			name: "workflow with three steps",
			fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
				Name:     "echo env",
				Image:    "bash",
				Commands: []string{"env"},
			}, {
				Name:     "parallel echo 1",
				Image:    "bash",
				Commands: []string{"echo 1"},
			}, {
				Name:     "parallel echo 2",
				Image:    "bash",
				Commands: []string{"echo 2"},
			}}}},
			backConf: &backend_types.Config{
				Networks: defaultNetworks,
				Volumes:  defaultVolumes,
				Stages: []*backend_types.Stage{
					defaultCloneStage, {
						Steps: []*backend_types.Step{{
							Name:       "echo env",
							Type:       backend_types.StepTypeCommands,
							Image:      "bash",
							Commands:   []string{"env"},
							OnSuccess:  true,
							Failure:    "fail",
							Volumes:    []string{defaultVolumes[0].Name + ":/test"},
							WorkingDir: "/test/src/github.com/octocat/hello-world",
							Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"echo env"}}},
							ExtraHosts: []backend_types.HostAlias{},
						}},
					}, {
						Steps: []*backend_types.Step{{
							Name:       "parallel echo 1",
							Type:       backend_types.StepTypeCommands,
							Image:      "bash",
							Commands:   []string{"echo 1"},
							OnSuccess:  true,
							Failure:    "fail",
							Volumes:    []string{defaultVolumes[0].Name + ":/test"},
							WorkingDir: "/test/src/github.com/octocat/hello-world",
							Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"parallel echo 1"}}},
							ExtraHosts: []backend_types.HostAlias{},
						}},
					}, {
						Steps: []*backend_types.Step{{
							Name:       "parallel echo 2",
							Type:       backend_types.StepTypeCommands,
							Image:      "bash",
							Commands:   []string{"echo 2"},
							OnSuccess:  true,
							Failure:    "fail",
							Volumes:    []string{defaultVolumes[0].Name + ":/test"},
							WorkingDir: "/test/src/github.com/octocat/hello-world",
							Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"parallel echo 2"}}},
							ExtraHosts: []backend_types.HostAlias{},
						}},
					},
				},
			},
		},
		{
			name: "workflow with three steps and depends_on",
			fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
				Name:     "echo env",
				Image:    "bash",
				Commands: []string{"env"},
			}, {
				Name:      "echo 1",
				Image:     "bash",
				Commands:  []string{"echo 1"},
				DependsOn: []string{"echo env", "echo 2"},
			}, {
				Name:     "echo 2",
				Image:    "bash",
				Commands: []string{"echo 2"},
			}}}},
			backConf: &backend_types.Config{
				Networks: defaultNetworks,
				Volumes:  defaultVolumes,
				Stages: []*backend_types.Stage{defaultCloneStage, {
					Steps: []*backend_types.Step{{
						Name:       "echo env",
						Type:       backend_types.StepTypeCommands,
						Image:      "bash",
						Commands:   []string{"env"},
						OnSuccess:  true,
						Failure:    "fail",
						Volumes:    []string{defaultVolumes[0].Name + ":/test"},
						WorkingDir: "/test/src/github.com/octocat/hello-world",
						Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"echo env"}}},
						ExtraHosts: []backend_types.HostAlias{},
					}, {
						Name:       "echo 2",
						Type:       backend_types.StepTypeCommands,
						Image:      "bash",
						Commands:   []string{"echo 2"},
						OnSuccess:  true,
						Failure:    "fail",
						Volumes:    []string{defaultVolumes[0].Name + ":/test"},
						WorkingDir: "/test/src/github.com/octocat/hello-world",
						Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"echo 2"}}},
						ExtraHosts: []backend_types.HostAlias{},
					}},
				}, {
					Steps: []*backend_types.Step{{
						Name:       "echo 1",
						Type:       backend_types.StepTypeCommands,
						Image:      "bash",
						Commands:   []string{"echo 1"},
						OnSuccess:  true,
						Failure:    "fail",
						Volumes:    []string{defaultVolumes[0].Name + ":/test"},
						WorkingDir: "/test/src/github.com/octocat/hello-world",
						Networks:   []backend_types.Conn{{Name: "test_default", Aliases: []string{"echo 1"}}},
						ExtraHosts: []backend_types.HostAlias{},
					}},
				}},
			},
		},
		{
			name: "workflow with missing secret",
			fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
				Name:     "step",
				Image:    "bash",
				Commands: []string{"env"},
				Secrets:  []string{"missing"},
			}}}},
			backConf:    nil,
			expectedErr: "secret \"missing\" not found",
		},
		{
			name: "workflow with broken step dependency",
			fronConf: &yaml_types.Workflow{Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
				Name:      "dummy",
				Image:     "dummy_img",
				DependsOn: []string{"not exist"},
			}}}},
			backConf:    nil,
			expectedErr: "step 'dummy' depends on unknown step 'not exist'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backConf, err := compiler.Compile(test.fronConf)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), test.expectedErr)
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

func TestSecretMatch(t *testing.T) {
	tcl := []*struct {
		name   string
		secret Secret
		event  string
		match  bool
	}{
		{
			name:   "should match event",
			secret: Secret{Events: []string{"pull_request"}},
			event:  "pull_request",
			match:  true,
		},
		{
			name:   "should not match event",
			secret: Secret{Events: []string{"pull_request"}},
			event:  "push",
			match:  false,
		},
		{
			name:   "should match when no event filters defined",
			secret: Secret{},
			event:  "pull_request",
			match:  true,
		},
		{
			name:   "pull close should match pull",
			secret: Secret{Events: []string{"pull_request"}},
			event:  "pull_request_closed",
			match:  true,
		},
	}

	for _, tc := range tcl {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.match, tc.secret.Match(tc.event))
		})
	}
}

func TestCompilerCompilePrivileged(t *testing.T) {
	compiler := New(
		WithEscalated("test/image"),
	)

	fronConf := &yaml_types.Workflow{
		SkipClone: true,
		Steps: yaml_types.ContainerList{
			ContainerList: []*yaml_types.Container{
				{
					Name:      "privileged-plugin",
					Image:     "test/image",
					DependsOn: []string{}, // no dependencies =>  enable dag mode & all steps are executed in parallel
				},
				{
					Name:     "no-plugin",
					Image:    "test/image",
					Commands: []string{"echo 'i am not a plugin anymore'"},
				},
				{
					Name:  "not-privileged-image",
					Image: "some/other-image",
				},
			},
		},
	}

	backConf, err := compiler.Compile(fronConf)
	assert.NoError(t, err)

	assert.Len(t, backConf.Stages, 1)
	assert.Len(t, backConf.Stages[0].Steps, 3)
	assert.True(t, backConf.Stages[0].Steps[0].Privileged)
	assert.False(t, backConf.Stages[0].Steps[1].Privileged)
	assert.False(t, backConf.Stages[0].Steps[2].Privileged)
}
