package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"codeberg.org/6543/go-yaml2json"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaDefinition []byte

// Lint lints an io.Reader against the Woodpecker schema.json
func Lint(r io.Reader) ([]gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaDefinition)
	buff := new(bytes.Buffer)
	err := yaml2json.StreamConverter(r, buff)
	if err != nil {
		return nil, fmt.Errorf("failed to load yml file %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(buff.Bytes())
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed %w", err)
	}

	if !result.Valid() {
		return result.Errors(), fmt.Errorf("config not valid")
	}

	return nil, nil
}
