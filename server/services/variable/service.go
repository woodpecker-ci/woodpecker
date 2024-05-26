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

package variable

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

//go:generate mockery --name Service --output mocks --case underscore

// Service defines a service for managing variables.
type Service interface {
	VariableListPipeline(*model.Repo, *model.Pipeline) ([]*model.Variable, error)
	// Repository variables
	VariableFind(*model.Repo, string) (*model.Variable, error)
	VariableList(*model.Repo, *model.ListOptions) ([]*model.Variable, error)
	VariableCreate(*model.Repo, *model.Variable) error
	VariableUpdate(*model.Repo, *model.Variable) error
	VariableDelete(*model.Repo, string) error
	// Organization variables
	OrgVariableFind(int64, string) (*model.Variable, error)
	OrgVariableList(int64, *model.ListOptions) ([]*model.Variable, error)
	OrgVariableCreate(int64, *model.Variable) error
	OrgVariableUpdate(int64, *model.Variable) error
	OrgVariableDelete(int64, string) error
	// Global variables
	GlobalVariableFind(string) (*model.Variable, error)
	GlobalVariableList(*model.ListOptions) ([]*model.Variable, error)
	GlobalVariableCreate(*model.Variable) error
	GlobalVariableUpdate(*model.Variable) error
	GlobalVariableDelete(string) error
}
