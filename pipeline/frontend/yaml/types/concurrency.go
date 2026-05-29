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

import "fmt"

// Concurrency limits how many instances of a workflow may run at the same
// time. It can be unmarshalled from:
//   - an integer: `concurrency: 1` (limit only, group defaults to the workflow name)
//   - an object: `concurrency: {limit: 1, group: deploy}`
type Concurrency struct {
	// Limit is the maximum number of workflows sharing the same group that
	// are allowed to run at the same time. A value <= 0 disables the limit.
	Limit int `yaml:"limit,omitempty"`
	// Group identifies which workflows are mutually limited. Workflows that
	// resolve to the same group within a repository are serialized according
	// to the limit. When empty it defaults to the workflow name, so different
	// runs of the same workflow are limited against each other.
	Group string `yaml:"group,omitempty"`
}

// UnmarshalYAML implements the Unmarshaler interface.
func (c *Concurrency) UnmarshalYAML(unmarshal func(any) error) error {
	// shorthand: `concurrency: <int>`
	var limit int
	if err := unmarshal(&limit); err == nil {
		c.Limit = limit
		return nil
	}

	// full form: `concurrency: {limit: <int>, group: <string>}`
	// use an alias type to avoid recursing into this UnmarshalYAML.
	type concurrencyAlias Concurrency
	var tmp concurrencyAlias
	if err := unmarshal(&tmp); err != nil {
		return fmt.Errorf("failed to unmarshal concurrency: %w", err)
	}
	*c = Concurrency(tmp)
	return nil
}

// IsZero treats a disabled (limit <= 0) concurrency as empty for omitempty.
func (c Concurrency) IsZero() bool {
	return c.Limit <= 0 && c.Group == ""
}
