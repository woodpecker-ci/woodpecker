package registry

import (
	"errors"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

type combined struct {
	registries []model.ReadOnlyRegistryService
	dbRegistry model.RegistryService
}

func Combined(dbRegistry model.RegistryService, registries ...model.ReadOnlyRegistryService) model.RegistryService {
	return &combined{
		registries: registries,
		dbRegistry: dbRegistry,
	}
}

func (c combined) RegistryListPipeline(repo *model.Repo, pipeline *model.Pipeline) ([]*model.Registry, error) {
	// Prioritize registries from the database
	s, err := c.dbRegistry.RegistryListPipeline(repo, pipeline)
	if err != nil {
		return nil, err
	}

	registries := make([]*model.Registry, 0, len(s))
	uniq := make(map[string]struct{}, len(s))
	for _, registry := range registries {
		uniq[registry.Address] = struct{}{}
		registries = append(registries, registry)
	}

	// Add registries from read-only sources that are not in databse
	for _, reg := range c.registries {
		regs, err := reg.RegistryList()
		if err != nil {
			return nil, err
		}
		for _, registry := range regs {
			if _, ok := uniq[registry.Address]; ok {
				continue
			}
			uniq[registry.Address] = struct{}{}
			registries = append(registries, registry)
		}
	}
	return registries, nil
}

func (c combined) RegistryFind(repo *model.Repo, name string) (*model.Registry, error) {
	return c.dbRegistry.RegistryFind(repo, name)
}

func (c combined) RegistryList(repo *model.Repo) ([]*model.Registry, error) {
	return c.dbRegistry.RegistryList(repo)
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

func (c combined) OrgRegistryFind(owner, addr string) (*model.Registry, error) {
	return c.dbRegistry.OrgRegistryFind(owner, addr)
}

func (c combined) OrgRegistryList(owner string) ([]*model.Registry, error) {
	return c.dbRegistry.OrgRegistryList(owner)
}

func (c combined) OrgRegistryCreate(owner string, registry *model.Registry) error {
	return c.dbRegistry.OrgRegistryCreate(owner, registry)
}

func (c combined) OrgRegistryUpdate(owner string, registry *model.Registry) error {
	return c.dbRegistry.OrgRegistryUpdate(owner, registry)
}

func (c combined) OrgRegistryDelete(owner, addr string) error {
	return c.dbRegistry.OrgRegistryDelete(owner, addr)
}

func (c combined) GlobalRegistryFind(addr string) (*model.Registry, error) {
	registry, err := c.dbRegistry.GlobalRegistryFind(addr)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		return nil, err
	}
	if registry != nil {
		return registry, nil
	}
	for _, reg := range c.registries {
		if registry, err := reg.RegistryFind(addr); err == nil {
			return registry, nil
		}
	}
	return nil, types.RecordNotExist
}

func (c combined) GlobalRegistryList() ([]*model.Registry, error) {
	dbRegistries, err := c.dbRegistry.GlobalRegistryList()
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
		list, err := registry.RegistryList()
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

	// Append database stored registries to the end of the list.
	return append(registries, dbRegistries...), nil
}

func (c combined) GlobalRegistryCreate(registry *model.Registry) error {
	return c.dbRegistry.GlobalRegistryCreate(registry)
}

func (c combined) GlobalRegistryUpdate(registry *model.Registry) error {
	return c.dbRegistry.GlobalRegistryUpdate(registry)
}

func (c combined) GlobalRegistryDelete(addr string) error {
	return c.dbRegistry.GlobalRegistryDelete(addr)
}
