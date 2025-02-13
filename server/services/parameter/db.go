// Copyright 2024 Woodpecker Authors
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

package parameter

import (
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

type db struct {
	store store.Store
}

// NewDB returns a new parameter service.
func NewDB(store store.Store) Service {
	return &db{store: store}
}

func (d *db) ParameterFind(repo *model.Repo, name string) (*model.Parameter, error) {
	return d.store.ParameterFind(repo, name)
}

func (d *db) ParameterFindByID(repo *model.Repo, id int64) (*model.Parameter, error) {
	return d.store.ParameterFindByID(repo, id)
}

func (d *db) ParameterFindByNameAndBranch(repo *model.Repo, name string, branch string) (*model.Parameter, error) {
	return d.store.ParameterFindByNameAndBranch(repo, name, branch)
}

func (d *db) ParameterList(repo *model.Repo) ([]*model.Parameter, error) {
	return d.store.ParameterList(repo)
}

func (d *db) ParameterCreate(repo *model.Repo, parameter *model.Parameter) error {
	return d.store.ParameterCreate(repo, parameter)
}

func (d *db) ParameterUpdate(repo *model.Repo, parameter *model.Parameter) error {
	return d.store.ParameterUpdate(repo, parameter)
}

func (d *db) ParameterDelete(repo *model.Repo, name string) error {
	return d.store.ParameterDelete(repo, name)
}

func (d *db) ParameterDeleteByID(repo *model.Repo, id int64) error {
	return d.store.ParameterDeleteByID(repo, id)
}
