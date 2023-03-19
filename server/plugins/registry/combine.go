package registry

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
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

// RegistryList TODO respect model.PaginationData
func (c combined) RegistryList(repo *model.Repo, _ *model.PaginationData) ([]*model.Registry, error) {
	var registries []*model.Registry
	for _, registry := range c.registries {
		// TODO get all
		list, err := registry.RegistryList(repo, &model.PaginationData{Page: 1, PerPage: 50})
		if err != nil {
			return nil, err
		}
		registries = append(registries, list...)
	}
	return registries, nil
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
