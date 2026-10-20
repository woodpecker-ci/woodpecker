// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matrix

import (
	"sort"
	"strings"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
)

const (
	limitVariables = 10
	maxAxes        = 64
)

// Matrix represents the pipeline matrix.
type Matrix map[string][]string

// Axis represents a single permutation of entries from the pipeline matrix.
type Axis map[string]string

// String returns a string representation of an Axis as a comma-separated list
// of environment variables.
func (a Axis) String() string {
	var envs []string
	for name, value := range a {
		envs = append(envs, name+"="+value)
	}
	return strings.Join(envs, " ")
}

// Parse parses the Yaml matrix definition.
func Parse(data []byte) ([]Axis, error) {
	includedAxes, err := parseInclude(data)
	if err == nil && len(includedAxes) != 0 {
		return includedAxes, nil
	}

	matrix, err := parseMatrix(data)
	if err != nil {
		return nil, err
	}

	if len(matrix) == 0 {
		return []Axis{}, nil
	}

	return calcAxes(matrix)
}

// ParseString parses the Yaml string matrix definition.
func ParseString(data string) ([]Axis, error) {
	return Parse([]byte(data))
}

func calcAxes(matrix Matrix) ([]Axis, error) {
	// calculate number of permutations and extract the list of variable
	// names (ie go_version, redis_version, etc)
	var axisCount int
	var names []string
	for name, values := range matrix {
		// an axis without values contributes no permutations
		if len(values) == 0 {
			continue
		}
		axisCount *= len(values)
		if axisCount == 0 {
			axisCount = len(values)
		}
		names = append(names, name)
	}
	// map iteration order is random: sort the names so the permutation order
	// is deterministic and derived workflow numbering is stable across
	// repeated compilations of the same pipeline.
	sort.Strings(names)

	// structure to hold the transformed result set
	var axes []Axis

	// for each axis calculate the unique set of values that should be used.
	for axisIndex := 0; axisIndex < axisCount; axisIndex++ {
		axis := map[string]string{}
		// number of consecutive axes that share the same value of a variable
		axesPerValue := axisCount
		for variableIndex, name := range names {
			values := matrix[name]
			axesPerValue /= len(values)
			valueIndex := axisIndex / axesPerValue % len(values)
			axis[name] = values[valueIndex]

			// enforce a maximum number of variables in the pipeline matrix.
			if variableIndex >= limitVariables {
				return nil, ErrorMaxAxisVariablesReached
			}
		}

		// append to the list of axis.
		axes = append(axes, axis)

		// enforce limit
		if axisIndex >= maxAxes {
			return nil, ErrorMaxAxesReached
		}

	}

	return axes, nil
}

func parseMatrix(data []byte) (Matrix, error) {
	config := struct {
		Matrix map[string][]string
	}{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, &pipeline_errors.PipelineError{Message: err.Error(), Type: pipeline_errors.PipelineErrorTypeCompiler}
	}
	return config.Matrix, nil
}

func parseInclude(data []byte) ([]Axis, error) {
	config := struct {
		Matrix struct {
			Include []Axis
		}
	}{}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, &pipeline_errors.PipelineError{Message: err.Error(), Type: pipeline_errors.PipelineErrorTypeCompiler}
	}

	if len(config.Matrix.Include) > maxAxes {
		return nil, ErrorMaxAxesReached
	}

	for _, values := range config.Matrix.Include {
		if len(values) > limitVariables {
			return nil, ErrorMaxAxisVariablesReached
		}
	}

	return config.Matrix.Include, nil
}
