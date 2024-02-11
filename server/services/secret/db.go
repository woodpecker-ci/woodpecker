// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type db struct {
	store model.SecretStore
}

// NewDB returns a new local secret service.
func NewDB(store model.SecretStore) Service {
	return &db{store: store}
}

func (d *db) SecretFind(_ context.Context, repo *model.Repo, name string) (*model.Secret, error) {
	return d.store.SecretFind(repo, name)
}

func (d *db) SecretList(_ context.Context, repo *model.Repo, p *model.ListOptions) ([]*model.Secret, error) {
	return d.store.SecretList(repo, false, p)
}

func (d *db) SecretListPipeline(_ context.Context, repo *model.Repo, _ *model.Pipeline, p *model.ListOptions) ([]*model.Secret, error) {
	s, err := d.store.SecretList(repo, true, p)
	if err != nil {
		return nil, err
	}

	// Return only secrets with unique name
	// Priority order in case of duplicate names are repository, user/organization, global
	secrets := make([]*model.Secret, 0, len(s))
	uniq := make(map[string]struct{})
	for _, condition := range []struct {
		IsRepository   bool
		IsOrganization bool
		IsGlobal       bool
	}{
		{IsRepository: true},
		{IsOrganization: true},
		{IsGlobal: true},
	} {
		for _, secret := range s {
			if secret.IsRepository() != condition.IsRepository || secret.IsOrganization() != condition.IsOrganization || secret.IsGlobal() != condition.IsGlobal {
				continue
			}
			if _, ok := uniq[secret.Name]; ok {
				continue
			}
			uniq[secret.Name] = struct{}{}
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}

func (d *db) SecretCreate(_ context.Context, _ *model.Repo, in *model.Secret) error {
	return d.store.SecretCreate(in)
}

func (d *db) SecretUpdate(_ context.Context, _ *model.Repo, in *model.Secret) error {
	return d.store.SecretUpdate(in)
}

func (d *db) SecretDelete(_ context.Context, repo *model.Repo, name string) error {
	secret, err := d.store.SecretFind(repo, name)
	if err != nil {
		return err
	}
	return d.store.SecretDelete(secret)
}

func (d *db) OrgSecretFind(_ context.Context, owner int64, name string) (*model.Secret, error) {
	return d.store.OrgSecretFind(owner, name)
}

func (d *db) OrgSecretList(_ context.Context, owner int64, p *model.ListOptions) ([]*model.Secret, error) {
	return d.store.OrgSecretList(owner, p)
}

func (d *db) OrgSecretCreate(_ context.Context, _ int64, in *model.Secret) error {
	return d.store.SecretCreate(in)
}

func (d *db) OrgSecretUpdate(_ context.Context, _ int64, in *model.Secret) error {
	return d.store.SecretUpdate(in)
}

func (d *db) OrgSecretDelete(_ context.Context, owner int64, name string) error {
	secret, err := d.store.OrgSecretFind(owner, name)
	if err != nil {
		return err
	}
	return d.store.SecretDelete(secret)
}

func (d *db) GlobalSecretFind(_ context.Context, owner string) (*model.Secret, error) {
	return d.store.GlobalSecretFind(owner)
}

func (d *db) GlobalSecretList(_ context.Context, p *model.ListOptions) ([]*model.Secret, error) {
	return d.store.GlobalSecretList(p)
}

func (d *db) GlobalSecretCreate(_ context.Context, in *model.Secret) error {
	return d.store.SecretCreate(in)
}

func (d *db) GlobalSecretUpdate(_ context.Context, in *model.Secret) error {
	return d.store.SecretUpdate(in)
}

func (d *db) GlobalSecretDelete(_ context.Context, name string) error {
	secret, err := d.store.GlobalSecretFind(name)
	if err != nil {
		return err
	}
	return d.store.SecretDelete(secret)
}
