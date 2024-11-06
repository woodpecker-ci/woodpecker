// Copyright 2023 Woodpecker Authors
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

// *********************************************************
// This is a generator tool, to update the openapi.json file
// *********************************************************

//go:build generate
// +build generate

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/docs"
)

func main() {
	// set openapi infos
	setupOpenApiStaticConfig()

	basePath := path.Join("..", "..")
	filePath := path.Join(basePath, "docs", "openapi.json")

	// generate openapi file
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	doc := docs.SwaggerInfo.ReadDoc()
	doc, err = removeHost(doc)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(doc)
	if err != nil {
		panic(err)
	}

	fmt.Println("generated openapi.json")

	// convert to OpenApi3
	if err := toOpenApi3(filePath, filePath); err != nil {
		fmt.Printf("converting '%s' from openapi v2 to v3 failed\n", filePath)
		panic(err)
	}
}

func removeHost(jsonIn string) (string, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonIn), &m); err != nil {
		return "", err
	}
	delete(m, "host")
	raw, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func toOpenApi3(input, output string) error {
	data2, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var doc2 openapi2.T
	err = json.Unmarshal(data2, &doc2)
	if err != nil {
		return fmt.Errorf("unmarshal input: %w", err)
	}

	doc3, err := openapi2conv.ToV3(&doc2)
	if err != nil {
		return fmt.Errorf("convert openapi v2 to v3: %w", err)
	}
	err = doc3.Validate(context.Background())
	if err != nil {
		return err
	}

	data, err := json.Marshal(doc3)
	if err != nil {
		return fmt.Errorf("Marshal converted: %w", err)
	}

	if err = os.WriteFile(output, data, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
