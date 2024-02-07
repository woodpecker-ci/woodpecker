package registry

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

// RegistryService defines a service for managing registries.
type RegistryService interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
	RegistryCreate(*model.Repo, *model.Registry) error
	RegistryUpdate(*model.Repo, *model.Registry) error
	RegistryDelete(*model.Repo, string) error
}

// ReadOnlyRegistryService defines a service for managing registries.
type ReadOnlyRegistryService interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
}
