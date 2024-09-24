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
	"fmt"

	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types/base"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/utils"
)

type (
	// ContainerList denotes an ordered collection of containers.
	ContainerList struct {
		ContainerList []*Container
	}

	// Container defines a container.
	Container struct {
		BackendOptions map[string]any     `yaml:"backend_options,omitempty"`
		Commands       base.StringOrSlice `yaml:"commands,omitempty"`
		Entrypoint     base.StringOrSlice `yaml:"entrypoint,omitempty"`
		Detached       bool               `yaml:"detach,omitempty"`
		Directory      string             `yaml:"directory,omitempty"`
		Failure        string             `yaml:"failure,omitempty"`
		Image          string             `yaml:"image,omitempty"`
		Name           string             `yaml:"name,omitempty"`
		Pull           bool               `yaml:"pull,omitempty"`
		Settings       map[string]any     `yaml:"settings"`
		Volumes        Volumes            `yaml:"volumes,omitempty"`
		When           constraint.When    `yaml:"when,omitempty"`
		Ports          []string           `yaml:"ports,omitempty"`
		DependsOn      base.StringOrSlice `yaml:"depends_on,omitempty"`

		Secrets     []string       `yaml:"secrets,omitempty"`
		Environment map[string]any `yaml:"environment,omitempty"`

		// Docker and Kubernetes Specific
		Privileged bool `yaml:"privileged,omitempty"`

		// Undocumented
		CPUQuota     base.StringOrInt    `yaml:"cpu_quota,omitempty"`
		CPUSet       string              `yaml:"cpuset,omitempty"`
		CPUShares    base.StringOrInt    `yaml:"cpu_shares,omitempty"`
		Devices      []string            `yaml:"devices,omitempty"`
		DNSSearch    base.StringOrSlice  `yaml:"dns_search,omitempty"`
		DNS          base.StringOrSlice  `yaml:"dns,omitempty"`
		ExtraHosts   []string            `yaml:"extra_hosts,omitempty"`
		MemLimit     base.MemStringOrInt `yaml:"mem_limit,omitempty"`
		MemSwapLimit base.MemStringOrInt `yaml:"memswap_limit,omitempty"`
		NetworkMode  string              `yaml:"network_mode,omitempty"`
		ShmSize      base.MemStringOrInt `yaml:"shm_size,omitempty"`
		Tmpfs        []string            `yaml:"tmpfs,omitempty"`
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
	return len(c.Commands) == 0 &&
		len(c.Entrypoint) == 0 &&
		len(c.Environment) == 0 &&
		len(c.Secrets) == 0
}

func (c *Container) IsTrustedCloneImage(trustedClonePlugins []string) bool {
	return c.IsPlugin() && utils.MatchImageDynamic(c.Image, trustedClonePlugins...)
}
