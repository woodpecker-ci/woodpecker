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
)

type (
	// WorkflowNetworks defines a collection of networks.
	WorkflowNetworks struct {
		WorkflowNetworks []*WorkflowNetwork
	}

	// WorkflowNetwork defines a container network.
	WorkflowNetwork struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (n *WorkflowNetworks) UnmarshalYAML(value *yaml.Node) error {
	networks := map[string]WorkflowNetwork{}
	err := value.Decode(&networks)

	for key, nn := range networks {
		if nn.Name == "" {
			nn.Name = fmt.Sprintf("%v", key)
		}
		if nn.Driver == "" {
			nn.Driver = "bridge"
		}
		n.WorkflowNetworks = append(n.WorkflowNetworks, &nn)
	}
	return err
}
