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
	"go.woodpecker-ci.org/woodpecker/server/model"
)

type combined struct {
	registries []model.ReadOnlyRegistryService
	dbRegistry model.RegistryService
}

func Combined(dbRegistry model.RegistryService, registries ...model.ReadOnlyRegistryService) model.RegistryService {
	registries = append(registries, dbRegistry)
	return &combined{
		registries: registries,
		dbRegistry: dbRegistry,
	}
}

func (c combined) RegistryFind(repo *model.Repo, name string) (*model.Registry, error) {
	for _, registry := range c.registries {
		res, err := registry.RegistryFind(repo, name)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func (c combined) RegistryList(repo *model.Repo, p *model.ListOptions) ([]*model.Registry, error) {
	var registries []*model.Registry
	for _, registry := range c.registries {
		list, err := registry.RegistryList(repo, &model.ListOptions{All: true})
		if err != nil {
			return nil, err
		}
		registries = append(registries, list...)
	}
	return model.ApplyPagination(p, registries), nil
}

func (c combined) RegistryCreate(repo *model.Repo, registry *model.Registry) error {
	return c.dbRegistry.RegistryCreate(repo, registry)
}

func (c combined) RegistryUpdate(repo *model.Repo, registry *model.Registry) error {
	return c.dbRegistry.RegistryUpdate(repo, registry)
}

func (c combined) RegistryDelete(repo *model.Repo, name string) error {
	return c.dbRegistry.RegistryDelete(repo, name)
}
