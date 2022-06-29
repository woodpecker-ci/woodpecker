package registry

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type db struct {
	store model.RegistryStore
}

// New returns a new local registry service.
func NewBuiltin(store model.RegistryStore) RegistryExtension {
	return &db{store}
}

func (b *db) RegistryFind(ctx context.Context, repo *model.Repo, name string) (*model.Registry, error) {
	return b.store.RegistryFind(repo, name)
}

func (b *db) RegistryList(ctx context.Context, repo *model.Repo) ([]*model.Registry, error) {
	return b.store.RegistryList(repo)
}

func (b *db) RegistryCreate(ctx context.Context, repo *model.Repo, in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) RegistryUpdate(ctx context.Context, repo *model.Repo, in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) RegistryDelete(ctx context.Context, repo *model.Repo, addr string) error {
	return b.store.RegistryDelete(repo, addr)
}
