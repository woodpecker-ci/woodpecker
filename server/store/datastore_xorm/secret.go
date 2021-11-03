// Copyright 2021 Woodpecker Authors
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

package datastore_xorm

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	secret := &model.Secret{
		RepoID: repo.ID,
		Name:   name,
	}
	return secret, wrapGet(s.engine.Get(secret))
}

func (s storage) SecretList(repo *model.Repo) ([]*model.Secret, error) {
	secrets := make([]*model.Secret, 0, perPage)
	return secrets, s.engine.Where("secret_repo_id = ?", repo.ID).Find(&secrets)
}

func (s storage) SecretCreate(secret *model.Secret) error {
	_, err := s.engine.InsertOne(secret)
	return err
}

func (s storage) SecretUpdate(secret *model.Secret) error {
	_, err := s.engine.ID(secret.ID).Update(&secret)
	return err
}

func (s storage) SecretDelete(secret *model.Secret) error {
	_, err := s.engine.ID(secret.ID).Delete(new(model.Secret))
	return err
}
