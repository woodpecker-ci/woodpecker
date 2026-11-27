// Copyright 2024 Woodpecker Authors
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

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	yaml_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types"
)

func TestConvertPortNumber(t *testing.T) {
	portDef := "1234"
	actualPort, err := convertPort(portDef)
	assert.NoError(t, err)
	assert.Equal(t, backend_types.Port{
		Number:   1234,
		Protocol: "",
	}, actualPort)
}

func TestConvertPortUdp(t *testing.T) {
	portDef := "1234/udp"
	actualPort, err := convertPort(portDef)
	assert.NoError(t, err)
	assert.Equal(t, backend_types.Port{
		Number:   1234,
		Protocol: "udp",
	}, actualPort)
}

func TestConvertPortWrongOrder(t *testing.T) {
	portDef := "tcp/1234"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func TestConvertPortWrongDelimiter(t *testing.T) {
	portDef := "1234|udp"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func TestConvertPortWrong(t *testing.T) {
	portDef := "http"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func TestWorkspaceVolumeAlwaysAdded(t *testing.T) {
	compiler := New(
		WithPrefix("test"),
		WithWorkspace("/woodpecker", "src"),
		WithLocal(true),
	)

	workflow := &yaml_types.Workflow{
		SkipClone: true,
		Steps: yaml_types.ContainerList{
			ContainerList: []*yaml_types.Container{{
				Name:     "test",
				Image:    "alpine",
				Commands: []string{"echo hello"},
			}},
		},
	}

	config, err := compiler.Compile(workflow)
	assert.NoError(t, err)

	assert.Equal(t, "test_default", config.Volume)

	assert.NotEmpty(t, config.Stages, "expected at least one stage")
	assert.NotEmpty(t, config.Stages[0].Steps, "expected at least one step")
	assert.Contains(t, config.Stages[0].Steps[0].Volumes, "test_default:/woodpecker",
		"workspace volume must always be added, even in local mode")
}

func TestWorkspaceModes(t *testing.T) {
	compiler := New(
		WithPrefix("test"),
		WithWorkspace("/woodpecker", "src"),
		WithVolumes(
			"/home/me/repo:/woodpecker/src",
			`C:\repo:/woodpecker/src`,
			"/etc/ssl/certs:/etc/ssl/certs",
			"/opt/cache:/woodpecker/cache:rw",
		),
	)

	workflow := &yaml_types.Workflow{
		SkipClone: true,
		Steps: yaml_types.ContainerList{
			ContainerList: []*yaml_types.Container{{
				Name:      "readonly",
				Image:     "alpine",
				Commands:  []string{"ls"},
				Workspace: "readonly",
			}, {
				Name:      "readonly-plugin",
				Image:     "example/plugin",
				Workspace: "readonly",
			}, {
				Name:      "overlay",
				Image:     "alpine",
				Commands:  []string{"touch file"},
				Workspace: "overlay",
			}, {
				Name:      "write",
				Image:     "alpine",
				Commands:  []string{"touch file"},
				Workspace: "write",
			}, {
				Name:     "default",
				Image:    "alpine",
				Commands: []string{"touch file"},
			}},
		},
		Services: yaml_types.ContainerList{
			ContainerList: []*yaml_types.Container{{
				Name:      "service",
				Image:     "postgres",
				Workspace: "readonly",
			}},
		},
	}

	config, err := compiler.Compile(workflow)
	assert.NoError(t, err)

	steps := make(map[string]*backend_types.Step)
	for _, stage := range config.Stages {
		for _, step := range stage.Steps {
			steps[step.Name] = step
		}
	}

	readOnlyVolumes := []string{
		"test_default:/woodpecker:ro",
		"/home/me/repo:/woodpecker/src:ro",
		`C:\repo:/woodpecker/src:ro`,
		"/etc/ssl/certs:/etc/ssl/certs",
		"/opt/cache:/woodpecker/cache:rw",
	}
	writableVolumes := []string{
		"test_default:/woodpecker",
		"/home/me/repo:/woodpecker/src",
		`C:\repo:/woodpecker/src`,
		"/etc/ssl/certs:/etc/ssl/certs",
		"/opt/cache:/woodpecker/cache:rw",
	}

	assert.Equal(t, backend_types.WorkspaceModeReadOnly, steps["readonly"].WorkspaceMode)
	assert.Equal(t, readOnlyVolumes, steps["readonly"].Volumes)
	assert.Equal(t, backend_types.WorkspaceModeReadOnly, steps["readonly-plugin"].WorkspaceMode)
	assert.Equal(t, backend_types.WorkspaceModeReadOnly, steps["service"].WorkspaceMode)
	assert.Equal(t, "test_default:/woodpecker:ro", steps["service"].Volumes[0])

	// the backend replaces the workspace volume, global volumes in the workspace must not keep changes
	assert.Equal(t, backend_types.WorkspaceModeOverlay, steps["overlay"].WorkspaceMode)
	assert.Equal(t, []string{
		"test_default:/woodpecker",
		"/home/me/repo:/woodpecker/src:ro",
		`C:\repo:/woodpecker/src:ro`,
		"/etc/ssl/certs:/etc/ssl/certs",
		"/opt/cache:/woodpecker/cache:rw",
	}, steps["overlay"].Volumes)

	assert.Empty(t, steps["write"].WorkspaceMode)
	assert.Equal(t, writableVolumes, steps["write"].Volumes)
	assert.Empty(t, steps["default"].WorkspaceMode)
	assert.Equal(t, writableVolumes, steps["default"].Volumes)
}

func TestWorkspaceModeInvalid(t *testing.T) {
	compiler := New(WithPrefix("test"), WithWorkspace("/woodpecker", "src"))
	_, err := compiler.Compile(&yaml_types.Workflow{
		SkipClone: true,
		Steps: yaml_types.ContainerList{ContainerList: []*yaml_types.Container{{
			Name:      "step",
			Image:     "alpine",
			Commands:  []string{"ls"},
			Workspace: "readwrite",
		}}},
	})
	assert.ErrorContains(t, err, `unknown workspace mode "readwrite"`)
}
