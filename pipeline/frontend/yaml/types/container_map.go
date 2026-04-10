// Copyright 2026 Woodpecker Authors
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

// ContainerMap contains collection of containers.
type ContainerMap struct {
	ContainerMap map[string]*Container
	Duplicated   []string
}

// UnmarshalYAML implements the Unmarshaler interface.
func (c *ContainerMap) UnmarshalYAML(value *yaml.Node) error {
	c.ContainerMap = make(map[string]*Container, len(value.Content)/2+1)
	switch value.Kind {
	// We support maps ...
	case yaml.MappingNode:
		for i, n := range value.Content {
			if i%2 == 1 {
				container := &Container{}
				if err := n.Decode(container); err != nil {
					return err
				}

				// service name is host name so we set it as name
				container.Name = fmt.Sprintf("%v", value.Content[i-1].Value)
				if container.Name == "" {
					return fmt.Errorf("container map does not allow empty key item")
				}

				c.ContainerMap[container.Name] = container
			}
		}

	// ... and lists
	case yaml.SequenceNode:
		for i, n := range value.Content {
			container := &Container{}
			if err := n.Decode(container); err != nil {
				return err
			}

			if container.Name == "" {
				container.Name = fmt.Sprintf("step-%d", i)
			}

			if _, exist := c.ContainerMap[container.Name]; exist {
				c.Duplicated = append(c.Duplicated, container.Name)
			} else {
				c.ContainerMap[container.Name] = container
			}
		}

	default:
		return fmt.Errorf("yaml node type[%d]: '%s' not supported", value.Kind, value.Tag)
	}

	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (c ContainerMap) MarshalYAML() (any, error) {
	// we just relay on the map key for names
	for _, v := range c.ContainerMap {
		v.Name = ""
	}
	return c.ContainerMap, nil
}
