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
	"encoding/base64"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/stretchr/testify/assert"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
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
		UUID:        "f5182",
		Type:        backend.StepTypeCommands,
		Commands:    []string{"echo \"hello world\"", "ls"},
		Image:       "alpine",
		Environment: map[string]string{"SHELL": "/bin/zsh"},
	}

	testPluginStep = &backend.Step{
		Name:        "lint",
		UUID:        "d841e",
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
	assert.EqualValues(t, "wp_workflowNameTest-hello-f5182", toContainerName(testCmdStep, "workflowNameTest"))
	assert.EqualValues(t, "wp_workflowNameTest-lint-d841e", toContainerName(testPluginStep, "workflowNameTest"))
}

func TestStepToConfig(t *testing.T) {
	// StepTypeCommands
	conf := testEngine.toConfig(testCmdStep, BackendOptions{})
	if assert.NotNil(t, conf) {
		assert.EqualValues(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, conf.Entrypoint)
		assert.Nil(t, conf.Cmd)
		assert.EqualValues(t, testCmdStep.UUID, conf.Labels["wp_uuid"])
	}

	// StepTypePlugin
	conf = testEngine.toConfig(testPluginStep, BackendOptions{})
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
	engine := docker{info: system.Info{OSType: "linux", Architecture: "riscv64"}}

	conf := engine.toConfig(&backend.Step{
		Name:     "test",
		UUID:     "09238932",
		Commands: []string{"go test"},
	}, BackendOptions{})

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
			"CI_SCRIPT=CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVApta2RpciAtcCAiIgpjZCAiIgoKZWNobyArICdnbyB0ZXN0JwpnbyB0ZXN0Cg==",
			"SHELL=/bin/sh",
		},
	}, conf)
}

func TestToConfigFull(t *testing.T) {
	engine := docker{
		info: system.Info{OSType: "linux", Architecture: "riscv64"},
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
		Name:          "test",
		UUID:          "09238932",
		Type:          backend.StepTypeCommands,
		Image:         "golang:1.2.3",
		Pull:          true,
		Detached:      true,
		Privileged:    true,
		WorkingDir:    "/src/abc",
		WorkspaceBase: "/src",
		Environment:   map[string]string{"TAGS": "sqlite"},
		Commands:      []string{"go test", "go vet ./..."},
		ExtraHosts:    []backend.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:       []string{"/cache:/cache"},
		Tmpfs:         []string{"/tmp"},
		Devices:       []string{"/dev/sdc"},
		Networks:      []backend.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:           []string{"9.9.9.9", "8.8.8.8"},
		DNSSearch:     nil,
		OnFailure:     true,
		OnSuccess:     true,
		Failure:       "fail",
		AuthConfig:    backend.Auth{Username: "user", Password: "123456"},
		NetworkMode:   "bridge",
		Ports:         []backend.Port{{Number: 21}, {Number: 22}},
	}, BackendOptions{})

	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		Image:        "golang:1.2.3",
		WorkingDir:   "/src",
		AttachStdout: true,
		AttachStderr: true,
		Entrypoint:   []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"},
		Labels: map[string]string{
			"wp_step": "test",
			"wp_uuid": "09238932",
		},
		Env: []string{
			"CI_SCRIPT=CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVApta2RpciAtcCAiL3NyYy9hYmMiCmNkICIvc3JjL2FiYyIKCmVjaG8gKyAnZ28gdGVzdCcKZ28gdGVzdAoKZWNobyArICdnbyB2ZXQgLi8uLi4nCmdvIHZldCAuLy4uLgo=",
			"SHELL=/bin/sh",
			"TAGS=sqlite",
		},
		Volumes: map[string]struct{}{
			"/cache": {},
		},
	}, conf)
}

