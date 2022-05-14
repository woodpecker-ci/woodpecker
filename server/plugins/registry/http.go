package registry

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// New returns a new local secret service.
func NewHTTP(endpoint string, privateKey crypto.PrivateKey) model.RegistryService {
	return &http{endpoint, privateKey}
}

func FromRepo(repo *model.Repo) model.RegistryService {
	// TODO: create & use global server key
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return NewHTTP(repo.RegistryEndpoint, privEd25519Key)
}

func (b *http) RegistryFind(ctx context.Context, repo *model.Repo, name string) (registry *model.Registry, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(context.TODO(), "GET", path, b.privateKey, nil, registry)
	return registry, err
}

func (b *http) RegistryList(ctx context.Context, repo *model.Repo) (registries []*model.Registry, err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "GET", path, b.privateKey, nil, registries)
	return registries, err
}

func (b *http) RegistryCreate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "POST", path, b.privateKey, in, nil)
	return err
}

func (b *http) RegistryUpdate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(context.TODO(), "PUT", path, b.privateKey, in, nil)
	return err
}

func (b *http) RegistryDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/secrets/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(context.TODO(), "DELETE", path, b.privateKey, nil, nil)
	return err
}
