package yaml

import (
	"fmt"

	"codeberg.org/6543/xyaml"

	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/types"
)

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*types.Workflow, error) {
	out := new(types.Workflow)
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

	// support deprecated pipeline keyword
	if len(out.PipelineDontUseIt.ContainerList) != 0 && len(out.Steps.ContainerList) == 0 {
		out.Steps.ContainerList = out.PipelineDontUseIt.ContainerList
	}
	out.PipelineDontUseIt.ContainerList = nil

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*types.Workflow, error) {
	return ParseBytes(
		[]byte(s),
	)
}