func TestToWindowsConfig(t *testing.T) {
	engine := docker{
		info: system.Info{OSType: "windows", Architecture: "x86_64"},
		config: config{
			enableIPv6: true,
		},
	}

	conf := engine.toConfig(&backend.Step{
		Name:          "test",
		UUID:          "23434553",
		Type:          backend.StepTypeCommands,
		Image:         "golang:1.2.3",
		WorkingDir:    "/src/abc",
		WorkspaceBase: "/src",
		Environment: map[string]string{
			"TAGS":         "sqlite",
			"CI_WORKSPACE": "/src",
		},
		Commands:    []string{"go test", "go vet ./..."},
		ExtraHosts:  []backend.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:     []string{"wp_default_abc:/src", "/cache:/cache/some/more", "test:/test"},
		Networks:    []backend.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:         []string{"9.9.9.9", "8.8.8.8"},
		Failure:     "fail",
		AuthConfig:  backend.Auth{Username: "user", Password: "123456"},
		NetworkMode: "nat",
		Ports:       []backend.Port{{Number: 21}, {Number: 22}},
	}, BackendOptions{})

	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		Image:        "golang:1.2.3",
		WorkingDir:   "C:/src",
		AttachStdout: true,
		AttachStderr: true,
		Entrypoint:   []string{"powershell", "-noprofile", "-noninteractive", "-command", "[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"},
		Labels: map[string]string{
			"wp_step": "test",
			"wp_uuid": "23434553",
		},
		Env: []string{
			"CI_SCRIPT=CiRFcnJvckFjdGlvblByZWZlcmVuY2UgPSAnU3RvcCc7CmlmICgtbm90IChUZXN0LVBhdGggIkM6L3NyYy9hYmMiKSkgeyBOZXctSXRlbSAtUGF0aCAiQzovc3JjL2FiYyIgLUl0ZW1UeXBlIERpcmVjdG9yeSAtRm9yY2UgfTsKaWYgKC1ub3QgW0Vudmlyb25tZW50XTo6R2V0RW52aXJvbm1lbnRWYXJpYWJsZSgnSE9NRScpKSB7IFtFbnZpcm9ubWVudF06OlNldEVudmlyb25tZW50VmFyaWFibGUoJ0hPTUUnLCAnYzpccm9vdCcpIH07CmlmICgtbm90IChUZXN0LVBhdGggIiRlbnY6SE9NRSIpKSB7IE5ldy1JdGVtIC1QYXRoICIkZW52OkhPTUUiIC1JdGVtVHlwZSBEaXJlY3RvcnkgLUZvcmNlIH07CmlmICgkRW52OkNJX05FVFJDX01BQ0hJTkUpIHsKJG5ldHJjPVtzdHJpbmddOjpGb3JtYXQoInswfVxfbmV0cmMiLCRFbnY6SE9NRSk7CiJtYWNoaW5lICRFbnY6Q0lfTkVUUkNfTUFDSElORSIgPj4gJG5ldHJjOwoibG9naW4gJEVudjpDSV9ORVRSQ19VU0VSTkFNRSIgPj4gJG5ldHJjOwoicGFzc3dvcmQgJEVudjpDSV9ORVRSQ19QQVNTV09SRCIgPj4gJG5ldHJjOwp9OwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9ORVRSQ19QQVNTV09SRCIsJG51bGwpOwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9TQ1JJUFQiLCRudWxsKTsKY2QgIkM6L3NyYy9hYmMiOwoKV3JpdGUtT3V0cHV0ICgnKyAiZ28gdGVzdCInKTsKJiBnbyB0ZXN0OyBpZiAoJExBU1RFWElUQ09ERSAtbmUgMCkge2V4aXQgJExBU1RFWElUQ09ERX0KCldyaXRlLU91dHB1dCAoJysgImdvIHZldCAuLy4uLiInKTsKJiBnbyB2ZXQgLi8uLi47IGlmICgkTEFTVEVYSVRDT0RFIC1uZSAwKSB7ZXhpdCAkTEFTVEVYSVRDT0RFfQo=",
			"CI_WORKSPACE=C:/src",
			"SHELL=powershell.exe",
			"TAGS=sqlite",
		},
		Volumes: map[string]struct{}{
			"C:/cache/some/more": {},
			"C:/src":             {},
			"C:/test":            {},
		},
	}, conf)

	ciScript, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(conf.Env[0], "CI_SCRIPT="))
	if assert.NoError(t, err) {
		assert.EqualValues(t, `
$ErrorActionPreference = 'Stop';
if (-not (Test-Path "C:/src/abc")) { New-Item -Path "C:/src/abc" -ItemType Directory -Force };
if (-not [Environment]::GetEnvironmentVariable('HOME')) { [Environment]::SetEnvironmentVariable('HOME', 'c:\root') };
if (-not (Test-Path "$env:HOME")) { New-Item -Path "$env:HOME" -ItemType Directory -Force };
if ($Env:CI_NETRC_MACHINE) {
$netrc=[string]::Format("{0}\_netrc",$Env:HOME);
"machine $Env:CI_NETRC_MACHINE" >> $netrc;
"login $Env:CI_NETRC_USERNAME" >> $netrc;
"password $Env:CI_NETRC_PASSWORD" >> $netrc;
};
[Environment]::SetEnvironmentVariable("CI_NETRC_PASSWORD",$null);
[Environment]::SetEnvironmentVariable("CI_SCRIPT",$null);
cd "C:/src/abc";

Write-Output ('+ "go test"');
& go test; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}

Write-Output ('+ "go vet ./..."');
& go vet ./...; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}
`, string(ciScript))
	}
}
