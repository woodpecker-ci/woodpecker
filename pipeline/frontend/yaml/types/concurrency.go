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

	"go.yaml.in/yaml/v4"
)

// Concurrency limits how many instances of a workflow may run at the same
// time. It can be unmarshaled from:
//   - an integer: `concurrency: 1` (limit only, default per-workflow group)
//   - an object: `concurrency: {limit: 1, group: deploy}`
type Concurrency struct {
	// Limit is the maximum number of workflows sharing the same group that
	// are allowed to run at the same time. A value <= 0 disables the limit.
	Limit int `yaml:"limit,omitempty"`
	// Group identifies which workflows are mutually limited. Workflows that
	// resolve to the same group within a repository are serialized according
	// to the limit. When empty the limit applies per workflow, so different
	// runs of the same workflow are limited against each other.
	Group string `yaml:"group,omitempty"`
}

// UnmarshalYAML implements the Unmarshaler interface. It inspects the YAML
// node kind to decide how to decode, instead of speculatively decoding into an
// int and falling back on error:
//   - a scalar (`concurrency: 1`) sets only the limit
//   - a mapping (`concurrency: {limit: 1, group: deploy}`) sets both fields
func (c *Concurrency) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.DocumentNode && len(value.Content) == 1 {
		value = value.Content[0]
	}
	// resolve anchors/aliases so the kind switch sees the referenced node.
	if value.Kind == yaml.AliasNode {
		value = value.Alias
	}

	switch value.Kind {
	// shorthand: `concurrency: <int>`
	case yaml.ScalarNode:
		var limit int
		if err := value.Decode(&limit); err != nil {
			return fmt.Errorf("failed to unmarshal concurrency limit: %w", err)
		}
		c.Limit = limit
		return nil

	// full form: `concurrency: {limit: <int>, group: <string>}`
	case yaml.MappingNode:
		// alias type avoids recursing into this UnmarshalYAML.
		type concurrencyAlias Concurrency
		var tmp concurrencyAlias
		if err := value.Decode(&tmp); err != nil {
			return fmt.Errorf("failed to unmarshal concurrency: %w", err)
		}
		*c = Concurrency(tmp)
		return nil

	default:
		return fmt.Errorf("failed to unmarshal concurrency: expected an integer or a mapping, got %v", value.Kind)
	}
}

// MarshalYAML implements the Marshaler interface. It mirrors UnmarshalYAML so
// the config round-trips: when only a limit is set (no explicit group) it emits
// the shorthand `concurrency: <int>`, otherwise the full `{limit, group}` form.
func (c Concurrency) MarshalYAML() (any, error) {
	if c.Group == "" {
		return c.Limit, nil
	}
	// alias type avoids recursing into this MarshalYAML.
	type concurrencyAlias Concurrency
	return concurrencyAlias(c), nil
}

// IsZero treats a disabled (limit <= 0) concurrency as empty for omitempty.
func (c Concurrency) IsZero() bool {
	return c.Limit <= 0 && c.Group == ""
}
