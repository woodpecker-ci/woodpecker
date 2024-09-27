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

package main

import (
	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/docs"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

// Generate docs/swagger.json via:
//go:generate go run woodpecker_docs_gen.go swagger.go
//go:generate go run github.com/getkin/kin-openapi/cmd/validate@latest ../../docs/swagger.json

// setupSwaggerStaticConfig initializes static content only (contacts, title and description)
// for dynamic configuration of e.g. hostname, etc. see router.setupSwaggerConfigAndRoutes
//
//	@contact.name	Woodpecker CI Community
//	@contact.url	https://woodpecker-ci.org/
func setupSwaggerStaticConfig() {
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.InfoInstanceName = "api"
	docs.SwaggerInfo.Title = "Woodpecker CI API"
	docs.SwaggerInfo.Version = version.String()
	docs.SwaggerInfo.Description = "Woodpecker is a simple, yet powerful CI/CD engine with great extensibility.\n" +
		"To get a personal access token (PAT) for authentication, please log in your Woodpecker server,\n" +
		"and go to you personal profile page, by clicking the user icon at the top right."
}
