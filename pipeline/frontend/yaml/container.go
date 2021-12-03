package yaml

import (
	"fmt"

	libcompose "github.com/docker/libcompose/yaml"
	"gopkg.in/yaml.v3"
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
		AuthConfig    AuthConfig                `yaml:"auth_config,omitempty"`
		CapAdd        []string                  `yaml:"cap_add,omitempty"`
		CapDrop       []string                  `yaml:"cap_drop,omitempty"`
		Command       libcompose.Command        `yaml:"command,omitempty"`
		Commands      libcompose.Stringorslice  `yaml:"commands,omitempty"`
		CPUQuota      libcompose.StringorInt    `yaml:"cpu_quota,omitempty"`
		CPUSet        string                    `yaml:"cpuset,omitempty"`
		CPUShares     libcompose.StringorInt    `yaml:"cpu_shares,omitempty"`
		Detached      bool                      `yaml:"detach,omitempty"`
		Devices       []string                  `yaml:"devices,omitempty"`
		Tmpfs         []string                  `yaml:"tmpfs,omitempty"`
		DNS           libcompose.Stringorslice  `yaml:"dns,omitempty"`
		DNSSearch     libcompose.Stringorslice  `yaml:"dns_search,omitempty"`
		Entrypoint    libcompose.Command        `yaml:"entrypoint,omitempty"`
		Environment   libcompose.SliceorMap     `yaml:"environment,omitempty"`
		ExtraHosts    []string                  `yaml:"extra_hosts,omitempty"`
		Group         string                    `yaml:"group,omitempty"`
		Image         string                    `yaml:"image,omitempty"`
		Isolation     string                    `yaml:"isolation,omitempty"`
		Labels        libcompose.SliceorMap     `yaml:"labels,omitempty"`
		MemLimit      libcompose.MemStringorInt `yaml:"mem_limit,omitempty"`
		MemSwapLimit  libcompose.MemStringorInt `yaml:"memswap_limit,omitempty"`
		MemSwappiness libcompose.MemStringorInt `yaml:"mem_swappiness,omitempty"`
		Name          string                    `yaml:"name,omitempty"`
		NetworkMode   string                    `yaml:"network_mode,omitempty"`
		IpcMode       string                    `yaml:"ipc_mode,omitempty"`
		Networks      libcompose.Networks       `yaml:"networks,omitempty"`
		Privileged    bool                      `yaml:"privileged,omitempty"`
		Pull          bool                      `yaml:"pull,omitempty"`
		ShmSize       libcompose.MemStringorInt `yaml:"shm_size,omitempty"`
		Ulimits       libcompose.Ulimits        `yaml:"ulimits,omitempty"`
		Volumes       libcompose.Volumes        `yaml:"volumes,omitempty"`
		Secrets       Secrets                   `yaml:"secrets,omitempty"`
		Sysctls       libcompose.SliceorMap     `yaml:"sysctls,omitempty"`
		Constraints   Constraints               `yaml:"when,omitempty"`
		Settings      Settings                  `yaml:"settings"`
		Vargs         map[string]interface{}    `yaml:",inline"`
	}

	// Settings is a map of settings
	Settings struct {
		Params map[string]interface{} `yaml:",inline"`
	}
)

// UnmarshalYAML implements the Unmarshaller interface.
func (c *Containers) UnmarshalYAML(value *yaml.Node) error {
	containers := map[string]Container{}
	err := value.Decode(&containers)
	if err != nil {
		return err
	}

	for i, n := range value.Content {
		if i%2 == 1 {
			container := Container{}
			err := n.Decode(&container)
			if err != nil {
				return err
			}

			if container.Name == "" {
				container.Name = fmt.Sprintf("%v", value.Content[i-1].Value)
			}
			c.Containers = append(c.Containers, &container)
		}
	}

	// TODO: drop Vargs in favour of Settings in v1.16.0 release
	for _, cc := range c.Containers {
		if cc.Settings.Params == nil && cc.Vargs != nil {
			cc.Settings.Params = make(map[string]interface{})
		}
		for k, v := range cc.Vargs {
			cc.Settings.Params[k] = v
		}
	}
	return nil
}
