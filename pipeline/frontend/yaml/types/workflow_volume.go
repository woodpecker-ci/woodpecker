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
	// WorkflowVolumes defines a collection of volumes.
	WorkflowVolumes struct {
		WorkflowVolumes []*WorkflowVolume
	}

	// WorkflowVolume defines a container volume.
	WorkflowVolume struct {
		Name       string            `yaml:"name,omitempty"`
		Driver     string            `yaml:"driver,omitempty"`
		DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	}
)

// UnmarshalYAML implements the Unmarshaler interface.
func (v *WorkflowVolumes) UnmarshalYAML(value *yaml.Node) error {
	y, _ := yaml.Marshal(value)

	volumes := map[string]WorkflowVolume{}
	err := yaml.Unmarshal(y, &volumes)
	if err != nil {
		return err
	}

	for key, vv := range volumes {
		if vv.Name == "" {
			vv.Name = fmt.Sprintf("%v", key)
		}
		if vv.Driver == "" {
			vv.Driver = "local"
		}
		v.WorkflowVolumes = append(v.WorkflowVolumes, &vv)
	}
	return err
}
