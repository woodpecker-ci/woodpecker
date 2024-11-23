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

package datastore

import (
	"xorm.io/builder"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

const orderVariablesBy = "name"

func (s storage) VariableFind(repo *model.Repo, name string) (*model.Variable, error) {
	variable := new(model.Variable)
	return variable, wrapGet(s.engine.Where(
		builder.Eq{"repo_id": repo.ID, "name": name},
	).Get(variable))
}

func (s storage) VariableList(repo *model.Repo, includeGlobalAndOrgVariables bool, p *model.ListOptions) ([]*model.Variable, error) {
	var variables []*model.Variable
	var cond builder.Cond = builder.Eq{"repo_id": repo.ID}
	if includeGlobalAndOrgVariables {
		cond = cond.Or(builder.Eq{"org_id": repo.OrgID}).
			Or(builder.And(builder.Eq{"org_id": 0}, builder.Eq{"repo_id": 0}))
	}
	return variables, s.paginate(p).Where(cond).OrderBy(orderVariablesBy).Find(&variables)
}

func (s storage) VariableListAll() ([]*model.Variable, error) {
	var variables []*model.Variable
	return variables, s.engine.Find(&variables)
}

func (s storage) VariableCreate(variable *model.Variable) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(variable)
	return err
}

func (s storage) VariableUpdate(variable *model.Variable) error {
	_, err := s.engine.ID(variable.ID).AllCols().Update(variable)
	return err
}

func (s storage) VariableDelete(variable *model.Variable) error {
	return wrapDelete(s.engine.ID(variable.ID).Delete(new(model.Variable)))
}

func (s storage) OrgVariableFind(orgID int64, name string) (*model.Variable, error) {
	variable := new(model.Variable)
	return variable, wrapGet(s.engine.Where(
		builder.Eq{"org_id": orgID, "name": name},
	).Get(variable))
}

func (s storage) OrgVariableList(orgID int64, p *model.ListOptions) ([]*model.Variable, error) {
	variables := make([]*model.Variable, 0)
	return variables, s.paginate(p).Where("org_id = ?", orgID).OrderBy(orderVariablesBy).Find(&variables)
}

func (s storage) GlobalVariableFind(name string) (*model.Variable, error) {
	variable := new(model.Variable)
	return variable, wrapGet(s.engine.Where(
		builder.Eq{"org_id": 0, "repo_id": 0, "name": name},
	).Get(variable))
}

func (s storage) GlobalVariableList(p *model.ListOptions) ([]*model.Variable, error) {
	variables := make([]*model.Variable, 0)
	return variables, s.paginate(p).Where(
		builder.Eq{"org_id": 0, "repo_id": 0},
	).OrderBy(orderVariablesBy).Find(&variables)
}
