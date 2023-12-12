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

package secrets

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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

func (b *builtin) SecretList(repo *model.Repo, p *model.ListOptions) ([]*model.Secret, error) {
	return b.store.SecretList(repo, false, p)
}

func (b *builtin) SecretListPipeline(repo *model.Repo, _ *model.Pipeline, p *model.ListOptions) ([]*model.Secret, error) {
	s, err := b.store.SecretList(repo, true, p)
	if err != nil {
		return nil, err
	}

	// Return only secrets with unique name
	// Priority order in case of duplicate names are repository, user/organization, global
	secrets := make([]*model.Secret, 0, len(s))
	uniq := make(map[string]struct{})
	for _, cond := range []struct {
		Global       bool
		Organization bool
	}{
		{},
		{Organization: true},
		{Global: true},
	} {
		for _, secret := range s {
			if secret.Global() != cond.Global || secret.Organization() != cond.Organization {
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

func (b *builtin) SecretCreate(_ *model.Repo, in *model.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) SecretUpdate(_ *model.Repo, in *model.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) SecretDelete(repo *model.Repo, name string) error {
	secret, err := b.store.SecretFind(repo, name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}

func (b *builtin) OrgSecretFind(owner int64, name string) (*model.Secret, error) {
	return b.store.OrgSecretFind(owner, name)
}

func (b *builtin) OrgSecretList(owner int64, p *model.ListOptions) ([]*model.Secret, error) {
	return b.store.OrgSecretList(owner, p)
}

func (b *builtin) OrgSecretCreate(_ int64, in *model.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) OrgSecretUpdate(_ int64, in *model.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) OrgSecretDelete(owner int64, name string) error {
	secret, err := b.store.OrgSecretFind(owner, name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}

func (b *builtin) GlobalSecretFind(owner string) (*model.Secret, error) {
	return b.store.GlobalSecretFind(owner)
}

func (b *builtin) GlobalSecretList(p *model.ListOptions) ([]*model.Secret, error) {
	return b.store.GlobalSecretList(p)
}

func (b *builtin) GlobalSecretCreate(in *model.Secret) error {
	return b.store.SecretCreate(in)
}

func (b *builtin) GlobalSecretUpdate(in *model.Secret) error {
	return b.store.SecretUpdate(in)
}

func (b *builtin) GlobalSecretDelete(name string) error {
	secret, err := b.store.GlobalSecretFind(name)
	if err != nil {
		return err
	}
	return b.store.SecretDelete(secret)
}
