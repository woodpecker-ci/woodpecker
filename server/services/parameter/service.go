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

package parameter

import "go.woodpecker-ci.org/woodpecker/v3/server/model"

//go:generate mockery --name Service --output mocks --case underscore

// Service defines a service for managing parameters.
type Service interface {
	// Repository parameters
	ParameterFind(*model.Repo, string) (*model.Parameter, error)
	ParameterFindByID(*model.Repo, int64) (*model.Parameter, error)
	ParameterFindByNameAndBranch(repo *model.Repo, name string, branch string) (*model.Parameter, error)
	ParameterList(*model.Repo) ([]*model.Parameter, error)
	ParameterCreate(*model.Repo, *model.Parameter) error
	ParameterUpdate(*model.Repo, *model.Parameter) error
	ParameterDelete(*model.Repo, string) error
	ParameterDeleteByID(*model.Repo, int64) error
}
