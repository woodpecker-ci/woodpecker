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

package constraint

import (
	"fmt"
	"maps"
	"path"
	"slices"

	"github.com/expr-lang/expr"
	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	yamlBaseTypes "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

type (
	// When defines a set of runtime constraints.
	When struct {
		// If true then read from a list of constraint
		Constraints []Constraint
	}

	Constraint struct {
		Ref      List                        `yaml:"ref,omitempty"`
		Repo     List                        `yaml:"repo,omitempty"`
		Instance List                        `yaml:"instance,omitempty"`
		Platform List                        `yaml:"platform,omitempty"`
		Branch   List                        `yaml:"branch,omitempty"`
		Cron     List                        `yaml:"cron,omitempty"`
		Status   List                        `yaml:"status,omitempty"`
		Matrix   Map                         `yaml:"matrix,omitempty"`
		Local    optional.Option[bool]       `yaml:"local,omitempty"`
		Path     Path                        `yaml:"path,omitempty"`
		Evaluate string                      `yaml:"evaluate,omitempty"`
		Event    yamlBaseTypes.StringOrSlice `yaml:"event,omitempty"`
	}
)

func (when *When) IsEmpty() bool {
	return len(when.Constraints) == 0
}

// Returns true if at least one of the internal constraints is true.
func (when *When) Match(metadata metadata.Metadata, global bool, env map[string]string) (bool, error) {
	for _, c := range when.Constraints {
		match, err := c.Match(metadata, global, env)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}

	if when.IsEmpty() {
		// test against default Constraints
		empty := &Constraint{}
		return empty.Match(metadata, global, env)
	}
	return false, nil
}

func (when *When) IncludesStatusFailure() bool {
	for _, c := range when.Constraints {
		if c.Status.Includes("failure") {
			return true
		}
	}

	return false
}

func (when *When) IncludesStatusSuccess() bool {
	// "success" acts differently than "failure" in that it's
	// presumed to be included unless it's specifically not part
	// of the list
	if when.IsEmpty() {
		return true
	}
	for _, c := range when.Constraints {
		if len(c.Status.Include) == 0 || c.Status.Includes("success") {
			return true
		}
	}
	return false
}

// False if (any) non local.
func (when *When) IsLocal() bool {
	for _, c := range when.Constraints {
		if !c.Local.ValueOrDefault(true) {
			return false
		}
	}
	return true
}

func (when *When) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		if err := value.Decode(&when.Constraints); err != nil {
			return err
		}

	case yaml.MappingNode:
		c := Constraint{}
		if err := value.Decode(&c); err != nil {
			return err
		}
		when.Constraints = append(when.Constraints, c)

	default:
		return fmt.Errorf("not supported yaml kind: %v", value.Kind)
	}

	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (when When) MarshalYAML() (any, error) {
	// clean up local if true make it none as we will default to true
	for i := range when.Constraints {
		if when.Constraints[i].Local.ValueOrDefault(true) {
			when.Constraints[i].Local = optional.None[bool]()
		}
	}

	switch len(when.Constraints) {
	case 0:
		return nil, nil
	case 1:
		return when.Constraints[0], nil
	default:
		return when.Constraints, nil
	}
}

// Match returns true if all constraints match the given input. If a single
// constraint fails a false value is returned.
func (c *Constraint) Match(m metadata.Metadata, global bool, env map[string]string) (bool, error) {
	match := true
	if !global {
		// apply step only filters
		match = c.Matrix.Match(m.Workflow.Matrix)
	}

	match = match && c.Platform.Match(m.Sys.Platform) &&
		(len(c.Event) == 0 || slices.Contains(c.Event, m.Curr.Event)) &&
		c.Repo.Match(path.Join(m.Repo.Owner, m.Repo.Name)) &&
		c.Ref.Match(m.Curr.Commit.Ref) &&
		c.Instance.Match(m.Sys.Host)

	// changed files filter apply only for pull-request and push events
	if metadata.EventIsPull(m.Curr.Event) || m.Curr.Event == metadata.EventPush {
		match = match && c.Path.Match(m.Curr.Commit.ChangedFiles, m.Curr.Commit.Message)
	}

	if m.Curr.Event != metadata.EventTag {
		match = match && c.Branch.Match(m.Curr.Commit.Branch)
	}

	if m.Curr.Event == metadata.EventCron {
		match = match && c.Cron.Match(m.Curr.Cron)
	}

	if c.Evaluate != "" {
		if env == nil {
			env = m.Environ()
		} else {
			maps.Copy(env, m.Environ())
		}
		out, err := expr.Compile(c.Evaluate, expr.Env(env), expr.AllowUndefinedVariables(), expr.AsBool())
		if err != nil {
			return false, err
		}
		result, err := expr.Run(out, env)
		if err != nil {
			return false, err
		}
		bResult, ok := result.(bool)
		if !ok {
			return false, fmt.Errorf("could not parse result: %v", result)
		}
		match = match && bResult
	}

	return match, nil
}
