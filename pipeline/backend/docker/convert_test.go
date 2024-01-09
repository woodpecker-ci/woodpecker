// Copyright 2022 Woodpecker Authors
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

package docker

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

func TestSplitVolumeParts(t *testing.T) {
	testdata := []struct {
		from    string
		to      []string
		success bool
	}{
		{
			from:    `Z::Z::rw`,
			to:      []string{`Z:`, `Z:`, `rw`},
			success: true,
		},
		{
			from:    `Z:\:Z:\:rw`,
			to:      []string{`Z:\`, `Z:\`, `rw`},
			success: true,
		},
		{
			from:    `Z:\git\refs:Z:\git\refs:rw`,
			to:      []string{`Z:\git\refs`, `Z:\git\refs`, `rw`},
			success: true,
		},
		{
			from:    `Z:\git\refs:Z:\git\refs`,
			to:      []string{`Z:\git\refs`, `Z:\git\refs`},
			success: true,
		},
		{
			from:    `Z:/:Z:/:rw`,
			to:      []string{`Z:/`, `Z:/`, `rw`},
			success: true,
		},
		{
			from:    `Z:/git/refs:Z:/git/refs:rw`,
			to:      []string{`Z:/git/refs`, `Z:/git/refs`, `rw`},
			success: true,
		},
		{
			from:    `Z:/git/refs:Z:/git/refs`,
			to:      []string{`Z:/git/refs`, `Z:/git/refs`},
			success: true,
		},
		{
			from:    `/test:/test`,
			to:      []string{`/test`, `/test`},
			success: true,
		},
		{
			from:    `test:/test`,
			to:      []string{`test`, `/test`},
			success: true,
		},
		{
			from:    `test:test`,
			to:      []string{`test`, `test`},
			success: true,
		},
	}
	for _, test := range testdata {
		results, err := splitVolumeParts(test.from)
		if test.success != (err == nil) {
			if reflect.DeepEqual(results, test.to) != test.success {
				t.Errorf("Expect %q matches %q is %v", test.from, results, test.to)
			}
		}
	}
}

// dummy vars to test against
var (
	testCmdStep = &backend.Step{
		Name:        "hello",
		Alias:       "hello",
		UUID:        "f51821af-4cb8-435e-a3c2-3a684185d828",
		Type:        backend.StepTypeCommands,
		Commands:    []string{"echo \"hello world\"", "ls"},
		Image:       "alpine",
		Environment: map[string]string{"SHELL": "/bin/zsh"},
	}

	testPluginStep = &backend.Step{
		Name:        "lint",
		Alias:       "lint",
		UUID:        "d841ee40-e66e-4275-bb3f-55bf89744b21",
		Type:        backend.StepTypePlugin,
		Image:       "mstruebing/editorconfig-checker",
		Environment: make(map[string]string),
	}

	testEngine = &docker{
		info: types.Info{
			Architecture:    "x86_64",
			OSType:          "linux",
			DefaultRuntime:  "runc",
			DockerRootDir:   "/var/lib/docker",
			OperatingSystem: "Archlinux",
			Name:            "SOME_HOSTNAME",
		},
	}
)

func TestToContainerName(t *testing.T) {
	assert.EqualValues(t, "wp_f51821af-4cb8-435e-a3c2-3a684185d828", toContainerName(testCmdStep))
	assert.EqualValues(t, "wp_d841ee40-e66e-4275-bb3f-55bf89744b21", toContainerName(testPluginStep))
}

func TestStepToConfig(t *testing.T) {
	// StepTypeCommands
	conf := testEngine.toConfig(testCmdStep)
	if assert.NotNil(t, conf) {
		assert.EqualValues(t, []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, conf.Cmd)
		assert.EqualValues(t, testCmdStep.UUID, conf.Labels["wp_uuid"])
	}

	// StepTypePlugin
	conf = testEngine.toConfig(testPluginStep)
	if assert.NotNil(t, conf) {
		assert.Nil(t, conf.Cmd)
		assert.EqualValues(t, testPluginStep.UUID, conf.Labels["wp_uuid"])
	}
}

func TestToEnv(t *testing.T) {
	assert.Nil(t, toEnv(nil))
	assert.EqualValues(t, []string{"A=B"}, toEnv(map[string]string{"A": "B"}))
	assert.ElementsMatch(t, []string{"A=B=C", "T=T"}, toEnv(map[string]string{"A": "B=C", "": "Z", "T": "T"}))
}

func TestToVol(t *testing.T) {
	assert.Nil(t, toVol(nil))
	assert.EqualValues(t, map[string]struct{}{"/test": {}}, toVol([]string{"test:/test"}))
}

func TestEncodeAuthToBase64(t *testing.T) {
	res, err := encodeAuthToBase64(backend.Auth{})
	assert.NoError(t, err)
	assert.EqualValues(t, "e30=", res)

	res, err = encodeAuthToBase64(backend.Auth{Username: "user", Password: "pwd", Email: "m@il.com"})
	assert.NoError(t, err)
	assert.EqualValues(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InB3ZCIsImVtYWlsIjoibUBpbC5jb20ifQ==", res)
}
