package schema

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/xeipuuv/gojsonschema"

	"github.com/woodpecker-ci/woodpecker/shared/yml"
)

//go:embed schema.json
var schemaDefinition []byte

// Lint lints an io.Reader against the Woodpecker schema.json
func Lint(r io.Reader) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)
	j, err := yml.LoadYmlReaderAsJSON(r)
	if err != nil {
		return nil, fmt.Errorf("failed to load yml file %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(j)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed %w", err)
	}

	if !result.Valid() {
		return result.Errors(), fmt.Errorf("config not valid")
	}

	return nil, nil
}
