package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// Config defines a pipeline configuration.
	Config struct {
		Cache     types.Stringorslice
		Platform  string
		Branches  constraint.Constraint
		Workspace Workspace
		Clone     Containers
		Pipeline  Containers
		Services  Containers
		Networks  Networks
		Volumes   Volumes
		Labels    types.SliceorMap
		DependsOn []string `yaml:"depends_on,omitempty"`
		RunsOn    []string `yaml:"runs_on,omitempty"`
		SkipClone bool     `yaml:"skip_clone"`
	}

	// Workspace defines a pipeline workspace.
	Workspace struct {
		Base string
		Path string
	}
)

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
