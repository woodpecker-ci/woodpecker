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

func (b *db) RegistryListPipeline(repo *model.Repo, _ *model.Pipeline) ([]*model.Registry, error) {
	r, err := b.store.RegistryList(repo, true)
	if err != nil {
		return nil, err
	}

	// Return only registries with unique address
	// Priority order in case of duplicate addresses are repository, user/organization, global
	registries := make([]*model.Registry, 0, len(r))
	uniq := make(map[string]struct{})
	for _, cond := range []struct {
		Global       bool
		Organization bool
	}{
		{},
		{Organization: true},
		{Global: true},
	} {
		for _, registry := range r {
			if registry.Global() == cond.Global && registry.Organization() == cond.Organization {
				continue
			}
			if _, ok := uniq[registry.Address]; ok {
				continue
			}
			uniq[registry.Address] = struct{}{}
			registries = append(registries, registry)
		}
	}
	return registries, nil
}

func (b *db) RegistryFind(repo *model.Repo, name string) (*model.Registry, error) {
	return b.store.RegistryFind(repo, name)
}

func (b *db) RegistryList(repo *model.Repo) ([]*model.Registry, error) {
	return b.store.RegistryList(repo, false)
}

func (b *db) RegistryCreate(_ *model.Repo, in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) RegistryUpdate(_ *model.Repo, in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) RegistryDelete(repo *model.Repo, addr string) error {
	registry, err := b.store.RegistryFind(repo, addr)
	if err != nil {
		return err
	}
	return b.store.RegistryDelete(registry)
}

func (b *db) OrgRegistryFind(owner, addr string) (*model.Registry, error) {
	return b.store.OrgRegistryFind(owner, addr)
}

func (b *db) OrgRegistryList(owner string) ([]*model.Registry, error) {
	return b.store.OrgRegistryList(owner)
}

func (b *db) OrgRegistryCreate(_ string, in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) OrgRegistryUpdate(_ string, in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) OrgRegistryDelete(owner, addr string) error {
	registry, err := b.store.OrgRegistryFind(owner, addr)
	if err != nil {
		return err
	}
	return b.store.RegistryDelete(registry)
}

func (b *db) GlobalRegistryFind(addr string) (*model.Registry, error) {
	return b.store.GlobalRegistryFind(addr)
}

func (b *db) GlobalRegistryList() ([]*model.Registry, error) {
	return b.store.GlobalRegistryList()
}

func (b *db) GlobalRegistryCreate(in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *db) GlobalRegistryUpdate(in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *db) GlobalRegistryDelete(addr string) error {
	registry, err := b.store.GlobalRegistryFind(addr)
	if err != nil {
		return err
	}
	return b.store.RegistryDelete(registry)
}
