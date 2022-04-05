package registry

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
func NewHTTP(endpoint, secret string) model.RegistryService {
	return &http{endpoint: endpoint, secret: secret}
}

func FromRepo(repo *model.Repo) model.RegistryService {
	return NewHTTP(repo.RegistryEndpoint, repo.ExtensionSecret)
}

func (b *http) RegistryFind(ctx context.Context, repo *model.Repo, name string) (registry *model.Registry, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(context.TODO(), "GET", path, b.secret, nil, registry)
	return registry, err
}

func (b *http) RegistryList(ctx context.Context, repo *model.Repo) (registries []*model.Registry, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "GET", path, b.secret, nil, registries)
	return registries, err
}

func (b *http) RegistryCreate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "POST", path, b.secret, in, nil)
	return err
}

func (b *http) RegistryUpdate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "PUT", path, b.secret, in, nil)
	return err
}

func (b *http) RegistryDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(context.TODO(), "DELETE", path, b.secret, nil, nil)
	return err
}
