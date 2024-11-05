// Copyright 2024 Woodpecker Authors
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

package environment

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

//go:generate mockery --name Service --output mocks --case underscore

// Service defines a service for managing environment variables.
type Service interface {
	EnvironList(*model.Repo) ([]*model.Environ, error)
}
