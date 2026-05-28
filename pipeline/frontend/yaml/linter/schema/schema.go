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

package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"

	"codeberg.org/6543/go-yaml2json/v2"
	"codeberg.org/6543/xyaml/v2"
	"github.com/xeipuuv/gojsonschema"
	"go.yaml.in/yaml/v4"
)

//go:embed schema.json
var schemaDefinition []byte

// Lint lints an io.Reader against the Woodpecker `schema.json`.
func Lint(r io.Reader) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)

	// read yaml config
	rBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to load yml file %w", err)
	}

	// resolve sequence merges
	yamlDoc := new(yaml.Node)
	if err := xyaml.Unmarshal(rBytes, yamlDoc); err != nil {
		return nil, fmt.Errorf("failed to parse yml file %w", err)
	}

	// convert to json
	jsonDoc, err := yaml2json.ConvertNode(yamlDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonDoc)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed %w", err)
	}

	if !result.Valid() {
		return filterRedundantCompositionErrors(result.Errors()), fmt.Errorf("config not valid")
	}

	return nil, nil
}

func LintString(s string) ([]gojsonschema.ResultError, error) {
	return Lint(bytes.NewBufferString(s))
}

func filterRedundantCompositionErrors(schemaErrors []gojsonschema.ResultError) []gojsonschema.ResultError {
	filtered := make([]gojsonschema.ResultError, 0, len(schemaErrors))
	for index, schemaError := range schemaErrors {
		if isCompositionError(schemaError) && hasSpecificSchemaError(schemaErrors, index, schemaError.Field()) {
			continue
		}

		filtered = append(filtered, schemaError)
	}

	return filtered
}

func isCompositionError(schemaError gojsonschema.ResultError) bool {
	switch schemaError.Type() {
	case "number_one_of", "number_any_of":
		return true
	default:
		return false
	}
}

func hasSpecificSchemaError(schemaErrors []gojsonschema.ResultError, currentIndex int, field string) bool {
	for index, schemaError := range schemaErrors {
		if index == currentIndex || isCompositionError(schemaError) {
			continue
		}

		if isSameFieldOrChild(schemaError.Field(), field) {
			return true
		}
	}

	return false
}

func isSameFieldOrChild(field, parent string) bool {
	if parent == "(root)" {
		return true
	}

	return field == parent || strings.HasPrefix(field, parent+".")
}
