// Copyright 2021 Woodpecker Authors
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

const orderRegistriesBy = "id"

func (s storage) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	reg := new(model.Registry)
	return reg, wrapGet(s.engine.Where(
		builder.Eq{"repo_id": repo.ID, "address": addr},
	).Get(reg))
}

func (s storage) RegistryList(repo *model.Repo, includeGlobalAndOrg bool, p *model.ListOptions) ([]*model.Registry, error) {
	var regs []*model.Registry
	var cond builder.Cond = builder.Eq{"repo_id": repo.ID}
	if includeGlobalAndOrg {
		cond = cond.Or(builder.Eq{"org_id": repo.OrgID}).
			Or(builder.And(builder.Eq{"org_id": 0}, builder.Eq{"repo_id": 0}))
	}
	return regs, s.paginate(p).Where(cond).OrderBy(orderRegistriesBy).Find(&regs)
}

func (s storage) RegistryListAll() ([]*model.Registry, error) {
	var registries []*model.Registry
	return registries, s.engine.Find(&registries)
}

func (s storage) RegistryCreate(registry *model.Registry) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(registry)
	return err
}

func (s storage) RegistryUpdate(registry *model.Registry) error {
	_, err := s.engine.ID(registry.ID).AllCols().Update(registry)
	return err
}

func (s storage) RegistryDelete(registry *model.Registry) error {
	return wrapDelete(s.engine.ID(registry.ID).Delete(new(model.Registry)))
}

func (s storage) OrgRegistryFind(orgID int64, name string) (*model.Registry, error) {
	registry := new(model.Registry)
	return registry, wrapGet(s.engine.Where(
		builder.Eq{"org_id": orgID, "address": name},
	).Get(registry))
}

func (s storage) OrgRegistryList(orgID int64, p *model.ListOptions) ([]*model.Registry, error) {
	registries := make([]*model.Registry, 0)
	return registries, s.paginate(p).Where("org_id = ?", orgID).OrderBy(orderRegistriesBy).Find(&registries)
}

func (s storage) GlobalRegistryFind(name string) (*model.Registry, error) {
	registry := new(model.Registry)
	return registry, wrapGet(s.engine.Where(
		builder.Eq{"org_id": 0, "repo_id": 0, "address": name},
	).Get(registry))
}

func (s storage) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	registries := make([]*model.Registry, 0)
	return registries, s.paginate(p).Where(
		builder.Eq{"org_id": 0, "repo_id": 0},
	).OrderBy(orderRegistriesBy).Find(&registries)
}
