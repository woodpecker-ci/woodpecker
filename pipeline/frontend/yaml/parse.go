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

package yaml

import (
	"fmt"

	"codeberg.org/6543/xyaml"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
)

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*types.Workflow, error) {
	out := new(types.Workflow)
	err := xyaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}

	// support deprecated branch filter
	if out.BranchesDoNotUseIt != nil {
		switch {
		case out.When.Constraints == nil:
			out.When.Constraints = []constraint.Constraint{{Branch: *out.BranchesDoNotUseIt}}
		case len(out.When.Constraints) == 1 && out.When.Constraints[0].Branch.IsEmpty():
			out.When.Constraints[0].Branch = *out.BranchesDoNotUseIt
		default:
			return nil, fmt.Errorf("could not apply deprecated branches filter into global when filter")
		}
		out.BranchesDoNotUseIt = nil
	}

	// support deprecated pipeline keyword
	if len(out.PipelineDoNotUseIt.ContainerList) != 0 && len(out.Steps.ContainerList) == 0 {
		out.Steps.ContainerList = out.PipelineDoNotUseIt.ContainerList
	}

	// support deprecated platform filter
	if out.PlatformDoNotUseIt != "" {
		if out.Labels == nil {
			out.Labels = make(map[string]string)
		}
		if _, set := out.Labels["platform"]; !set {
			out.Labels["platform"] = out.PlatformDoNotUseIt
		}
		out.PlatformDoNotUseIt = ""
	}
	out.PipelineDoNotUseIt.ContainerList = nil

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*types.Workflow, error) {
	return ParseBytes(
		[]byte(s),
	)
}
