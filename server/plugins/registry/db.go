package registry

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
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

func (b *db) RegistryList(repo *model.Repo, p *model.PaginationData) ([]*model.Registry, error) {
	return b.store.RegistryList(repo, p)
}

func (b *db) RegistryCreate(_ *model.Repo, in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) RegistryUpdate(_ *model.Repo, in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) RegistryDelete(repo *model.Repo, addr string) error {
	return b.store.RegistryDelete(repo, addr)
}
