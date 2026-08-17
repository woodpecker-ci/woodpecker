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

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
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
	testCmdStep = &backend_types.Step{
		Name:        "hello",
		UUID:        "f51821af-4cb8-435e-a3c2-3a684185d828",
		Type:        backend_types.StepTypeCommands,
		Commands:    []string{"echo \"hello world\"", "ls"},
		Image:       "alpine",
		Environment: map[string]string{"SHELL": "/bin/zsh"},
	}

	testPluginStep = &backend_types.Step{
		Name:        "lint",
		UUID:        "d841ee40-e66e-4275-bb3f-55bf89744b21",
		Type:        backend_types.StepTypePlugin,
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

func TestToHostConfigApparmorProfile(t *testing.T) {
	hostConfig, err := toHostConfig(testCmdStep, &config{apparmor: "osgeo-woodie"})

	assert.NoError(t, err)
	assert.EqualValues(t, []string{"apparmor=osgeo-woodie"}, hostConfig.SecurityOpt)
}

func TestToHostConfigApparmorProfileDefault(t *testing.T) {
	hostConfig, err := toHostConfig(testCmdStep, &config{})

	assert.NoError(t, err)
	assert.Nil(t, hostConfig.SecurityOpt)
}

func TestStepToConfig(t *testing.T) {
	// StepTypeCommands
	conf, err := testEngine.toConfig(testCmdStep, BackendOptions{})
	require.NoError(t, err)
	if assert.NotNil(t, conf) {
		assert.EqualValues(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, conf.Entrypoint)
		assert.Nil(t, conf.Cmd)
		assert.EqualValues(t, testCmdStep.UUID, conf.Labels["wp_uuid"])
	}

	// StepTypePlugin
	conf, err = testEngine.toConfig(testPluginStep, BackendOptions{})
	require.NoError(t, err)
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
	res, err := encodeAuthToBase64(backend_types.Auth{})
	assert.NoError(t, err)
	assert.EqualValues(t, "e30=", res)

	res, err = encodeAuthToBase64(backend_types.Auth{Username: "user", Password: "pwd"})
	assert.NoError(t, err)
	assert.EqualValues(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InB3ZCJ9", res)
}

func TestToConfigSmall(t *testing.T) {
	engine := docker{info: system.Info{OSType: "linux", Architecture: "riscv64"}}

	conf, err := engine.toConfig(&backend_types.Step{
		Name:     "test",
		UUID:     "09238932",
		Commands: []string{"go test"},
	}, BackendOptions{})

	require.NoError(t, err)
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

	conf, err := engine.toConfig(&backend_types.Step{
		Name:          "test",
		UUID:          "09238932",
		Type:          backend_types.StepTypeCommands,
		Image:         "golang:1.2.3",
		Pull:          true,
		Detached:      true,
		Privileged:    true,
		WorkingDir:    "/src/abc",
		WorkspaceBase: "/src",
		Environment:   map[string]string{"TAGS": "sqlite"},
		Commands:      []string{"go test", "go vet ./..."},
		ExtraHosts:    []backend_types.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:       []string{"/cache:/cache"},
		Tmpfs:         []string{"/tmp"},
		Devices:       []string{"/dev/sdc"},
		Networks:      []backend_types.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:           []string{"9.9.9.9", "8.8.8.8"},
		DNSSearch:     nil,
		OnFailure:     true,
		OnSuccess:     true,
		Failure:       "fail",
		AuthConfig:    backend_types.Auth{Username: "user", Password: "123456"},
		NetworkMode:   "bridge",
		Ports:         []backend_types.Port{{Number: 21}, {Number: 22}},
	}, BackendOptions{})

	require.NoError(t, err)
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

// windowsCIScriptBase64 is the UTF-16LE encoded script, base64 encoded for powershell's -encodedcommand.
const windowsCIScriptBase64 = "CgAkAEwAQQBTAFQARQBYAEkAVABDAE8ARABFACAAPQAgADAACgAkAEUAcgByAG8AcgBBAGMAdABpAG8AbgBQAHIAZQBmAGUAcgBlAG4AYwBlACAAPQAgACcAUwB0AG8AcAAnADsACgBpAGYAIAAoAC0AbgBvAHQAIAAoAFQAZQBzAHQALQBQAGEAdABoACAAIgBDADoALwBzAHIAYwAvAGEAYgBjACIAKQApACAAewAgAE4AZQB3AC0ASQB0AGUAbQAgAC0AUABhAHQAaAAgACIAQwA6AC8AcwByAGMALwBhAGIAYwAiACAALQBJAHQAZQBtAFQAeQBwAGUAIABEAGkAcgBlAGMAdABvAHIAeQAgAC0ARgBvAHIAYwBlACAAfQA7AAoAaQBmACAAKAAtAG4AbwB0ACAAWwBFAG4AdgBpAHIAbwBuAG0AZQBuAHQAXQA6ADoARwBlAHQARQBuAHYAaQByAG8AbgBtAGUAbgB0AFYAYQByAGkAYQBiAGwAZQAoACcASABPAE0ARQAnACkAKQAgAHsAIABbAEUAbgB2AGkAcgBvAG4AbQBlAG4AdABdADoAOgBTAGUAdABFAG4AdgBpAHIAbwBuAG0AZQBuAHQAVgBhAHIAaQBhAGIAbABlACgAJwBIAE8ATQBFACcALAAgACcAYwA6AFwAcgBvAG8AdAAnACkAIAB9ADsACgBpAGYAIAAoAC0AbgBvAHQAIAAoAFQAZQBzAHQALQBQAGEAdABoACAAIgAkAGUAbgB2ADoASABPAE0ARQAiACkAKQAgAHsAIABOAGUAdwAtAEkAdABlAG0AIAAtAFAAYQB0AGgAIAAiACQAZQBuAHYAOgBIAE8ATQBFACIAIAAtAEkAdABlAG0AVAB5AHAAZQAgAEQAaQByAGUAYwB0AG8AcgB5ACAALQBGAG8AcgBjAGUAIAB9ADsACgBpAGYAIAAoACQARQBuAHYAOgBDAEkAXwBOAEUAVABSAEMAXwBNAEEAQwBIAEkATgBFACkAIAB7AAoAJABuAGUAdAByAGMAPQBbAHMAdAByAGkAbgBnAF0AOgA6AEYAbwByAG0AYQB0ACgAIgB7ADAAfQBcAF8AbgBlAHQAcgBjACIALAAkAEUAbgB2ADoASABPAE0ARQApADsACgAiAG0AYQBjAGgAaQBuAGUAIAAkAEUAbgB2ADoAQwBJAF8ATgBFAFQAUgBDAF8ATQBBAEMASABJAE4ARQAiACAAPgA+ACAAJABuAGUAdAByAGMAOwAKACIAbABvAGcAaQBuACAAJABFAG4AdgA6AEMASQBfAE4ARQBUAFIAQwBfAFUAUwBFAFIATgBBAE0ARQAiACAAPgA+ACAAJABuAGUAdAByAGMAOwAKACIAcABhAHMAcwB3AG8AcgBkACAAJABFAG4AdgA6AEMASQBfAE4ARQBUAFIAQwBfAFAAQQBTAFMAVwBPAFIARAAiACAAPgA+ACAAJABuAGUAdAByAGMAOwAKAH0AOwAKAFsARQBuAHYAaQByAG8AbgBtAGUAbgB0AF0AOgA6AFMAZQB0AEUAbgB2AGkAcgBvAG4AbQBlAG4AdABWAGEAcgBpAGEAYgBsAGUAKAAiAEMASQBfAE4ARQBUAFIAQwBfAFAAQQBTAFMAVwBPAFIARAAiACwAJABuAHUAbABsACkAOwAKAFsARQBuAHYAaQByAG8AbgBtAGUAbgB0AF0AOgA6AFMAZQB0AEUAbgB2AGkAcgBvAG4AbQBlAG4AdABWAGEAcgBpAGEAYgBsAGUAKAAiAEMASQBfAFMAQwBSAEkAUABUACIALAAkAG4AdQBsAGwAKQA7AAoAYwBkACAAIgBDADoALwBzAHIAYwAvAGEAYgBjACIAOwAKAAoAVwByAGkAdABlAC0ATwB1AHQAcAB1AHQAIAAoACcAKwAgACIAZwBvACAAdABlAHMAdAAiACcAKQA7AAoAJgAgAGcAbwAgAHQAZQBzAHQAOwAgAGkAZgAgACgAJABMAEEAUwBUAEUAWABJAFQAQwBPAEQARQAgAC0AbgBlACAAMAApACAAewBlAHgAaQB0ACAAJABMAEEAUwBUAEUAWABJAFQAQwBPAEQARQB9AAoACgBXAHIAaQB0AGUALQBPAHUAdABwAHUAdAAgACgAJwArACAAIgBnAG8AIAB2AGUAdAAgAC4ALwAuAC4ALgAiACcAKQA7AAoAJgAgAGcAbwAgAHYAZQB0ACAALgAvAC4ALgAuADsAIABpAGYAIAAoACQATABBAFMAVABFAFgASQBUAEMATwBEAEUAIAAtAG4AZQAgADAAKQAgAHsAZQB4AGkAdAAgACQATABBAFMAVABFAFgASQBUAEMATwBEAEUAfQAKAA=="

func TestToWindowsConfig(t *testing.T) {
	engine := docker{
		info: system.Info{OSType: "windows", Architecture: "x86_64"},
		config: config{
			enableIPv6: true,
		},
	}

	conf, err := engine.toConfig(&backend_types.Step{
		Name:          "test",
		UUID:          "23434553",
		Type:          backend_types.StepTypeCommands,
		Image:         "golang:1.2.3",
		WorkingDir:    "/src/abc",
		WorkspaceBase: "/src",
		Environment: map[string]string{
			"TAGS":         "sqlite",
			"CI_WORKSPACE": "/src",
		},
		Commands:    []string{"go test", "go vet ./..."},
		ExtraHosts:  []backend_types.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:     []string{"wp_default_abc:/src", "/cache:/cache/some/more", "test:/test"},
		Networks:    []backend_types.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:         []string{"9.9.9.9", "8.8.8.8"},
		Failure:     "fail",
		AuthConfig:  backend_types.Auth{Username: "user", Password: "123456"},
		NetworkMode: "nat",
		Ports:       []backend_types.Port{{Number: 21}, {Number: 22}},
	}, BackendOptions{})

	require.NoError(t, err)
	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		Image:        "golang:1.2.3",
		WorkingDir:   "C:/src",
		AttachStdout: true,
		AttachStderr: true,
		Entrypoint:   []string{"powershell", "-noprofile", "-noninteractive", "-encodedcommand", windowsCIScriptBase64},
		Labels: map[string]string{
			"wp_step": "test",
			"wp_uuid": "23434553",
		},
		Env: []string{
			"CI_SCRIPT=" + windowsCIScriptBase64,
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

	utf16leScript, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(conf.Env[0], "CI_SCRIPT="))
	require.NoError(t, err)
	ciScript, err := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Bytes(utf16leScript)
	if assert.NoError(t, err) {
		assert.EqualValues(t, `
$LASTEXITCODE = 0
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
