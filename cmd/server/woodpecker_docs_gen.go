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
	"encoding/json"
	"os"
	"path"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/docs"

	"github.com/tidwall/pretty"
)

func main() {
	// set swagger infos
	setupSwaggerStaticConfig()

	// generate swagger file
	f, err := os.Create(path.Join("..", "..", "docs", "swagger.json"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	doc := docs.SwaggerInfo.ReadDoc()
	doc = removeHost(doc)
	_, err = f.WriteString(doc)
	if err != nil {
		panic(err)
	}
}

func removeHost(jsonIn string) string {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonIn), &m); err != nil {
		panic(err)
	}
	delete(m, "host")
	raw, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(pretty.Pretty(raw))
}
