package secrets

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// SecretExtension defines a service for managing secrets.
type SecretExtension interface {
	SecretFind(context.Context, *model.Repo, string) (*model.Secret, error)
	SecretList(context.Context, *model.Repo) ([]*model.Secret, error)
	SecretCreate(context.Context, *model.Repo, *model.Secret) error
	SecretUpdate(context.Context, *model.Repo, *model.Secret) error
	SecretDelete(context.Context, *model.Repo, string) error
}
