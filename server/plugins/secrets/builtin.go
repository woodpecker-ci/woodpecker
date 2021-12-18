package secrets

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type builtin struct {
	context.Context
	store model.SecretStore
}

// New returns a new local secret service.
func New(ctx context.Context, store model.SecretStore) model.SecretService {
	return &builtin{store: store, Context: ctx}
}

func (b *builtin) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	return b.store.SecretFind(repo, name)
}

func (b *builtin) SecretList(repo *model.Repo) ([]*model.Secret, error) {
	return b.store.SecretList(repo)
}

func (b *builtin) SecretListBuild(repo *model.Repo, build *model.Build) ([]*model.Secret, error) {
	return b.store.SecretList(repo)
}

func (b *builtin) SecretCreate(repo *model.Repo, in *model.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) SecretUpdate(repo *model.Repo, in *model.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) SecretDelete(repo *model.Repo, name string) error {
	secret, err := b.store.SecretFind(repo, name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}
