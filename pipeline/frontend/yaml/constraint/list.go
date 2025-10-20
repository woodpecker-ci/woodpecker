// Copyright 2025 Woodpecker Authors
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

package constraint

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"

	yamlBaseTypes "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
)

// List defines a runtime constraint for exclude & include string slices.
type List struct {
	Include []string
	Exclude []string
}

// IsEmpty return true if a constraint has no conditions.
func (c List) IsEmpty() bool {
	return len(c.Include) == 0 && len(c.Exclude) == 0
}

// Match returns true if the string matches the include patterns and does not
// match any of the exclude patterns.
func (c *List) Match(v string) bool {
	if c == nil {
		return true
	}
	if c.Excludes(v) {
		return false
	}
	if c.Includes(v) {
		return true
	}
	if len(c.Include) == 0 {
		return true
	}
	return false
}

// Includes returns true if the string matches the include patterns.
func (c *List) Includes(v string) bool {
	for _, pattern := range c.Include {
		if ok, _ := doublestar.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// Excludes returns true if the string matches the exclude patterns.
func (c *List) Excludes(v string) bool {
	for _, pattern := range c.Exclude {
		if ok, _ := doublestar.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// UnmarshalYAML unmarshal the constraint.
func (c *List) UnmarshalYAML(value *yaml.Node) error {
	out1 := struct {
		Include yamlBaseTypes.StringOrSlice
		Exclude yamlBaseTypes.StringOrSlice
	}{}

	var out2 yamlBaseTypes.StringOrSlice

	err1 := value.Decode(&out1)
	err2 := value.Decode(&out2)

	c.Exclude = out1.Exclude
	c.Include = append( //nolint:gocritic
		out1.Include,
		out2...,
	)

	if err1 != nil && err2 != nil {
		y, _ := yaml.Marshal(value)
		return fmt.Errorf("could not parse condition: %s: %w", y, multierr.Append(err1, err2))
	}

	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (c List) MarshalYAML() (any, error) {
	switch {
	case len(c.Include) == 0 && len(c.Exclude) == 0:
		return nil, nil
	case len(c.Exclude) == 0:
		return yamlBaseTypes.StringOrSlice(c.Include), nil
	default:
		// we can not return type List as it would lead to infinite recursion :/
		return struct {
			Include yamlBaseTypes.StringOrSlice `yaml:"include,omitempty"`
			Exclude yamlBaseTypes.StringOrSlice `yaml:"exclude,omitempty"`
		}{
			Include: c.Include,
			Exclude: c.Exclude,
		}, nil
	}
}
