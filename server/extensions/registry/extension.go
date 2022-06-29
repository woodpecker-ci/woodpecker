package registry

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// RegistryExtension defines a service for managing registries.
type RegistryExtension interface {
	RegistryFind(context.Context, *model.Repo, string) (*model.Registry, error)
	RegistryList(context.Context, *model.Repo) ([]*model.Registry, error)
	RegistryCreate(context.Context, *model.Repo, *model.Registry) error
	RegistryUpdate(context.Context, *model.Repo, *model.Registry) error
	RegistryDelete(context.Context, *model.Repo, string) error
}

// ReadOnlyRegistryExtension defines a service for managing registries.
type ReadOnlyRegistryExtension interface {
	RegistryFind(context.Context, *model.Repo, string) (*model.Registry, error)
	RegistryList(context.Context, *model.Repo) ([]*model.Registry, error)
}
