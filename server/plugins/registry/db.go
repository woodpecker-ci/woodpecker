package registry

import (
	"github.com/woodpecker-ci/woodpecker/model"
)

type db struct {
	store model.RegistryStore
}

// New returns a new local registry service.
func New(store model.RegistryStore) model.RegistryService {
	return &db{store}
}

func (b *db) RegistryFind(repo *model.Repo, name string) (*model.Registry, error) {
	return b.store.RegistryFind(repo, name)
}

func (b *db) RegistryList(repo *model.Repo) ([]*model.Registry, error) {
	return b.store.RegistryList(repo)
}

func (b *db) RegistryCreate(repo *model.Repo, in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) RegistryUpdate(repo *model.Repo, in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) RegistryDelete(repo *model.Repo, addr string) error {
	registry, err := b.RegistryFind(repo, addr)
	if err != nil {
		return err
	}
	return b.store.RegistryDelete(registry)
}
