package yaml

import (
	"reflect"
	"testing"

	libcompose "github.com/docker/libcompose/yaml"
	"github.com/kr/pretty"
	"gopkg.in/yaml.v3"
)

var containerYaml = []byte(`
image: golang:latest
auth_config:
  username: janedoe
  password: password
cap_add: [ ALL ]
cap_drop: [ NET_ADMIN, SYS_ADMIN ]
command: bundle exec thin -p 3000
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
entrypoint: /code/entrypoint.sh
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
`)

func TestUnmarshalContainer(t *testing.T) {
	want := Container{
		AuthConfig: AuthConfig{
			Username: "janedoe",
			Password: "password",
		},
		CapAdd:        []string{"ALL"},
		CapDrop:       []string{"NET_ADMIN", "SYS_ADMIN"},
		Command:       libcompose.Command{"bundle", "exec", "thin", "-p", "3000"},
		Commands:      libcompose.Stringorslice{"go build", "go test"},
		CPUQuota:      libcompose.StringorInt(11),
		CPUSet:        "1,2",
		CPUShares:     libcompose.StringorInt(99),
		Detached:      true,
		Devices:       []string{"/dev/ttyUSB0:/dev/ttyUSB0"},
		DNS:           libcompose.Stringorslice{"8.8.8.8"},
		DNSSearch:     libcompose.Stringorslice{"example.com"},
		Entrypoint:    libcompose.Command{"/code/entrypoint.sh"},
		Environment:   libcompose.SliceorMap{"RACK_ENV": "development", "SHOW": "true"},
		ExtraHosts:    []string{"somehost:162.242.195.82", "otherhost:50.31.209.229"},
		Image:         "golang:latest",
		Isolation:     "hyperv",
		Labels:        libcompose.SliceorMap{"com.example.type": "build", "com.example.team": "frontend"},
		MemLimit:      libcompose.MemStringorInt(1024),
		MemSwapLimit:  libcompose.MemStringorInt(1024),
		MemSwappiness: libcompose.MemStringorInt(1024),
		Name:          "my-build-container",
		Networks: libcompose.Networks{
			Networks: []*libcompose.Network{
				{Name: "some-network"},
				{Name: "other-network"},
			},
		},
		NetworkMode: "bridge",
		Pull:        true,
		Privileged:  true,
		ShmSize:     libcompose.MemStringorInt(1024),
		Tmpfs:       libcompose.Stringorslice{"/var/lib/test"},
		Volumes: libcompose.Volumes{
			Volumes: []*libcompose.Volume{
				{Source: "", Destination: "/var/lib/mysql"},
				{Source: "/opt/data", Destination: "/var/lib/mysql"},
				{Source: "/etc/configs", Destination: "/etc/configs/", AccessMode: "ro"},
			},
		},
		Constraints: Constraints{
			Branch: Constraint{
				Include: []string{"master"},
			},
		},
	}
	got := Container{}
	err := yaml.Unmarshal(containerYaml, &got)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(want, got) {
		t.Errorf("problem parsing container")
		pretty.Ldiff(t, want, got)
	}
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
			from: "test: { name: unit_test, image: node }",
			want: []*Container{
				{
					Name:  "unit_test",
					Image: "node",
				},
			},
		},
	}
	for _, test := range testdata {
		in := []byte(test.from)
		got := Containers{}
		err := yaml.Unmarshal(in, &got)
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(test.want, got.Containers) {
			t.Errorf("problem parsing containers %q", test.from)
			pretty.Ldiff(t, test.want, got.Containers)
		}
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
		if err == nil {
			t.Errorf("wanted error for containers %q", test)
		}
	}
}
