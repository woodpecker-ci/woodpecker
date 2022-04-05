package secrets

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type builtin struct {
	store model.SecretStore
}

// New returns a new local secret service.
func NewBuiltin(store model.SecretStore) model.SecretService {
	return &builtin{store: store}
}

func (b *builtin) SecretFind(ctx context.Context, repo *model.Repo, name string) (*model.Secret, error) {
	return b.store.SecretFind(repo, name)
}

func (b *builtin) SecretList(ctx context.Context, repo *model.Repo) ([]*model.Secret, error) {
	return b.store.SecretList(repo)
}

func (b *builtin) SecretCreate(ctx context.Context, repo *model.Repo, in *model.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) SecretUpdate(ctx context.Context, repo *model.Repo, in *model.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) SecretDelete(ctx context.Context, repo *model.Repo, name string) error {
	secret, err := b.store.SecretFind(repo, name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}
