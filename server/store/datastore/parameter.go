// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func (s storage) ParameterFind(repo *model.Repo, name string) (*model.Parameter, error) {
	parameter := new(model.Parameter)
	return parameter, wrapGet(s.engine.Where("parameter_repo_id = ? AND parameter_name = ?", repo.ID, name).Get(parameter))
}

func (s storage) ParameterFindByID(repo *model.Repo, id int64) (*model.Parameter, error) {
	parameter := new(model.Parameter)
	return parameter, wrapGet(s.engine.Where("parameter_repo_id = ? AND parameter_id = ?", repo.ID, id).Get(parameter))
}

func (s storage) ParameterFindByNameAndBranch(repo *model.Repo, name string, branch string) (*model.Parameter, error) {
	parameter := new(model.Parameter)
	return parameter, wrapGet(s.engine.Where("parameter_repo_id = ? AND parameter_name = ? AND parameter_branch = ?", repo.ID, name, branch).Get(parameter))
}

func (s storage) ParameterList(repo *model.Repo) ([]*model.Parameter, error) {
	var parameters []*model.Parameter
	return parameters, s.engine.Where("parameter_repo_id = ?", repo.ID).OrderBy("parameter_name").Find(&parameters)
}

func (s storage) ParameterCreate(repo *model.Repo, parameter *model.Parameter) error {
	if err := parameter.Validate(); err != nil {
		return err
	}
	parameter.RepoID = repo.ID
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(parameter)
	return err
}

func (s storage) ParameterUpdate(repo *model.Repo, parameter *model.Parameter) error {
	if err := parameter.Validate(); err != nil {
		return err
	}
	parameter.RepoID = repo.ID
	_, err := s.engine.Where("parameter_repo_id = ? AND parameter_id = ?", repo.ID, parameter.ID).
		AllCols().
		Update(parameter)
	return err
}

func (s storage) ParameterDelete(repo *model.Repo, name string) error {
	_, err := s.engine.Where("parameter_repo_id = ? AND parameter_id = ?", repo.ID, name).
		Delete(&model.Parameter{})
	return err
}

func (s storage) ParameterDeleteByID(repo *model.Repo, id int64) error {
	_, err := s.engine.Where("parameter_repo_id = ? AND parameter_id = ?", repo.ID, id).
		Delete(&model.Parameter{})
	return err
}
