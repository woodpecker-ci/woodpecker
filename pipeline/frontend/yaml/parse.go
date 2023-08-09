package yaml

import (
	"fmt"

	"codeberg.org/6543/xyaml"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*types.Workflow, error) {
	out := new(types.Workflow)
	err := xyaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}

	// fail hard on deprecated branch filter
	if out.BranchesDontUseIt != nil {
		return nil, fmt.Errorf("\"branches:\" filter got removed, use \"branch\" in global when filter")
	}

	// fail hard on deprecated pipeline keyword
	if len(out.PipelineDontUseIt.ContainerList) != 0 {
		return nil, fmt.Errorf("\"pipeline:\" got removed, user \"steps:\"")
	}

	// support deprecated platform filter
	if out.PlatformDontUseIt == "" {
		if _, set := out.Labels["platform"]; !set {
			out.Labels["platform"] = out.PlatformDontUseIt
		}
	}

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*types.Workflow, error) {
	return ParseBytes(
		[]byte(s),
	)
}
