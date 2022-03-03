package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// AuthConfig defines registry authentication credentials.
	AuthConfig struct {
		Username string
		Password string
		Email    string
	}

	// Containers denotes an ordered collection of containers.
	Containers struct {
		Containers []*Container
	}

	// Container defines a container.
	Container struct {
		AuthConfig    AuthConfig             `yaml:"auth_config,omitempty"`
		CapAdd        []string               `yaml:"cap_add,omitempty"`
		CapDrop       []string               `yaml:"cap_drop,omitempty"`
		Command       types.Command          `yaml:"command,omitempty"`
		Commands      types.Stringorslice    `yaml:"commands,omitempty"`
		CPUQuota      types.StringorInt      `yaml:"cpu_quota,omitempty"`
		CPUSet        string                 `yaml:"cpuset,omitempty"`
		CPUShares     types.StringorInt      `yaml:"cpu_shares,omitempty"`
		Detached      bool                   `yaml:"detach,omitempty"`
		Devices       []string               `yaml:"devices,omitempty"`
		Tmpfs         []string               `yaml:"tmpfs,omitempty"`
		DNS           types.Stringorslice    `yaml:"dns,omitempty"`
		DNSSearch     types.Stringorslice    `yaml:"dns_search,omitempty"`
		Entrypoint    types.Command          `yaml:"entrypoint,omitempty"`
		Environment   types.SliceorMap       `yaml:"environment,omitempty"`
		ExtraHosts    []string               `yaml:"extra_hosts,omitempty"`
		Group         string                 `yaml:"group,omitempty"`
		Image         string                 `yaml:"image,omitempty"`
		Isolation     string                 `yaml:"isolation,omitempty"`
		Labels        types.SliceorMap       `yaml:"labels,omitempty"`
		MemLimit      types.MemStringorInt   `yaml:"mem_limit,omitempty"`
		MemSwapLimit  types.MemStringorInt   `yaml:"memswap_limit,omitempty"`
		MemSwappiness types.MemStringorInt   `yaml:"mem_swappiness,omitempty"`
		Name          string                 `yaml:"name,omitempty"`
		NetworkMode   string                 `yaml:"network_mode,omitempty"`
		IpcMode       string                 `yaml:"ipc_mode,omitempty"`
		Networks      types.Networks         `yaml:"networks,omitempty"`
		Privileged    bool                   `yaml:"privileged,omitempty"`
		Pull          bool                   `yaml:"pull,omitempty"`
		ShmSize       types.MemStringorInt   `yaml:"shm_size,omitempty"`
		Ulimits       types.Ulimits          `yaml:"ulimits,omitempty"`
		Volumes       types.Volumes          `yaml:"volumes,omitempty"`
		Secrets       Secrets                `yaml:"secrets,omitempty"`
		Sysctls       types.SliceorMap       `yaml:"sysctls,omitempty"`
		Constraints   constraint.Constraints `yaml:"when,omitempty"`
		Settings      map[string]interface{} `yaml:"settings"`
		// Deprecated
		Vargs map[string]interface{} `yaml:",inline"` // TODO: remove deprecated with v0.16.0
	}
)

func (c *Container) MatchConstraints(meta frontend.Metadata) bool {
	return c.Constraints.Match(meta)
}

// UnmarshalYAML implements the Unmarshaler interface.
func (c *Containers) UnmarshalYAML(value *yaml.Node) error {
	// TODO: Deprecate pipeline as map to achieve proper yaml
	decodeFromMap := func() ([]*Container, error) {
		containers := []*Container{}
		// We cannot use decode on specific values
		// since if we try to load from a map, the order
		// will not be kept. Therefore use value.Content
		// and take the map values i%2=1
		for i, n := range value.Content {
			if i%2 == 1 {
				container := Container{}
				err := n.Decode(&container)
				if err != nil {
					return containers, err
				}

				if container.Name == "" {
					container.Name = fmt.Sprintf("%v", value.Content[i-1].Value)
				}

				containers = append(containers, &container)
			}
		}
		return containers, nil
	}

	decodeFromList := func() ([]*Container, error) {
		containers := []*Container{}
		err := value.Decode(&containers)
		if err != nil {
			return containers, err
		}
		for i, container := range containers {
			if container.Name == "" {
				container.Name = fmt.Sprintf("step-%d", i)
			}
		}
		return containers, nil
	}

	containers, err := decodeFromMap()
	if err != nil {
		containers, err = decodeFromList()
	}
	if err != nil {
		return err
	}

	// Load and validate the container values.
	c.Containers = containers

	// TODO drop Vargs in favor of Settings in v0.16.0 release
	for _, cc := range c.Containers {
		if cc.Settings == nil && cc.Vargs != nil {
			cc.Settings = make(map[string]interface{})
		}
		for k, v := range cc.Vargs {
			cc.Settings[k] = v
		}
		cc.Vargs = nil
	}

	return nil
}
