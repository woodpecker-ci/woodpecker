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

package datastore

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) ForgeGet(id int64) (*model.Forge, error) {
	forge := new(model.Forge)
	return forge, wrapGet(s.engine.ID(id).Get(forge))
}

func (s storage) ForgeFind(repo *model.Repo) (*model.Forge, error) {
	forge := new(model.Forge)
	return forge, wrapGet(s.engine.Where("forge_id=?", repo.ForgeID).Get(forge))
}

func (s storage) ForgeList() ([]*model.Forge, error) {
	forges := make([]*model.Forge, 0, 10)
	return forges, s.engine.Find(&forges)
}

func (s storage) ForgeCreate(forge *model.Forge) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(forge)
	return err
}

func (s storage) ForgeUpdate(forge *model.Forge) error {
	_, err := s.engine.ID(forge.ID).AllCols().Update(forge)
	return err
}

func (s storage) ForgeDelete(forge *model.Forge) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.ID(forge.ID).Delete(new(model.Forge)); err != nil {
		return err
	}

	return sess.Commit()
}
