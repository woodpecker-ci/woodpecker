package types

import (
	"fmt"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

type (
	// ContainerList denotes an ordered collection of containers.
	ContainerList struct {
		ContainerList []*Container
	}

	// Container defines a container.
	Container struct {
		BackendOptions BackendOptions         `yaml:"backend_options,omitempty"`
		Commands       base.StringOrSlice     `yaml:"commands,omitempty"`
		Detached       bool                   `yaml:"detach,omitempty"`
		Directory      string                 `yaml:"directory,omitempty"`
		Environment    base.SliceOrMap        `yaml:"environment,omitempty"`
		Failure        string                 `yaml:"failure,omitempty"`
		Group          string                 `yaml:"group,omitempty"`
		Image          string                 `yaml:"image,omitempty"`
		Name           string                 `yaml:"name,omitempty"`
		Pull           bool                   `yaml:"pull,omitempty"`
		Secrets        Secrets                `yaml:"secrets,omitempty"`
		Settings       map[string]interface{} `yaml:"settings"`
		Volumes        Volumes                `yaml:"volumes,omitempty"`
		When           constraint.When        `yaml:"when,omitempty"`

		// Docker Specific
		Privileged bool `yaml:"privileged,omitempty"`

		// Undocumented
		CapAdd        []string            `yaml:"cap_add,omitempty"`
		CapDrop       []string            `yaml:"cap_drop,omitempty"`
		CPUQuota      base.StringOrInt    `yaml:"cpu_quota,omitempty"`
		CPUSet        string              `yaml:"cpuset,omitempty"`
		CPUShares     base.StringOrInt    `yaml:"cpu_shares,omitempty"`
		Devices       []string            `yaml:"devices,omitempty"`
		DNSSearch     base.StringOrSlice  `yaml:"dns_search,omitempty"`
		DNS           base.StringOrSlice  `yaml:"dns,omitempty"`
		ExtraHosts    []string            `yaml:"extra_hosts,omitempty"`
		IpcMode       string              `yaml:"ipc_mode,omitempty"`
		Isolation     string              `yaml:"isolation,omitempty"`
		MemLimit      base.MemStringOrInt `yaml:"mem_limit,omitempty"`
		MemSwapLimit  base.MemStringOrInt `yaml:"memswap_limit,omitempty"`
		MemSwappiness base.MemStringOrInt `yaml:"mem_swappiness,omitempty"`
		NetworkMode   string              `yaml:"network_mode,omitempty"`
		Networks      Networks            `yaml:"networks,omitempty"`
		ShmSize       base.MemStringOrInt `yaml:"shm_size,omitempty"`
		Sysctls       base.SliceOrMap     `yaml:"sysctls,omitempty"`
		Tmpfs         []string            `yaml:"tmpfs,omitempty"`
		Ulimits       Ulimits             `yaml:"ulimits,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (c *ContainerList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	// We support maps ...
	case yaml.MappingNode:
		c.ContainerList = make([]*Container, 0, len(value.Content)/2+1)
		// We cannot use decode on specific values
		// since if we try to load from a map, the order
		// will not be kept. Therefor use value.Content
		// and take the map values i%2=1
		for i, n := range value.Content {
			if i%2 == 1 {
				container := &Container{}
				if err := n.Decode(container); err != nil {
					return err
				}

				if container.Name == "" {
					container.Name = fmt.Sprintf("%v", value.Content[i-1].Value)
				}

				c.ContainerList = append(c.ContainerList, container)
			}
		}

	// ... and lists
	case yaml.SequenceNode:
		c.ContainerList = make([]*Container, 0, len(value.Content))
		for i, n := range value.Content {
			container := &Container{}
			if err := n.Decode(container); err != nil {
				return err
			}

			if container.Name == "" {
				container.Name = fmt.Sprintf("step-%d", i)
			}

			c.ContainerList = append(c.ContainerList, container)
		}

	default:
		return fmt.Errorf("yaml node type[%d]: '%s' not supported", value.Kind, value.Tag)
	}

	return nil
}

func (c *Container) IsPlugin() bool {
	return len(c.Commands) == 0
}

func (c *Container) IsTrustedCloneImage() bool {
	return c.IsPlugin() && slices.Contains(constant.TrustedCloneImages, c.Image)
}
