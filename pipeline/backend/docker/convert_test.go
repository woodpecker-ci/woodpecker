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
	"sort"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/stretchr/testify/assert"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
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
			assert.Equal(t, test.success, reflect.DeepEqual(results, test.to))
		}
	}
}

// dummy vars to test against.
var (
	testCmdStep = &backend.Step{
		Name:        "hello",
		UUID:        "f51821af-4cb8-435e-a3c2-3a684185d828",
		Type:        backend.StepTypeCommands,
		Commands:    []string{"echo \"hello world\"", "ls"},
		Image:       "alpine",
		Environment: map[string]string{"SHELL": "/bin/zsh"},
	}

	testPluginStep = &backend.Step{
		Name:        "lint",
		UUID:        "d841ee40-e66e-4275-bb3f-55bf89744b21",
		Type:        backend.StepTypePlugin,
		Image:       "mstruebing/editorconfig-checker",
		Environment: make(map[string]string),
	}

	testEngine = &docker{
		info: system.Info{
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
		assert.EqualValues(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, conf.Entrypoint)
		assert.Nil(t, conf.Cmd)
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

	res, err = encodeAuthToBase64(backend.Auth{Username: "user", Password: "pwd"})
	assert.NoError(t, err)
	assert.EqualValues(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InB3ZCJ9", res)
}

func TestToConfigSmall(t *testing.T) {
	engine := docker{info: system.Info{OSType: "linux/riscv64"}}

	conf := engine.toConfig(&backend.Step{
		Name:     "test",
		UUID:     "09238932",
		Commands: []string{"go test"},
	})

	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Entrypoint:   []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"},
		Labels: map[string]string{
			"wp_step": "test",
			"wp_uuid": "09238932",
		},
		Env: []string{
			"CI_SCRIPT=CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW" +
				"5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9" +
				"GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JE" +
				"CnVuc2V0IENJX1NDUklQVAoKZWNobyArICdnbyB0ZXN0JwpnbyB0ZXN0Cg==",
			"HOME=/root",
			"SHELL=/bin/sh",
		},
	}, conf)
}

func TestToConfigFull(t *testing.T) {
	engine := docker{
		info: system.Info{OSType: "linux/riscv64"},
		config: config{
			enableIPv6: true,
			resourceLimit: resourceLimit{
				MemSwapLimit: 12,
				MemLimit:     13,
				ShmSize:      14,
				CPUQuota:     15,
				CPUShares:    16,
			},
		},
	}

	conf := engine.toConfig(&backend.Step{
		Name:        "test",
		UUID:        "09238932",
		Type:        backend.StepTypeCommands,
		Image:       "golang:1.2.3",
		Pull:        true,
		Detached:    true,
		Privileged:  true,
		WorkingDir:  "/src/abc",
		Environment: map[string]string{"TAGS": "sqlite"},
		Commands:    []string{"go test", "go vet ./..."},
		ExtraHosts:  []backend.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:     []string{"/cache:/cache"},
		Tmpfs:       []string{"/tmp"},
		Devices:     []string{"/dev/sdc"},
		Networks:    []backend.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:         []string{"9.9.9.9", "8.8.8.8"},
		DNSSearch:   nil,
		OnFailure:   true,
		OnSuccess:   true,
		Failure:     "fail",
		AuthConfig:  backend.Auth{Username: "user", Password: "123456"},
		NetworkMode: "bridge",
		Ports:       []backend.Port{{Number: 21}, {Number: 22}},
	})

	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		Image:        "golang:1.2.3",
		WorkingDir:   "/src/abc",
		AttachStdout: true,
		AttachStderr: true,
		Entrypoint:   []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"},
		Labels: map[string]string{
			"wp_step": "test",
			"wp_uuid": "09238932",
		},
		Env: []string{
			"CI_SCRIPT=CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW" +
				"5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU" +
				"9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1" +
				"JECnVuc2V0IENJX1NDUklQVAoKZWNobyArICdnbyB0ZXN0JwpnbyB0ZXN0CgplY2hvICsgJ2dvIHZldCAuLy4uLicKZ28gdmV0IC" +
				"4vLi4uCg==",
			"HOME=/root",
			"SHELL=/bin/sh",
			"TAGS=sqlite",
		},
		Volumes: map[string]struct{}{
			"/cache": {},
		},
	}, conf)
}
