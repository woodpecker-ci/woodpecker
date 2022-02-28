package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// Config defines a pipeline configuration.
	Config struct {
		// TODO: change to When (also in Pipeline)
		ConstraintsArray []constraint.Constraints `yaml:"whenArray,omitempty"`
		Constraints      constraint.Constraints   `yaml:"when,omitempty"`
		Cache            types.Stringorslice
		Platform         string
		Branches         constraint.List
		Workspace        Workspace
		Clone            Containers
		Pipeline         Containers
		Services         Containers
		Networks         Networks
		Volumes          Volumes
		Labels           types.SliceorMap
		DependsOn        []string `yaml:"depends_on,omitempty"`
		RunsOn           []string `yaml:"runs_on,omitempty"`
		SkipClone        bool     `yaml:"skip_clone"`
	}

	// Workspace defines a pipeline workspace.
	Workspace struct {
		Base string
		Path string
	}
)

func (c *Config) MatchConstraints(meta frontend.Metadata) bool {
	if c.Constraints.Match(meta) {
		return true
	}

	for _, c := range c.ConstraintsArray {
		if c.Match(meta) {
			return true
		}
	}
	return false
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*Config, error) {
	out := new(Config)
	err := yaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}
