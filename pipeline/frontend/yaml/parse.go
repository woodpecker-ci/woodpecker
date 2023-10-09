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
	"codeberg.org/6543/xyaml"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
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
		return 1, nil
	}

	if ver.Version != Version {
		return 0, ErrUnsuportedVersion
	}
	return ver.Version, nil
}
