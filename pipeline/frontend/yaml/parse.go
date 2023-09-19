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

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
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

	// fail hard on deprecated branch filter
	if out.BranchesDontUseIt != nil {
		return nil, fmt.Errorf("\"branches:\" filter got removed, use \"branch\" in global when filter instead")
	}

	// fail hard on deprecated pipeline keyword
	if len(out.PipelineDontUseIt.ContainerList) != 0 {
		return nil, fmt.Errorf("\"pipeline:\" got removed, use \"steps:\" instead")
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
		return 0, ErrMissingVersion
	}

	if ver.Version != Version {
		return 0, ErrUnsuportedVersion
	}
	return ver.Version, nil
}
