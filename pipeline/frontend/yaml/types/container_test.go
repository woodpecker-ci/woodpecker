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

package types

import (
	"testing"

	"github.com/docker/docker/api/types/strslice"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types/base"
)

var containerYaml = []byte(`
image: golang:latest
commands:
  - go build
  - go test
cpu_quota: 11
cpuset: 1,2
cpu_shares: 99
detach: true
devices:
  - /dev/ttyUSB0:/dev/ttyUSB0
directory: example/
dns: 8.8.8.8
dns_search: example.com
entrypoint: [/bin/sh, -c]
environment:
  RACK_ENV: development
  SHOW: true
extra_hosts:
 - somehost:162.242.195.82
 - otherhost:50.31.209.229
 - ipv6:2001:db8::10
name: my-build-container
network_mode: bridge
networks:
  - some-network
  - other-network
pull: true
privileged: true
shm_size: 1kb
mem_limit: 1kb
memswap_limit: 1kb
volumes:
  - /var/lib/mysql
  - /opt/data:/var/lib/mysql
  - /etc/configs:/etc/configs/:ro
tmpfs:
  - /var/lib/test
when:
  - branch: main
  - event: cron
    cron: job1
settings:
  foo: bar
  baz: false
ports:
  - 8080
  - 4443/tcp
  - 51820/udp
`)

func TestUnmarshalContainer(t *testing.T) {
	want := Container{
		Commands:     base.StringOrSlice{"go build", "go test"},
		CPUQuota:     base.StringOrInt(11),
		CPUSet:       "1,2",
		CPUShares:    base.StringOrInt(99),
		Detached:     true,
		Devices:      []string{"/dev/ttyUSB0:/dev/ttyUSB0"},
		Directory:    "example/",
		DNS:          base.StringOrSlice{"8.8.8.8"},
		DNSSearch:    base.StringOrSlice{"example.com"},
		Entrypoint:   []string{"/bin/sh", "-c"},
		Environment:  map[string]any{"RACK_ENV": "development", "SHOW": true},
		ExtraHosts:   []string{"somehost:162.242.195.82", "otherhost:50.31.209.229", "ipv6:2001:db8::10"},
		Image:        "golang:latest",
		MemLimit:     base.MemStringOrInt(1024),
		MemSwapLimit: base.MemStringOrInt(1024),
		Name:         "my-build-container",
		NetworkMode:  "bridge",
		Pull:         true,
		Privileged:   true,
		ShmSize:      base.MemStringOrInt(1024),
		Tmpfs:        base.StringOrSlice{"/var/lib/test"},
		Volumes: Volumes{
			Volumes: []*Volume{
				{Source: "", Destination: "/var/lib/mysql"},
				{Source: "/opt/data", Destination: "/var/lib/mysql"},
				{Source: "/etc/configs", Destination: "/etc/configs/", AccessMode: "ro"},
			},
		},
		When: constraint.When{
			Constraints: []constraint.Constraint{
				{
					Branch: constraint.List{
						Include: []string{"main"},
					},
				},
				{
					Event: base.StringOrSlice{"cron"},
					Cron: constraint.List{
						Include: []string{"job1"},
					},
				},
			},
		},
		Settings: map[string]any{
			"foo": "bar",
			"baz": false,
		},
		Ports: []string{"8080", "4443/tcp", "51820/udp"},
	}
	got := Container{}
	err := yaml.Unmarshal(containerYaml, &got)
	assert.NoError(t, err)
	assert.EqualValues(t, want, got, "problem parsing container")
}

