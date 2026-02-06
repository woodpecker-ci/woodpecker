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

import "github.com/bmatcuk/doublestar/v4"

// Map defines a runtime constraint for exclude & include map strings.
type Map struct {
	Include map[string]string `yaml:"include,omitempty"`
	Exclude map[string]string `yaml:"exclude,omitempty"`
}

// Match returns true if the params matches the include key values and does not
// match any of the exclude key values.
func (c *Map) Match(params map[string]string) bool {
	// when no includes or excludes automatically match
	if c == nil || len(c.Include) == 0 && len(c.Exclude) == 0 {
		return true
	}

	// Exclusions are processed first. So we can include everything and then
	// selectively include others.
	if len(c.Exclude) != 0 {
		var matches int

		for key, val := range c.Exclude {
			if ok, _ := doublestar.Match(val, params[key]); ok {
				matches++
			}
		}
		if matches == len(c.Exclude) {
			return false
		}
	}
	for key, val := range c.Include {
		if ok, _ := doublestar.Match(val, params[key]); !ok {
			return false
		}
	}
	return true
}

// UnmarshalYAML unmarshal the constraint map.
func (c *Map) UnmarshalYAML(unmarshal func(any) error) error {
	out1 := struct {
		Include map[string]string
		Exclude map[string]string
	}{
		Include: map[string]string{},
		Exclude: map[string]string{},
	}

	out2 := map[string]string{}

	_ = unmarshal(&out1) // it contains include and exclude statement
	_ = unmarshal(&out2) // it contains no include/exclude statement, assume include as default

	c.Include = out1.Include
	c.Exclude = out1.Exclude
	for k, v := range out2 {
		c.Include[k] = v
	}
	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (c Map) MarshalYAML() (any, error) {
	switch {
	case len(c.Include) == 0 && len(c.Exclude) == 0:
		return nil, nil
	case len(c.Exclude) == 0:
		return c.Include, nil
	case len(c.Include) == 0 && len(c.Exclude) != 0:
		return struct {
			Exclude map[string]string
		}{Exclude: c.Exclude}, nil
	default:
		// we can not return type Map as it would lead to infinite recursion :/
		return struct {
			Include map[string]string `yaml:"include,omitempty"`
			Exclude map[string]string `yaml:"exclude,omitempty"`
		}{
			Include: c.Include,
			Exclude: c.Exclude,
		}, nil
	}
}
