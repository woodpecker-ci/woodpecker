package registry

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/extensions/utils"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type http struct {
	endpoint   string
	privateKey crypto.PrivateKey
}

// New returns a new local secret service.
func NewHTTP(endpoint string, privateKey crypto.PrivateKey) RegistryExtension {
	return &http{endpoint, privateKey}
}

func FromRepo(repo *model.Repo) RegistryExtension {
	if repo.RegistryEndpoint == "" {
		return nil
	}

	// TODO: create & use global server key
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return NewHTTP(repo.RegistryEndpoint, privEd25519Key)
}

func (b *http) RegistryFind(ctx context.Context, repo *model.Repo, name string) (registry *model.Registry, err error) {
	path := fmt.Sprintf("%s/registries/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "GET", path, b.privateKey, nil, registry)
	return registry, err
}

func (b *http) RegistryList(ctx context.Context, repo *model.Repo) (registries []*model.Registry, err error) {
	path := fmt.Sprintf("%s/registries/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "GET", path, b.privateKey, nil, registries)
	return registries, err
}

func (b *http) RegistryCreate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/registries/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "POST", path, b.privateKey, in, nil)
	return err
}

func (b *http) RegistryUpdate(ctx context.Context, repo *model.Repo, in *model.Registry) (err error) {
	path := fmt.Sprintf("%s/registries/%s/%s", b.endpoint, repo.Owner, repo.Name)
	_, err = utils.Send(ctx, "PUT", path, b.privateKey, in, nil)
	return err
}

func (b *http) RegistryDelete(ctx context.Context, repo *model.Repo, name string) (err error) {
	path := fmt.Sprintf("%s/registries/%s/%s/%s", b.endpoint, repo.Owner, repo.Name, name)
	_, err = utils.Send(ctx, "DELETE", path, b.privateKey, nil, nil)
	return err
}
