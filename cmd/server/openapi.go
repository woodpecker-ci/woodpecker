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
	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/openapi"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

// Generate docs/openapi.json via:
//go:generate go run github.com/swaggo/swag/cmd/swag init -g cmd/server/openapi.go --outputTypes go -output openapi -d ../../
//go:generate go run openapi_json_gen.go openapi.go
//go:generate go run github.com/getkin/kin-openapi/cmd/validate@latest ../../docs/openapi.json

// setupOpenApiStaticConfig initializes static content (version) for the OpenAPI config.
//
//	@title			Woodpecker CI API
//	@description	Woodpecker is a simple, yet powerful CI/CD engine with great extensibility.
//	@description	To get a personal access token (PAT) for authentication, please log in your Woodpecker server,
//	@description	and go to you personal profile page, by clicking the user icon at the top right.
//	@BasePath		/api
//	@contact.name	Woodpecker CI
//	@contact.url	https://woodpecker-ci.org/
func setupOpenApiStaticConfig() {
	openapi.SwaggerInfo.Version = version.String()
}
