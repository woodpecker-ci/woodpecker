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

func (s storage) SenderFind(repo *model.Repo, login string) (*model.Sender, error) {
	sender := &model.Sender{
		RepoID: repo.ID,
		Login:  login,
	}
	return sender, wrapGet(s.engine.Get(sender))
}

func (s storage) SenderList(repo *model.Repo) ([]*model.Sender, error) {
	senders := make([]*model.Sender, 0, perPage)
	return senders, s.engine.Where("sender_repo_id = ?", repo.ID).Find(&senders)
}

func (s storage) SenderCreate(sender *model.Sender) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(sender)
	return err
}

func (s storage) SenderUpdate(sender *model.Sender) error {
	_, err := s.engine.ID(sender.ID).Update(sender)
	return err
}

func (s storage) SenderDelete(sender *model.Sender) error {
	_, err := s.engine.ID(sender.ID).Delete(new(model.Sender))
	return err
}
