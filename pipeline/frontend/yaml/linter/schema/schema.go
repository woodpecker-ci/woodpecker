package schema

import (
	_ "embed"
	"fmt"
	"io"

	"codeberg.org/6543/go-yaml2json"
	"codeberg.org/6543/xyaml"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed schema.json
var schemaDefinition []byte

// Lint lints an io.Reader against the Woodpecker schema.json
func Lint(r io.Reader) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)

	// read yaml config
	rBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to load yml file %w", err)
	}

	// resolve sequence merges
	yamlDoc := new(yaml.Node)
	if err := xyaml.Unmarshal(rBytes, yamlDoc); err != nil {
		return nil, fmt.Errorf("Failed to parse yml file %w", err)
	}

	// convert to json
	jsonDoc, err := yaml2json.ConvertNode(yamlDoc)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert yaml %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonDoc)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("Validation failed %w", err)
	}

	if !result.Valid() {
		return result.Errors(), fmt.Errorf("Config not valid")
	}

	return nil, nil
}

func LintString(s string) ([]gojsonschema.ResultError, error) {
	return Lint(bytes.NewBufferString(s))
}
