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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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

func TestToConfigSmall(t *testing.T) {
	engine := docker{info: types.Info{OSType: "linux/riscv64"}}

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
		Cmd:          []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"},
		Entrypoint:   []string{"/bin/sh", "-c"},
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
	engine := docker{info: types.Info{OSType: "linux/riscv64"}}

	conf := engine.toConfig(&backend.Step{
		Name:         "test",
		UUID:         "09238932",
		Type:         backend.StepTypeCommands,
		Image:        "golang:1.2.3",
		Pull:         true,
		Detached:     true,
		Privileged:   true,
		WorkingDir:   "/src/abc",
		Environment:  map[string]string{"TAGS": "sqlite"},
		Commands:     []string{"go test", "go vet ./..."},
		ExtraHosts:   []backend.HostAlias{{Name: "t", IP: "1.2.3.4"}},
		Volumes:      []string{"/cache:/cache"},
		Tmpfs:        []string{"/tmp"},
		Devices:      []string{"/dev/sdc"},
		Networks:     []backend.Conn{{Name: "extra-net", Aliases: []string{"extra.net"}}},
		DNS:          []string{"9.9.9.9", "8.8.8.8"},
		DNSSearch:    nil,
		MemSwapLimit: 12,
		MemLimit:     13,
		ShmSize:      14,
		CPUQuota:     15,
		CPUShares:    16,
		OnFailure:    true,
		OnSuccess:    true,
		Failure:      "fail",
		AuthConfig:   backend.Auth{Username: "user", Password: "123456"},
		NetworkMode:  "bridge",
		Ports:        []backend.Port{{Number: 21}, {Number: 22}},
	})

	assert.NotNil(t, conf)
	sort.Strings(conf.Env)
	assert.EqualValues(t, &container.Config{
		Image:        "golang:1.2.3",
		WorkingDir:   "/src/abc",
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"},
		Entrypoint:   []string{"/bin/sh", "-c"},
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
