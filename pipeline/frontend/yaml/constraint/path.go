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
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"

	yamlBaseTypes "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

// Path defines a runtime constrain for exclude & include paths.
type Path struct {
	Include       []string              `yaml:"include,omitempty"`
	Exclude       []string              `yaml:"exclude,omitempty"`
	IgnoreMessage string                `yaml:"ignore_message,omitempty"`
	OnEmpty       optional.Option[bool] `yaml:"on_empty,omitempty"`
}

// UnmarshalYAML unmarshal the constraint.
func (c *Path) UnmarshalYAML(value *yaml.Node) error {
	out1 := struct {
		Include       yamlBaseTypes.StringOrSlice `yaml:"include"`
		Exclude       yamlBaseTypes.StringOrSlice `yaml:"exclude"`
		IgnoreMessage string                      `yaml:"ignore_message"`
		OnEmpty       optional.Option[bool]       `yaml:"on_empty"`
	}{}

	var out2 yamlBaseTypes.StringOrSlice

	err1 := value.Decode(&out1)
	err2 := value.Decode(&out2)

	c.Exclude = out1.Exclude
	c.IgnoreMessage = out1.IgnoreMessage
	c.OnEmpty = out1.OnEmpty
	c.Include = append( //nolint:gocritic
		out1.Include,
		out2...,
	)

	if err1 != nil && err2 != nil {
		y, _ := yaml.Marshal(value)
		return fmt.Errorf("could not parse condition: %s", y)
	}

	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (c Path) MarshalYAML() (any, error) {
	// if only Include is set return simple syntax
	if len(c.Exclude) == 0 &&
		len(c.IgnoreMessage) == 0 &&
		c.OnEmpty.ValueOrDefault(true) {
		if len(c.Include) == 0 {
			return nil, nil
		}
		return yamlBaseTypes.StringOrSlice(c.Include), nil
	}

	// clean up on_empty if true make it none as we will default to true
	if c.OnEmpty.ValueOrDefault(true) {
		c.OnEmpty = optional.None[bool]()
	}

	// we can not return type Path as it would lead to infinite recursion :/
	return struct {
		Include       yamlBaseTypes.StringOrSlice `yaml:"include,omitempty"`
		Exclude       yamlBaseTypes.StringOrSlice `yaml:"exclude,omitempty"`
		IgnoreMessage string                      `yaml:"ignore_message,omitempty"`
		OnEmpty       optional.Option[bool]       `yaml:"on_empty,omitempty"`
	}{
		Include:       c.Include,
		Exclude:       c.Exclude,
		IgnoreMessage: c.IgnoreMessage,
		OnEmpty:       c.OnEmpty,
	}, nil
}

// Match returns true if file paths in string slice matches the include and not exclude patterns
// or if commit message contains ignore message.
func (c *Path) Match(v []string, message string) bool {
	// ignore file pattern matches if the commit message contains a pattern
	if len(c.IgnoreMessage) > 0 && strings.Contains(strings.ToLower(message), strings.ToLower(c.IgnoreMessage)) {
		return true
	}

	// return value based on 'on_empty', if there are no commit files (empty commit)
	if len(v) == 0 {
		return c.OnEmpty.ValueOrDefault(true)
	}

	if len(c.Exclude) > 0 && c.Excludes(v) {
		return false
	}
	if len(c.Include) > 0 && !c.Includes(v) {
		return false
	}
	return true
}

// Includes returns true if the string matches any of the include patterns.
func (c *Path) Includes(v []string) bool {
	for _, pattern := range c.Include {
		for _, file := range v {
			if ok, _ := doublestar.Match(pattern, file); ok {
				return true
			}
		}
	}
	return false
}

// Excludes returns true if all of the strings match any of the exclude patterns.
func (c *Path) Excludes(v []string) bool {
	for _, file := range v {
		matched := false
		for _, pattern := range c.Exclude {
			if ok, _ := doublestar.Match(pattern, file); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
