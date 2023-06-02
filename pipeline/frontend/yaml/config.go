package yaml

import (
	"fmt"

	"codeberg.org/6543/xyaml"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

type (
	// Config defines a pipeline configuration.
	Config struct {
		When      constraint.When `yaml:"when,omitempty"`
		Cache     types.StringOrSlice
		Platform  string
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
		// Deprecated use When.Branch
		BranchesDontUseIt *constraint.List `yaml:"branches,omitempty"`
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
	err := xyaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}

	// support deprecated branch filter
	if out.BranchesDontUseIt != nil {
		if out.When.Constraints == nil {
			out.When.Constraints = []constraint.Constraint{{Branch: *out.BranchesDontUseIt}}
		} else if len(out.When.Constraints) == 1 && out.When.Constraints[0].Branch.IsEmpty() {
			out.When.Constraints[0].Branch = *out.BranchesDontUseIt
		} else {
			return nil, fmt.Errorf("could not apply deprecated branches filter into global when filter")
		}
		out.BranchesDontUseIt = nil
	}

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}
