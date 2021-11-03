package schema

import (
	_ "embed"
	"fmt"

	"github.com/xeipuuv/gojsonschema"

	"github.com/woodpecker-ci/woodpecker/shared/yml"
)

//go:embed schema.json
var schemaDefinition []byte

func Lint(file string) (error, []gojsonschema.ResultError) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)
	j, err := yml.LoadYmlFileAsJson(file)
	if err != nil {
		return fmt.Errorf("Failed to load yml file %w", err), nil
	}

	documentLoader := gojsonschema.NewBytesLoader(j)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Validation failed %w", err), nil
	}

	if !result.Valid() {
		return fmt.Errorf("Config not valid"), result.Errors()
	}

	return nil, nil
}
