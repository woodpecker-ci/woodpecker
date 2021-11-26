package yaml

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// Constraints defines a set of runtime constraints.
	Constraints struct {
		Ref         Constraint
		Repo        Constraint
		Instance    Constraint
		Platform    Constraint
		Environment Constraint
		Event       Constraint
		Branch      Constraint
		Status      Constraint
		Matrix      ConstraintMap
		Local       types.BoolTrue
		Path        ConstraintPath
	}

	// Constraint defines a runtime constraint.
	Constraint struct {
		Include []string
		Exclude []string
	}

	// ConstraintMap defines a runtime constraint map.
	ConstraintMap struct {
		Include map[string]string
		Exclude map[string]string
	}

	// ConstraintPath defines a runtime constrain for paths
	ConstraintPath struct {
		Include       []string
		Exclude       []string
		IgnoreMessage string `yaml:"ignore_message,omitempty"`
	}
)

// Match returns true if all constraints match the given input. If a single
// constraint fails a false value is returned.
func (c *Constraints) Match(metadata frontend.Metadata) bool {
	return c.Platform.Match(metadata.Sys.Arch) &&
		c.Environment.Match(metadata.Curr.Target) &&
		c.Event.Match(metadata.Curr.Event) &&
		c.Branch.Match(metadata.Curr.Commit.Branch) &&
		c.Repo.Match(metadata.Repo.Name) &&
		c.Ref.Match(metadata.Curr.Commit.Ref) &&
		c.Instance.Match(metadata.Sys.Host) &&
		c.Matrix.Match(metadata.Job.Matrix) &&
		c.Path.Match(metadata.Curr.Commit.ChangedFiles, metadata.Curr.Commit.Message)
}

// Match returns true if the string matches the include patterns and does not
// match any of the exclude patterns.
func (c *Constraint) Match(v string) bool {
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
func (c *Constraint) Includes(v string) bool {
	for _, pattern := range c.Include {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// Excludes returns true if the string matches the exclude patterns.
func (c *Constraint) Excludes(v string) bool {
	for _, pattern := range c.Exclude {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// UnmarshalYAML unmarshals the constraint.
func (c *Constraint) UnmarshalYAML(value *yaml.Node) error {
	var out1 = struct {
		Include types.Stringorslice
		Exclude types.Stringorslice
	}{}

	var out2 types.Stringorslice

	err1 := value.Decode(&out1)
	err2 := value.Decode(&out2)

	c.Exclude = out1.Exclude
	c.Include = append(
		out1.Include,
		out2...,
	)

	if err1 != nil && err2 != nil {
		y, _ := yaml.Marshal(value)
		return fmt.Errorf("Could not parse condition: %s", y)
	}

	return nil
}

// Match returns true if the params matches the include key values and does not
// match any of the exclude key values.
func (c *ConstraintMap) Match(params map[string]string) bool {
	// when no includes or excludes automatically match
	if len(c.Include) == 0 && len(c.Exclude) == 0 {
		return true
	}
	// exclusions are processed first. So we can include everything and then
	// selectively include others.
	if len(c.Exclude) != 0 {
		var matches int

		for key, val := range c.Exclude {
			if params[key] == val {
				matches++
			}
		}
		if matches == len(c.Exclude) {
			return false
		}
	}
	for key, val := range c.Include {
		if params[key] != val {
			return false
		}
	}
	return true
}

// UnmarshalYAML unmarshals the constraint map.
func (c *ConstraintMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

// UnmarshalYAML unmarshals the constraint.
func (c *ConstraintPath) UnmarshalYAML(value *yaml.Node) error {
	var out1 = struct {
		Include       types.Stringorslice `yaml:"include,omitempty"`
		Exclude       types.Stringorslice `yaml:"exclude,omitempty"`
		IgnoreMessage string              `yaml:"ignore_message,omitempty"`
	}{}

	var out2 types.Stringorslice

	err1 := value.Decode(&out1)
	err2 := value.Decode(&out2)

	c.Exclude = out1.Exclude
	c.IgnoreMessage = out1.IgnoreMessage
	c.Include = append(
		out1.Include,
		out2...,
	)

	if err1 != nil && err2 != nil {
		y, _ := yaml.Marshal(value)
		return fmt.Errorf("Could not parse condition: %s", y)
	}

	return nil
}

// Match returns true if file paths in string slice matches the include and not exclude patterns
//  or if commit message contains ignore message.
func (c *ConstraintPath) Match(v []string, message string) bool {
	// ignore file pattern matches if the commit message contains a pattern
	if len(c.IgnoreMessage) > 0 && strings.Contains(strings.ToLower(message), strings.ToLower(c.IgnoreMessage)) {
		return true
	}
	// always match if there are no commit files (empty commit)
	if len(v) == 0 {
		return true
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
func (c *ConstraintPath) Includes(v []string) bool {
	for _, pattern := range c.Include {
		for _, file := range v {
			if ok, _ := doublestar.Match(pattern, file); ok {
				return true
			}
		}
	}
	return false
}

// Excludes returns true if the string matches any of the exclude patterns.
func (c *ConstraintPath) Excludes(v []string) bool {
	for _, pattern := range c.Exclude {
		for _, file := range v {
			if ok, _ := doublestar.Match(pattern, file); ok {
				return true
			}
		}
	}
	return false
}
