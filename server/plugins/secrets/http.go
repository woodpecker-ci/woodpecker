package secrets

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

type http struct {
	endpoint string
	secret   string
}

// New returns a new local secret service.
func NewHTTP(endpoint, secret string) model.SecretService {
	return &http{endpoint: endpoint, secret: secret}
}

func FromRepo(repo *model.Repo) model.SecretService {
	return NewHTTP(repo.SecretEndpoint, repo.ExtensionSecret)
}

func (b *http) SecretFind(ctx context.Context, repo *model.Repo, name string) (secret *model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "GET", path, b.secret, nil, secret)
	return secret, err
}

func (b *http) SecretList(ctx context.Context, repo *model.Repo) (secrets []*model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "GET", path, b.secret, nil, secrets)
	return secrets, err
}

func (b *http) SecretCreate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "POST", path, b.secret, in, nil)
	return err
}

func (b *http) SecretUpdate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "PUT", path, b.secret, in, nil)
	return err
}

func (b *http) SecretDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "DELETE", path, b.secret, nil, nil)
	return err
}
