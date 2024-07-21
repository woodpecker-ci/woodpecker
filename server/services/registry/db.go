// Copyright 2023 Woodpecker Authors
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

package registry

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type db struct {
	store store.Store
}

// New returns a new local registry service.
func NewDB(store store.Store) Service {
	return &db{store}
}

func (d *db) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	return d.store.RegistryFind(repo, addr)
}

func (d *db) RegistryList(repo *model.Repo, p *model.ListOptions) ([]*model.Registry, error) {
	return d.store.RegistryList(repo, false, p)
}

func (d *db) RegistryListPipeline(repo *model.Repo, _ *model.Pipeline) ([]*model.Registry, error) {
	r, err := d.store.RegistryList(repo, true, &model.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Return only registries with unique address
	// Priority order in case of duplicate addresses are repository, user/organization, global
	registries := make([]*model.Registry, 0, len(r))
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
		for _, registry := range r {
			if registry.IsRepository() != condition.IsRepository || registry.IsOrganization() != condition.IsOrganization || registry.IsGlobal() != condition.IsGlobal {
				continue
			}
			if _, ok := uniq[registry.Address]; ok {
				continue
			}
			uniq[registry.Address] = struct{}{}
			registries = append(registries, registry)
		}
	}
	return registries, nil
}

func (d *db) RegistryCreate(_ *model.Repo, in *model.Registry) error {
	return d.store.RegistryCreate(in)
}

func (d *db) RegistryUpdate(_ *model.Repo, in *model.Registry) error {
	return d.store.RegistryUpdate(in)
}

func (d *db) RegistryDelete(repo *model.Repo, addr string) error {
	registry, err := d.store.RegistryFind(repo, addr)
	if err != nil {
		return err
	}
	return d.store.RegistryDelete(registry)
}

func (d *db) OrgRegistryFind(owner int64, name string) (*model.Registry, error) {
	return d.store.OrgRegistryFind(owner, name)
}

func (d *db) OrgRegistryList(owner int64, p *model.ListOptions) ([]*model.Registry, error) {
	return d.store.OrgRegistryList(owner, p)
}

func (d *db) OrgRegistryCreate(_ int64, in *model.Registry) error {
	return d.store.RegistryCreate(in)
}

func (d *db) OrgRegistryUpdate(_ int64, in *model.Registry) error {
	return d.store.RegistryUpdate(in)
}

func (d *db) OrgRegistryDelete(owner int64, addr string) error {
	registry, err := d.store.OrgRegistryFind(owner, addr)
	if err != nil {
		return err
	}
	return d.store.RegistryDelete(registry)
}

func (d *db) GlobalRegistryFind(addr string) (*model.Registry, error) {
	return d.store.GlobalRegistryFind(addr)
}

func (d *db) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	return d.store.GlobalRegistryList(p)
}

func (d *db) GlobalRegistryCreate(in *model.Registry) error {
	return d.store.RegistryCreate(in)
}

func (d *db) GlobalRegistryUpdate(in *model.Registry) error {
	return d.store.RegistryUpdate(in)
}

func (d *db) GlobalRegistryDelete(addr string) error {
	registry, err := d.store.GlobalRegistryFind(addr)
	if err != nil {
		return err
	}
	return d.store.RegistryDelete(registry)
}
