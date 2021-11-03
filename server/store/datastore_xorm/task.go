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

func (s storage) TaskList() ([]*model.Task, error) {
	tasks := make([]*model.Task, 0, perPage)
	return tasks, s.engine.Find(&tasks)
}

func (s storage) TaskInsert(task *model.Task) error {
	_, err := s.engine.InsertOne(task)
	return err
}

func (s storage) TaskDelete(id string) error {
	_, err := s.engine.Where("task_id = ?", id).Delete(new(model.Task))
	return err
}
