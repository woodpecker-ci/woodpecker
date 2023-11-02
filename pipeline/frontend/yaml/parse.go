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

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
<<<<<<< HEAD
	"github.com/woodpecker-ci/woodpecker/shared/constant"
=======
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
>>>>>>> main
)

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*types.Workflow, error) {
	yamlVersion, err := checkVersion(b)
	if err != nil {
		return nil, &PipelineParseError{Err: err}
	}

	out := new(types.Workflow)
	err = xyaml.Unmarshal(b, out)
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
			return nil, fmt.Errorf("could not apply deprecated branches filter into global when filter")
		}
		out.BranchesDontUseIt = nil
	}

	// support deprecated pipeline keyword
	if len(out.PipelineDontUseIt.ContainerList) != 0 && len(out.Steps.ContainerList) == 0 {
		out.Steps.ContainerList = out.PipelineDontUseIt.ContainerList
	}

	// support deprecated platform filter
	if out.PlatformDontUseIt != "" {
		if out.Labels == nil {
			out.Labels = make(base.SliceOrMap)
		}
		if _, set := out.Labels["platform"]; !set {
			out.Labels["platform"] = out.PlatformDontUseIt
		}
		out.PlatformDontUseIt = ""
	}
	out.PipelineDontUseIt.ContainerList = nil
>>>>>>> main

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*types.Workflow, error) {
	return ParseBytes(
		[]byte(s),
	)
}

func checkVersion(b []byte) (int, error) {
	ver := struct {
		Version int `yaml:"version"`
	}{}
	_ = xyaml.Unmarshal(b, &ver)
	if ver.Version == 0 {
		// default: version 1
		return constant.DefaultPipelineVersion, nil
	}

	if ver.Version != Version {
		return 0, ErrUnsuportedVersion
	}
	return ver.Version, nil
}
