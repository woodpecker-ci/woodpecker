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

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type db struct {
	store store.Store
}

// NewDB returns a new local variable service.
func NewDB(store store.Store) Service {
	return &db{store: store}
}

func (d *db) VariableFind(repo *model.Repo, name string) (*model.Variable, error) {
	return d.store.VariableFind(repo, name)
}

func (d *db) VariableList(repo *model.Repo, p *model.ListOptions) ([]*model.Variable, error) {
	return d.store.VariableList(repo, false, p)
}

func (d *db) VariableListPipeline(repo *model.Repo, _ *model.Pipeline) ([]*model.Variable, error) {
	s, err := d.store.VariableList(repo, true, &model.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Return only variables with unique name
	// Priority order in case of duplicate names are repository, user/organization, global
	variables := make([]*model.Variable, 0, len(s))
	uniq := make(map[string]struct{})
	for _, condition := range []struct {
		IsRepository   bool
		IsOrganization bool
		IsGlobal       bool
	}{
		{IsRepository: true},
		{IsOrganization: true},
		{IsGlobal: true},
	} {
		for _, variable := range s {
			if variable.IsRepository() != condition.IsRepository || variable.IsOrganization() != condition.IsOrganization || variable.IsGlobal() != condition.IsGlobal {
				continue
			}
			if _, ok := uniq[variable.Name]; ok {
				continue
			}
			uniq[variable.Name] = struct{}{}
			variables = append(variables, variable)
		}
	}
	return variables, nil
}

func (d *db) VariableCreate(_ *model.Repo, in *model.Variable) error {
	return d.store.VariableCreate(in)
}

func (d *db) VariableUpdate(_ *model.Repo, in *model.Variable) error {
	return d.store.VariableUpdate(in)
}

func (d *db) VariableDelete(repo *model.Repo, name string) error {
	variable, err := d.store.VariableFind(repo, name)
	if err != nil {
		return err
	}
	return d.store.VariableDelete(variable)
}

func (d *db) OrgVariableFind(owner int64, name string) (*model.Variable, error) {
	return d.store.OrgVariableFind(owner, name)
}

func (d *db) OrgVariableList(owner int64, p *model.ListOptions) ([]*model.Variable, error) {
	return d.store.OrgVariableList(owner, p)
}

func (d *db) OrgVariableCreate(_ int64, in *model.Variable) error {
	return d.store.VariableCreate(in)
}

func (d *db) OrgVariableUpdate(_ int64, in *model.Variable) error {
	return d.store.VariableUpdate(in)
}

func (d *db) OrgVariableDelete(owner int64, name string) error {
	variable, err := d.store.OrgVariableFind(owner, name)
	if err != nil {
		return err
	}
	return d.store.VariableDelete(variable)
}

func (d *db) GlobalVariableFind(owner string) (*model.Variable, error) {
	return d.store.GlobalVariableFind(owner)
}

func (d *db) GlobalVariableList(p *model.ListOptions) ([]*model.Variable, error) {
	return d.store.GlobalVariableList(p)
}

func (d *db) GlobalVariableCreate(in *model.Variable) error {
	return d.store.VariableCreate(in)
}

func (d *db) GlobalVariableUpdate(in *model.Variable) error {
	return d.store.VariableUpdate(in)
}

func (d *db) GlobalVariableDelete(name string) error {
	variable, err := d.store.GlobalVariableFind(name)
	if err != nil {
		return err
	}
	return d.store.VariableDelete(variable)
}
