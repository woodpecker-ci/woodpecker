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

// ************************************************************************************************
// This is a generator tool, to update the Markdown documentation for the woodpecker-ci.org website
// ************************************************************************************************

//go:build generate
// +build generate

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-swagger/go-swagger/generator"

	"github.com/woodpecker-ci/woodpecker/cmd/server/docs"
)

const restApiMarkdownInto = `
# REST API

Woodpecker offers a comprehensive REST API, so you can integrate easily with from and with other tools.

## API specification

Starting with Woodpecker v1.0.0 a Swagger v2 API specification is served by the Woodpecker Server.
The typical URL looks like "http://woodpecker-host/swagger/doc.json", where you can fetch the API specification.

## Swagger API UI

Starting with Woodpecker v1.0.0 a Swagger web user interface (UI) is served by the Woodpecker Server.
Typically, you can open "http://woodpecker-host/swagger/index.html" in your browser, to explore the API documentation.

# API endpoint summary

This is a summary of available API endpoints.
Please, keep in mind this documentation reflects latest development changes
and might differ from your used server version.
Its recommended to consult the Swagger API UI of your Woodpecker server,
where you also have the chance to do manual exploration and live testing.

`

func main() {
	setupSwaggerStaticConfig()

	specFile := createTempFileWithSwaggerSpec()
	markdown := generateTempMarkdown(specFile)

	f, err := os.Create(path.Join("..", "..", "docs", "docs", "20-usage", "90-rest-api.md"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(restApiMarkdownInto)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(readGeneratedMarkdownAndSkipIntro(markdown))
	if err != nil {
		panic(err)
	}
}

func createTempFileWithSwaggerSpec() string {
	f, err := os.Create(path.Join("..", "..", "docs", "swagger.json"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(docs.SwaggerInfo.ReadDoc())
	if err != nil {
		panic(err)
	}
	return f.Name()
}

func generateTempMarkdown(specFile string) string {
	// HINT: we MUST use underscores, since the library tends to rename things
	tempFile := fmt.Sprintf("woodpecker_api_%d.md", time.Now().UnixMilli())
	markdownFile := path.Join(os.TempDir(), tempFile)

	opts := generator.GenOpts{
		GenOptsCommon: generator.GenOptsCommon{
			Spec: specFile,
		},
	}
	// TODO: contrib upstream a GenerateMarkdown that use io.Reader and io.Writer
	err := generator.GenerateMarkdown(markdownFile, []string{}, []string{}, &opts)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(markdownFile)
	if err != nil {
		panic(err)
	}
	defer os.Remove(markdownFile)

	return string(data)
}

func readGeneratedMarkdownAndSkipIntro(markdown string) string {
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	sb := strings.Builder{}
	foundActualContentStart := false
	for scanner.Scan() {
		text := scanner.Text()
		foundActualContentStart = foundActualContentStart || (strings.HasPrefix(text, "##") && strings.Contains(strings.ToLower(text), "all endpoints"))
		if foundActualContentStart {
			sb.WriteString(text + "\n")
		}
	}
	return sb.String()
}
