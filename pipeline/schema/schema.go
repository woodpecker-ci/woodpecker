package schema

import (
	_ "embed"
	"fmt"

	"github.com/xeipuuv/gojsonschema"

	"github.com/woodpecker-ci/woodpecker/shared/yml"
)

//go:embed schema.json
var schemaDefinition []byte

func Lint(file string) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)
	j, err := yml.LoadYmlFileAsJSON(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to load yml file %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(j)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("Validation failed %w", err)
	}

	if !result.Valid() {
		return result.Errors(), fmt.Errorf("Config not valid")
	}

	return nil, nil
}
