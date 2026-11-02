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
	"fmt"
	"sort"

	go_yaml "go.yaml.in/yaml/v4"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
)

const (
	limitVariables = 10
	maxWorkflows   = 64
)

var (
	ErrorMaxWorkflowsReached = fmt.Errorf("matrix config produced more than %d workflows", maxWorkflows)

	ErrorMaxAxisVariablesReached = fmt.Errorf("matrix config has more than %d variables in an axis", limitVariables)
)

// Matrix represents the pipeline matrix.
type Matrix map[string][]string

// Axis represents a single permutation of entries from the pipeline matrix.
type Axis map[string]string

// Parse parses the Yaml matrix definition.
func Parse(data []byte) ([]Axis, error) {
	value := new(go_yaml.Node)
	if err := yaml.Unmarshal(data, value); err != nil {
		return nil, compilerError(err)
	}

	if value.Kind == go_yaml.DocumentNode && len(value.Content) == 1 {
		value = value.Content[0]
	}

	if value == nil || value.Kind != go_yaml.MappingNode {
		return []Axis{}, nil
	}

	matrixNode := dereferenceAlias(mappingValue(value, "matrix"))
	if matrixNode == nil ||
		(matrixNode.Kind == go_yaml.ScalarNode && matrixNode.Tag == "!!null") ||
		matrixNode.Kind == go_yaml.MappingNode && len(matrixNode.Content) == 0 {
		return []Axis{}, nil
	}

	if includeNode := mappingValue(matrixNode, "include"); includeNode != nil {
		return parseInclude(includeNode)
	}

	matrix, err := parseMatrix(matrixNode)
	if err != nil {
		return nil, err
	}
	return calcAxes(matrix)
}

func calcAxes(matrix Matrix) ([]Axis, error) {
	// fail early if we know the outcome is more than allowed already
	if len(matrix) > maxWorkflows {
		return nil, ErrorMaxWorkflowsReached
	}

	// calculate number of permutations and extract the list of variable
	// names (ie go_version, redis_version, etc)
	var axisCount int
	var names []string
	for name, values := range matrix {
		// an axis without values contributes no permutations
		if len(values) == 0 {
			continue
		}

		if len(values) > limitVariables {
			return nil, ErrorMaxAxisVariablesReached
		}
		if axisCount == 0 {
			axisCount = len(values)
		} else {
			axisCount *= len(values)
		}

		// enforce limit
		if axisCount > maxWorkflows {
			return nil, ErrorMaxWorkflowsReached
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
		axis := make(map[string]string)
		// number of consecutive axes that share the same value of a variable
		axesPerValue := axisCount
		for _, name := range names {
			values := matrix[name]
			axesPerValue /= len(values)
			valueIndex := axisIndex / axesPerValue % len(values)
			axis[name] = values[valueIndex]
		}

		// append to the list of axis.
		axes = append(axes, axis)
	}

	return uniqueAxes(axes), nil
}

func parseMatrix(node *go_yaml.Node) (Matrix, error) {
	matrix := make(Matrix)
	if err := node.Decode(&matrix); err != nil {
		return nil, compilerError(err)
	}
	return matrix, nil
}

func parseInclude(node *go_yaml.Node) ([]Axis, error) {
	var axes []Axis
	if err := node.Decode(&axes); err != nil {
		return nil, compilerError(err)
	}

	if len(axes) > maxWorkflows {
		return nil, ErrorMaxWorkflowsReached
	}

	for _, values := range axes {
		if len(values) > limitVariables {
			return nil, ErrorMaxAxisVariablesReached
		}
	}

	return uniqueAxes(axes), nil
}

func uniqueAxes(axes []Axis) []Axis {
	unique := make([]Axis, 0, len(axes))
	for _, axis := range axes {
		duplicate := false
		for _, existing := range unique {
			if equalAxes(axis, existing) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, axis)
		}
	}
	return unique
}

func equalAxes(left, right Axis) bool {
	if len(left) != len(right) {
		return false
	}
	for name, value := range left {
		otherValue, ok := right[name]
		if !ok || otherValue != value {
			return false
		}
	}
	return true
}

func mappingValue(node *go_yaml.Node, key string) *go_yaml.Node {
	node = dereferenceAlias(node)
	if node == nil || node.Kind != go_yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func dereferenceAlias(node *go_yaml.Node) *go_yaml.Node {
	for node != nil && node.Kind == go_yaml.AliasNode {
		node = node.Alias
	}
	return node
}

func compilerError(err error) error {
	return &pipeline_errors.PipelineError{Message: err.Error(), Type: pipeline_errors.PipelineErrorTypeCompiler}
}
