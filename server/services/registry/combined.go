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
	"errors"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

type combined struct {
	registries []ReadOnlyService
	dbRegistry Service
}

func NewCombined(dbRegistry Service, registries ...ReadOnlyService) Service {
	registries = append(registries, dbRegistry)
	return &combined{
		registries: registries,
		dbRegistry: dbRegistry,
	}
}

func (c *combined) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	return c.dbRegistry.RegistryFind(repo, addr)
}

func (c *combined) RegistryList(repo *model.Repo, p *model.ListOptions) ([]*model.Registry, error) {
	return c.dbRegistry.RegistryList(repo, p)
}

func (c *combined) RegistryListPipeline(repo *model.Repo, pipeline *model.Pipeline) ([]*model.Registry, error) {
	dbRegistries, err := c.dbRegistry.RegistryListPipeline(repo, pipeline)
	if err != nil {
		return nil, err
	}

	registries := make([]*model.Registry, 0, len(dbRegistries))
	exists := make(map[string]struct{}, len(dbRegistries))

	// Assign database stored registries to the map to avoid duplicates
	// from the combined registries so to prioritize ones in database.
	for _, reg := range dbRegistries {
		exists[reg.Address] = struct{}{}
	}

	for _, registry := range c.registries {
		list, err := registry.GlobalRegistryList(&model.ListOptions{All: true})
		if err != nil {
			return nil, err
		}
		for _, reg := range list {
			if _, ok := exists[reg.Address]; ok {
				continue
			}
			exists[reg.Address] = struct{}{}
			registries = append(registries, reg)
		}
	}

	return append(registries, dbRegistries...), nil
}

func (c *combined) RegistryCreate(repo *model.Repo, registry *model.Registry) error {
	return c.dbRegistry.RegistryCreate(repo, registry)
}

func (c *combined) RegistryUpdate(repo *model.Repo, registry *model.Registry) error {
	return c.dbRegistry.RegistryUpdate(repo, registry)
}

func (c *combined) RegistryDelete(repo *model.Repo, addr string) error {
	return c.dbRegistry.RegistryDelete(repo, addr)
}

func (c *combined) OrgRegistryFind(owner int64, addr string) (*model.Registry, error) {
	return c.dbRegistry.OrgRegistryFind(owner, addr)
}

func (c *combined) OrgRegistryList(owner int64, p *model.ListOptions) ([]*model.Registry, error) {
	return c.dbRegistry.OrgRegistryList(owner, p)
}

func (c *combined) OrgRegistryCreate(owner int64, registry *model.Registry) error {
	return c.dbRegistry.OrgRegistryCreate(owner, registry)
}

func (c *combined) OrgRegistryUpdate(owner int64, registry *model.Registry) error {
	return c.dbRegistry.OrgRegistryUpdate(owner, registry)
}

func (c *combined) OrgRegistryDelete(owner int64, addr string) error {
	return c.dbRegistry.OrgRegistryDelete(owner, addr)
}

func (c *combined) GlobalRegistryFind(addr string) (*model.Registry, error) {
	registry, err := c.dbRegistry.GlobalRegistryFind(addr)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		return nil, err
	}
	if registry != nil {
		return registry, nil
	}
	for _, reg := range c.registries {
		if registry, err := reg.GlobalRegistryFind(addr); err == nil {
			return registry, nil
		}
	}
	return nil, types.RecordNotExist
}

func (c *combined) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	dbRegistries, err := c.dbRegistry.GlobalRegistryList(&model.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	registries := make([]*model.Registry, 0, len(dbRegistries))
	exists := make(map[string]struct{}, len(dbRegistries))

	// Assign database stored registries to the map to avoid duplicates
	// from the combined registries so to prioritize ones in database.
	for _, reg := range dbRegistries {
		exists[reg.Address] = struct{}{}
	}

	for _, registry := range c.registries {
		list, err := registry.GlobalRegistryList(&model.ListOptions{All: true})
		if err != nil {
			return nil, err
		}
		for _, reg := range list {
			if _, ok := exists[reg.Address]; ok {
				continue
			}
			exists[reg.Address] = struct{}{}
			registries = append(registries, reg)
		}
	}

	return model.ApplyPagination(p, append(registries, dbRegistries...)), nil
}

func (c *combined) GlobalRegistryCreate(registry *model.Registry) error {
	return c.dbRegistry.GlobalRegistryCreate(registry)
}

func (c *combined) GlobalRegistryUpdate(registry *model.Registry) error {
	return c.dbRegistry.GlobalRegistryUpdate(registry)
}

func (c *combined) GlobalRegistryDelete(addr string) error {
	return c.dbRegistry.GlobalRegistryDelete(addr)
}
