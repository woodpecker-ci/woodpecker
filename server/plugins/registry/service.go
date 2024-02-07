package registry

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

// Service defines a service for managing registries.
type Service interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
	RegistryCreate(*model.Repo, *model.Registry) error
	RegistryUpdate(*model.Repo, *model.Registry) error
	RegistryDelete(*model.Repo, string) error
}

// ReadOnlyService defines a service for managing registries.
type ReadOnlyService interface {
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo, *model.ListOptions) ([]*model.Registry, error)
}
