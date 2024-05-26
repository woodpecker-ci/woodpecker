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

import "gopkg.in/yaml.v3"

type (
	// Variables defines a collection of variables.
	Variables struct {
		Variables []*Variable
	}

	// Variable defines a container variable.
	Variable struct {
		Source string `yaml:"source"`
		Target string `yaml:"target"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (s *Variables) UnmarshalYAML(value *yaml.Node) error {
	y, _ := yaml.Marshal(value)

	var variables []string
	err := yaml.Unmarshal(y, &variables)
	if err == nil {
		for _, str := range variables {
			s.Variables = append(s.Variables, &Variable{
				Source: str,
				Target: str,
			})
		}
		return nil
	}
	return yaml.Unmarshal(y, &s.Variables)
}
