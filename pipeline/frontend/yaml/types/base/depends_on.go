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

package base

import (
	"errors"
	"fmt"
)

// Dependency represents a single dependency with an optional flag.
// When optional is true, the dependency is silently dropped if the
// referenced step or workflow is not present in the pipeline.
type Dependency struct {
	Name     string `yaml:"name"`
	Optional bool   `yaml:"optional,omitempty"`
}

// DependsOn represents a list of dependencies that can be unmarshalled from:
//   - a string: "step-a"
//   - a string array: ["step-a", "step-b"]
//   - an object array: [{name: "step-a", optional: true}]
//   - a mixed array: ["step-a", {name: "step-b", optional: true}]
type DependsOn []Dependency

// UnmarshalYAML implements the Unmarshaler interface.
func (d *DependsOn) UnmarshalYAML(unmarshal func(any) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		*d = DependsOn{{Name: stringType}}
		return nil
	}

	var sliceType []any
	if err := unmarshal(&sliceType); err == nil {
		deps := make(DependsOn, 0, len(sliceType))
		for _, item := range sliceType {
			switch v := item.(type) {
			case string:
				deps = append(deps, Dependency{Name: v})
			case map[string]any:
				dep, err := dependencyFromMap(v)
				if err != nil {
					return err
				}
				deps = append(deps, dep)
			default:
				return fmt.Errorf("cannot unmarshal '%v' of type %T into a dependency", item, item)
			}
		}
		*d = deps
		return nil
	}

	return errors.New("failed to unmarshal DependsOn")
}

func dependencyFromMap(m map[string]any) (Dependency, error) {
	dep := Dependency{}
	name, ok := m["name"]
	if !ok {
		return dep, fmt.Errorf("dependency object requires a 'name' field")
	}
	nameStr, ok := name.(string)
	if !ok {
		return dep, fmt.Errorf("dependency 'name' must be a string, got %T", name)
	}
	dep.Name = nameStr
	if opt, ok := m["optional"]; ok {
		optBool, ok := opt.(bool)
		if !ok {
			return dep, fmt.Errorf("dependency 'optional' must be a boolean, got %T", opt)
		}
		dep.Optional = optBool
	}
	return dep, nil
}

// MarshalYAML implements custom YAML marshaling.
// When all dependencies are non-optional, it marshals as a string (single)
// or string array (multiple) for backward compatibility.
// When any dependency is optional, it marshals as an object array.
func (d DependsOn) MarshalYAML() (any, error) {
	if len(d) == 0 {
		return nil, nil
	}

	hasOptional := false
	for _, dep := range d {
		if dep.Optional {
			hasOptional = true
			break
		}
	}

	if hasOptional {
		type depAlias Dependency
		out := make([]depAlias, len(d))
		for i, dep := range d {
			out[i] = depAlias(dep)
		}
		return out, nil
	}

	if len(d) == 1 {
		return d[0].Name, nil
	}
	names := make([]string, len(d))
	for i, dep := range d {
		names[i] = dep.Name
	}
	return names, nil
}

// Names returns all dependency names.
func (d DependsOn) Names() []string {
	if d == nil {
		return nil
	}
	names := make([]string, len(d))
	for i, dep := range d {
		names[i] = dep.Name
	}
	return names
}

// RequiredNames returns names of non-optional dependencies.
func (d DependsOn) RequiredNames() []string {
	if d == nil {
		return nil
	}
	var names []string
	for _, dep := range d {
		if !dep.Optional {
			names = append(names, dep.Name)
		}
	}
	return names
}

// OptionalNames returns names of optional dependencies.
func (d DependsOn) OptionalNames() []string {
	if d == nil {
		return nil
	}
	var names []string
	for _, dep := range d {
		if dep.Optional {
			names = append(names, dep.Name)
		}
	}
	return names
}
