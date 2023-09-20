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

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/utils"
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
		CPUQuota     base.StringOrInt    `yaml:"cpu_quota,omitempty"`
		CPUSet       string              `yaml:"cpuset,omitempty"`
		CPUShares    base.StringOrInt    `yaml:"cpu_shares,omitempty"`
		Devices      []string            `yaml:"devices,omitempty"`
		DNSSearch    base.StringOrSlice  `yaml:"dns_search,omitempty"`
		DNS          base.StringOrSlice  `yaml:"dns,omitempty"`
		ExtraHosts   []string            `yaml:"extra_hosts,omitempty"`
		IpcMode      string              `yaml:"ipc_mode,omitempty"`
		MemLimit     base.MemStringOrInt `yaml:"mem_limit,omitempty"`
		MemSwapLimit base.MemStringOrInt `yaml:"memswap_limit,omitempty"`
		NetworkMode  string              `yaml:"network_mode,omitempty"`
		Networks     Networks            `yaml:"networks,omitempty"`
		ShmSize      base.MemStringOrInt `yaml:"shm_size,omitempty"`
		Sysctls      base.SliceOrMap     `yaml:"sysctls,omitempty"`
		Tmpfs        []string            `yaml:"tmpfs,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (c *ContainerList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		if err := value.Decode(&c.ContainerList); err != nil {
			return err
		}

	case yaml.MappingNode:
		container := &Container{}
		if err := value.Decode(container); err != nil {
			return err
		}
		c.ContainerList = append(c.ContainerList, container)

	default:
		return fmt.Errorf("not supported yaml kind: %v", value.Kind)
	}

	return nil
}

func (c *Container) IsPlugin() bool {
	return len(c.Commands) == 0
}

func (c *Container) IsTrustedCloneImage() bool {
	return c.IsPlugin() && utils.MatchImage(c.Image, constant.TrustedCloneImages...)
}
