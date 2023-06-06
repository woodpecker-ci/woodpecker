package yaml

import (
	"fmt"

	"codeberg.org/6543/xyaml"
	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*types.Workflow, error) {
	yamlVersion, err := checkVersion(b)
	if err != nil {
		&PipelineParseError{Err: err}
	}

	out := new(types.Workflow)
	err := xyaml.Unmarshal(b, out)
	if err != nil {
		return nil, &PipelineParseError{Err: err}
	}

	// make sure detected version is set
	out.Version = yamlVersion

	// support deprecated branch filter
	if out.BranchesDontUseIt != nil {
		if out.When.Constraints == nil {
			out.When.Constraints = []constraint.Constraint{{Branch: *out.BranchesDontUseIt}}
		} else if len(out.When.Constraints) == 1 && out.When.Constraints[0].Branch.IsEmpty() {
			out.When.Constraints[0].Branch = *out.BranchesDontUseIt
		} else {
			return nil, &PipelineParseError{Err: fmt.Errorf("could not apply deprecated branches filter into global when filter")}
		}
		out.BranchesDontUseIt = nil
	}

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*types.Workflow, error) {
	return ParseBytes(
		[]byte(s),
	)
}

func checkVersion(b []byte) (string, error) {
	verStr := struct {
		Version string `yaml:"version"`
	}{}
	_ = yaml.Unmarshal(b, &verStr)
	// TODO: should we require a version number -> in therms of UX we should not, in terms of strong typisation we should
	if verStr == "" {
		verStr = Version
	}

	if verStr != Version {
		return "", ErrUnsuportedVersion
	}
	return verStr, nil
}
