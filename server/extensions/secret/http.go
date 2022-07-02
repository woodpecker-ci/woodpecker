package secret

import (
	"context"
	"crypto"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/extensions/utils"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// New returns a new local secret service.
func NewHTTP(endpoint string, privateKey crypto.PrivateKey) SecretExtension {
	return &http{endpoint, privateKey}
}

func (b *http) SecretFind(ctx context.Context, repo *model.Repo, name string) (secret *model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "GET", path, b.privateKey, nil, secret)
	return secret, err
}

func (b *http) SecretList(ctx context.Context, repo *model.Repo) (secrets []*model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "GET", path, b.privateKey, nil, secrets)
	return secrets, err
}

func (b *http) SecretCreate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "POST", path, b.privateKey, in, nil)
	return err
}

func (b *http) SecretUpdate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "PUT", path, b.privateKey, in, nil)
	return err
}

func (b *http) SecretDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "DELETE", path, b.privateKey, nil, nil)
	return err
}
