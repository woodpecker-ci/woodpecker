package constraint

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// When defines a set of runtime constraints.
	When struct {
		// If true then read from a list of constraint
		Constraints []Constraint
	}

	Constraint struct {
		Ref         List
		Repo        List
		Instance    List
		Platform    List
		Environment List
		Event       List
		Branch      List
		Cron        List
		Status      List
		Matrix      Map
		Local       types.BoolTrue
		Path        Path
	}

	// List defines a runtime constraint for exclude & include string slices.
	List struct {
		Include []string
		Exclude []string
	}

	// Map defines a runtime constraint for exclude & include map strings.
	Map struct {
		Include map[string]string
		Exclude map[string]string
	}

	// Path defines a runtime constrain for exclude & include paths.
	Path struct {
		Include       []string
		Exclude       []string
		IgnoreMessage string `yaml:"ignore_message,omitempty"`
	}
)

func (when *When) IsEmpty() bool {
	return len(when.Constraints) == 0
}

// Returns true if at least one of the internal constraints is true.
func (when *When) Match(metadata frontend.Metadata) bool {
	for _, c := range when.Constraints {
		if c.Match(metadata) {
			return true
		}
	}
	return when.IsEmpty()
}

func (when *When) IncludesStatus(status string) bool {
	for _, c := range when.Constraints {
		if c.Status.Includes(status) {
			return true
		}
	}

	return false
}

func (when *When) ExcludesStatus(status string) bool {
	for _, c := range when.Constraints {
		if !c.Status.Excludes(status) {
			return false
		}
	}

	return len(when.Constraints) > 0
}

// False if (any) non local
func (when *When) IsLocal() bool {
	for _, c := range when.Constraints {
		if !c.Local.Bool() {
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

// Match returns true if all constraints match the given input. If a single
// constraint fails a false value is returned.
func (c *Constraint) Match(metadata frontend.Metadata) bool {
	// if event filter is not set, set default
	if len(c.Event.Include) == 0 && len(c.Event.Exclude) == 0 {
		c.Event.Include = []string{
			frontend.EventPush,
			frontend.EventPull,
			frontend.EventTag,
			frontend.EventDeploy,
		}
	}

	match := c.Platform.Match(metadata.Sys.Platform) &&
		c.Environment.Match(metadata.Curr.Target) &&
		c.Event.Match(metadata.Curr.Event) &&
		c.Repo.Match(metadata.Repo.Name) &&
		c.Ref.Match(metadata.Curr.Commit.Ref) &&
		c.Instance.Match(metadata.Sys.Host) &&
		c.Matrix.Match(metadata.Job.Matrix)

	// changed files filter apply only for pull-request and push events
	if metadata.Curr.Event == frontend.EventPull || metadata.Curr.Event == frontend.EventPush {
		match = match && c.Path.Match(metadata.Curr.Commit.ChangedFiles, metadata.Curr.Commit.Message)
	}

	if metadata.Curr.Event != frontend.EventTag {
		match = match && c.Branch.Match(metadata.Curr.Commit.Branch)
	}

	if metadata.Curr.Event == frontend.EventCron {
		// cron title is storend in message
		match = match && c.Cron.Match(metadata.Curr.Commit.Message)
	}

	return match
}

// Match returns true if the string matches the include patterns and does not
// match any of the exclude patterns.
func (c *List) Match(v string) bool {
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

// UnmarshalYAML unmarshals the constraint.
func (c *List) UnmarshalYAML(value *yaml.Node) error {
	out1 := struct {
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
func (c *Map) Match(params map[string]string) bool {
	// when no includes or excludes automatically match
	if len(c.Include) == 0 && len(c.Exclude) == 0 {
		return true
	}

	// exclusions are processed first. So we can include everything and then
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

// UnmarshalYAML unmarshals the constraint map.
func (c *Map) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
func (c *Path) UnmarshalYAML(value *yaml.Node) error {
	out1 := struct {
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
// or if commit message contains ignore message.
func (c *Path) Match(v []string, message string) bool {
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

// Excludes returns true if the string matches any of the exclude patterns.
func (c *Path) Excludes(v []string) bool {
	for _, pattern := range c.Exclude {
		for _, file := range v {
			if ok, _ := doublestar.Match(pattern, file); ok {
				return true
			}
		}
	}
	return false
}