// TestUnmarshalContainers unmarshals a map of containers. The order is
// retained and the container key may be used as the container name if a
// name is not explicitly provided.
func TestUnmarshalContainers(t *testing.T) {
	testdata := []struct {
		from string
		want []*Container
	}{
		{
			from: "build: { image: golang }",
			want: []*Container{
				{
					Name:  "build",
					Image: "golang",
				},
			},
		},
		{
			from: "test: { name: unit_test, image: node, settings: { normal_setting: true } }",
			want: []*Container{
				{
					Name:  "unit_test",
					Image: "node",
					Settings: map[string]any{
						"normal_setting": true,
					},
				},
			},
		},
		{
			from: `publish-agent:
    image: print/env
    settings:
      repo: woodpeckerci/woodpecker-agent
      dry_run: true
      dockerfile: docker/Dockerfile.agent
      tag: [next, latest]
    secrets: [docker_username, docker_password]
    when:
      branch: ${CI_REPO_DEFAULT_BRANCH}
      event: push`,
			want: []*Container{
				{
					Name:    "publish-agent",
					Image:   "print/env",
					Secrets: []string{"docker_username", "docker_password"},
					Settings: map[string]any{
						"repo":       "woodpeckerci/woodpecker-agent",
						"dockerfile": "docker/Dockerfile.agent",
						"tag":        stringsToInterface("next", "latest"),
						"dry_run":    true,
					},
					When: constraint.When{
						Constraints: []constraint.Constraint{
							{
								Event:  base.StringOrSlice{"push"},
								Branch: constraint.List{Include: []string{"${CI_REPO_DEFAULT_BRANCH}"}},
							},
						},
					},
				},
			},
		},
		{
			from: `publish-cli:
    image: print/env
    settings:
      repo: woodpeckerci/woodpecker-cli
      dockerfile: docker/Dockerfile.cli
      tag: [next]
    when:
      branch: ${CI_REPO_DEFAULT_BRANCH}
      event: push`,
			want: []*Container{
				{
					Name:  "publish-cli",
					Image: "print/env",
					Settings: map[string]any{
						"repo":       "woodpeckerci/woodpecker-cli",
						"dockerfile": "docker/Dockerfile.cli",
						"tag":        stringsToInterface("next"),
					},
					When: constraint.When{
						Constraints: []constraint.Constraint{
							{
								Event:  base.StringOrSlice{"push"},
								Branch: constraint.List{Include: []string{"${CI_REPO_DEFAULT_BRANCH}"}},
							},
						},
					},
				},
			},
		},
		{
			from: `publish-cli:
    image: print/env
    when:
      - branch: ${CI_REPO_DEFAULT_BRANCH}
        event: push
      - event: pull_request`,
			want: []*Container{
				{
					Name:  "publish-cli",
					Image: "print/env",
					When: constraint.When{
						Constraints: []constraint.Constraint{
							{
								Event:  base.StringOrSlice{"push"},
								Branch: constraint.List{Include: []string{"${CI_REPO_DEFAULT_BRANCH}"}},
							},
							{
								Event: base.StringOrSlice{"pull_request"},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range testdata {
		in := []byte(test.from)
		got := ContainerList{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.ContainerList, "problem parsing containers %q", test.from)
	}
}

// TestUnmarshalContainersErr unmarshals a container map where invalid inputs
// are provided to verify error messages are returned.
func TestUnmarshalContainersErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}
	for _, test := range testdata {
		in := []byte(test)
		containers := new(ContainerList)
		err := yaml.Unmarshal(in, &containers)
		assert.Error(t, err, "wanted error for containers %q", test)
	}
}

func stringsToInterface(val ...string) []any {
	res := make([]any, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}

func TestIsPlugin(t *testing.T) {
	assert.True(t, (&Container{}).IsPlugin())
	assert.True(t, (&Container{
		Commands: base.StringOrSlice(strslice.StrSlice{}),
	}).IsPlugin())
	assert.False(t, (&Container{
		Commands: base.StringOrSlice(strslice.StrSlice{"echo 'this is not a plugin'"}),
	}).IsPlugin())
	assert.True(t, (&Container{
		Entrypoint: base.StringOrSlice(strslice.StrSlice{}),
	}).IsPlugin())
	assert.False(t, (&Container{
		Entrypoint: base.StringOrSlice(strslice.StrSlice{"echo 'this is not a plugin'"}),
	}).IsPlugin())
}
