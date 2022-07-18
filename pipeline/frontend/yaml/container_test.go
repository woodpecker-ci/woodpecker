package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

var containerYaml = []byte(`
image: golang:latest
auth_config:
  username: janedoe
  password: password
cap_add: [ ALL ]
cap_drop: [ NET_ADMIN, SYS_ADMIN ]
commands:
  - go build
  - go test
cpu_quota: 11
cpuset: 1,2
cpu_shares: 99
detach: true
devices:
  - /dev/ttyUSB0:/dev/ttyUSB0
dns: 8.8.8.8
dns_search: example.com
environment:
  - RACK_ENV=development
  - SHOW=true
extra_hosts:
 - somehost:162.242.195.82
 - otherhost:50.31.209.229
isolation: hyperv
name: my-build-container
network_mode: bridge
networks:
  - some-network
  - other-network
pull: true
privileged: true
labels:
  com.example.type: build
  com.example.team: frontend
shm_size: 1kb
mem_limit: 1kb
memswap_limit: 1kb
mem_swappiness: 1kb
volumes:
  - /var/lib/mysql
  - /opt/data:/var/lib/mysql
  - /etc/configs:/etc/configs/:ro
tmpfs:
  - /var/lib/test
when:
  branch: master
settings:
  foo: bar
  baz: false
`)

func TestUnmarshalContainer(t *testing.T) {
	want := Container{
		AuthConfig: AuthConfig{
			Username: "janedoe",
			Password: "password",
		},
		CapAdd:        []string{"ALL"},
		CapDrop:       []string{"NET_ADMIN", "SYS_ADMIN"},
		Commands:      types.Stringorslice{"go build", "go test"},
		CPUQuota:      types.StringorInt(11),
		CPUSet:        "1,2",
		CPUShares:     types.StringorInt(99),
		Detached:      true,
		Devices:       []string{"/dev/ttyUSB0:/dev/ttyUSB0"},
		DNS:           types.Stringorslice{"8.8.8.8"},
		DNSSearch:     types.Stringorslice{"example.com"},
		Environment:   types.SliceorMap{"RACK_ENV": "development", "SHOW": "true"},
		ExtraHosts:    []string{"somehost:162.242.195.82", "otherhost:50.31.209.229"},
		Image:         "golang:latest",
		Isolation:     "hyperv",
		Labels:        types.SliceorMap{"com.example.type": "build", "com.example.team": "frontend"},
		MemLimit:      types.MemStringorInt(1024),
		MemSwapLimit:  types.MemStringorInt(1024),
		MemSwappiness: types.MemStringorInt(1024),
		Name:          "my-build-container",
		Networks: types.Networks{
			Networks: []*types.Network{
				{Name: "some-network"},
				{Name: "other-network"},
			},
		},
		NetworkMode: "bridge",
		Pull:        true,
		Privileged:  true,
		ShmSize:     types.MemStringorInt(1024),
		Tmpfs:       types.Stringorslice{"/var/lib/test"},
		Volumes: types.Volumes{
			Volumes: []*types.Volume{
				{Source: "", Destination: "/var/lib/mysql"},
				{Source: "/opt/data", Destination: "/var/lib/mysql"},
				{Source: "/etc/configs", Destination: "/etc/configs/", AccessMode: "ro"},
			},
		},
		Constraints: constraint.Constraints{
			Branch: constraint.List{
				Include: []string{"master"},
			},
		},
		Settings: map[string]interface{}{
			"foo": "bar",
			"baz": false,
		},
	}
	got := Container{}
	err := yaml.Unmarshal(containerYaml, &got)
	assert.NoError(t, err)
	assert.EqualValues(t, want, got, "problem parsing container")
}

// TestUnmarshalContainersErr unmarshals a map of containers. The order is
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
					Settings: map[string]interface{}{
						"normal_setting": true,
					},
				},
			},
		},
		{
			from: `publish-agent:
    group: bundle
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
					Name:  "publish-agent",
					Image: "print/env",
					Group: "bundle",
					Secrets: Secrets{Secrets: []*Secret{{
						Source: "docker_username",
						Target: "docker_username",
					}, {
						Source: "docker_password",
						Target: "docker_password",
					}}},
					Settings: map[string]interface{}{
						"repo":       "woodpeckerci/woodpecker-agent",
						"dockerfile": "docker/Dockerfile.agent",
						"tag":        stringsToInterface("next", "latest"),
						"dry_run":    true,
					},
					Constraints: constraint.Constraints{
						Event:  constraint.List{Include: []string{"push"}},
						Branch: constraint.List{Include: []string{"${CI_REPO_DEFAULT_BRANCH}"}},
					},
				},
			},
		},
		{
			from: `publish-cli:
    group: docker
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
					Group: "docker",
					Settings: map[string]interface{}{
						"repo":       "woodpeckerci/woodpecker-cli",
						"dockerfile": "docker/Dockerfile.cli",
						"tag":        stringsToInterface("next"),
					},
					Constraints: constraint.Constraints{
						Event:  constraint.List{Include: []string{"push"}},
						Branch: constraint.List{Include: []string{"${CI_REPO_DEFAULT_BRANCH}"}},
					},
				},
			},
		},
	}
	for _, test := range testdata {
		in := []byte(test.from)
		got := Containers{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.Containers, "problem parsing containers %q", test.from)
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
		containers := new(Containers)
		err := yaml.Unmarshal(in, &containers)
		assert.Error(t, err, "wanted error for containers %q", test)
	}
}

func stringsToInterface(val ...string) []interface{} {
	res := make([]interface{}, len(val))
	for i := range val {
		res[i] = val[i]
	}
	return res
}
