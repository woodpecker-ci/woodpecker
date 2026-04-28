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
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/utils"
)

// Container defines a container.
type Container struct {
	// common
	Name        string             `yaml:"name,omitempty"`
	Image       string             `yaml:"image,omitempty"`
	Pull        bool               `yaml:"pull,omitempty"`
	Commands    base.StringOrSlice `yaml:"commands,omitempty"`
	Entrypoint  base.StringOrSlice `yaml:"entrypoint,omitempty"`
	Directory   string             `yaml:"directory,omitempty"`
	Settings    map[string]any     `yaml:"settings,omitempty"`
	Environment map[string]any     `yaml:"environment,omitempty"`
	// flow control
	DependsOn base.StringOrSlice `yaml:"depends_on,omitempty"`
	When      constraint.When    `yaml:"when,omitempty"`
	Failure   string             `yaml:"failure,omitempty"`
	Detached  bool               `yaml:"detach,omitempty"`
	// state
	Volumes Volumes `yaml:"volumes,omitempty"`
	// network
	Ports     []string           `yaml:"ports,omitempty"`
	DNS       base.StringOrSlice `yaml:"dns,omitempty"`
	DNSSearch base.StringOrSlice `yaml:"dns_search,omitempty"`
	// backend specific
	BackendOptions map[string]any `yaml:"backend_options,omitempty"`

	// ACTIVE DEVELOPMENT BELOW

	// Docker and Kubernetes Specific
	Privileged bool `yaml:"privileged,omitempty"`

	// Undocumented
	Devices     []string `yaml:"devices,omitempty"`
	ExtraHosts  []string `yaml:"extra_hosts,omitempty"`
	NetworkMode string   `yaml:"network_mode,omitempty"`
	Tmpfs       []string `yaml:"tmpfs,omitempty"`
}

func (c *Container) IsPlugin() bool {
	return len(c.Commands) == 0 &&
		len(c.Entrypoint) == 0 &&
		len(c.Environment) == 0
}

func (c *Container) IsTrustedCloneImage(trustedClonePlugins []string) bool {
	return c.IsPlugin() && utils.MatchImageDynamic(c.Image, trustedClonePlugins...)
}
