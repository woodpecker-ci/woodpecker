package registry

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type combined struct {
	registries   []model.ReadOnlyRegistryService
	mainRegistry model.RegistryService
}

func NewCombined(mainRegistry model.RegistryService, registries ...model.ReadOnlyRegistryService) model.RegistryService {
	registries = append(registries, mainRegistry)
	return &combined{
		registries:   registries,
		mainRegistry: mainRegistry,
	}
}

func (c combined) RegistryFind(ctx context.Context, repo *model.Repo, name string) (*model.Registry, error) {
	for _, registry := range c.registries {
		res, err := registry.RegistryFind(ctx, repo, name)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func (c combined) RegistryList(ctx context.Context, repo *model.Repo) ([]*model.Registry, error) {
	var registries []*model.Registry
	for _, registry := range c.registries {
		list, err := registry.RegistryList(ctx, repo)
		if err != nil {
			return nil, err
		}
		registries = append(registries, list...)
	}
	return registries, nil
}

func (c combined) RegistryCreate(ctx context.Context, repo *model.Repo, registry *model.Registry) error {
	return c.mainRegistry.RegistryCreate(ctx, repo, registry)
}

func (c combined) RegistryUpdate(ctx context.Context, repo *model.Repo, registry *model.Registry) error {
	return c.mainRegistry.RegistryUpdate(ctx, repo, registry)
}

func (c combined) RegistryDelete(ctx context.Context, repo *model.Repo, name string) error {
	return c.mainRegistry.RegistryDelete(ctx, repo, name)
}
