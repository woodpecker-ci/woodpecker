package secret

import (
	"context"
	"crypto"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/utils"
)

type http struct {
	Service
	endpoint   string
	privateKey crypto.PrivateKey
}

// New returns a new local secret service.
func NewHTTP(parent Service, endpoint string, privateKey crypto.PrivateKey) Service {
	return &http{parent, endpoint, privateKey}
}

func (h *http) SecretList(ctx context.Context, repo *model.Repo, p *model.ListOptions) (secrets []*model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%d?%s", h.endpoint, repo.ID, p.Encode())
	_, err = utils.Send(ctx, "GET", path, h.privateKey, nil, &secrets)
	return secrets, err
}

func (h *http) SecretFind(ctx context.Context, repo *model.Repo, name string) (secret *model.Secret, err error) {
	path := fmt.Sprintf("%s/secrets/%d/%s", h.endpoint, repo.ID, name)
	_, err = utils.Send(ctx, "GET", path, h.privateKey, nil, secret)
	return secret, err
}

func (h *http) SecretCreate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%d", h.endpoint, repo.ID)
	_, err = utils.Send(ctx, "POST", path, h.privateKey, in, nil)
	return err
}

func (h *http) SecretUpdate(ctx context.Context, repo *model.Repo, in *model.Secret) (err error) {
	path := fmt.Sprintf("%s/secrets/%d/%s", h.endpoint, repo.ID, repo.Name)
	_, err = utils.Send(ctx, "PUT", path, h.privateKey, in, nil)
	return err
}

func (h *http) SecretDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/secrets/%d/%s", h.endpoint, repo.ID, name)
	_, err = utils.Send(ctx, "DELETE", path, h.privateKey, nil, nil)
	return err
}
