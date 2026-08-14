// Copyright 2026 Woodpecker Authors
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
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func (s storage) ParameterCreate(parameter *model.Parameter) error {
	if err := parameter.Validate(); err != nil {
		return err
	}
	err := wrapInsert(s.engine.Insert(parameter))
	if errors.Is(err, types.ErrInsertDuplicateDetected) {
		return fmt.Errorf("create parameter failed, duplicate detected: %w", err)
	}
	return err
}

func (s storage) ParameterFind(repo *model.Repo, id int64) (*model.Parameter, error) {
	parameter := new(model.Parameter)
	return parameter, wrapGet(s.engine.ID(id).Where("repo_id = ?", repo.ID).Get(parameter))
}

func (s storage) ParameterList(repo *model.Repo, p *model.ListOptions) ([]*model.Parameter, error) {
	var parameters []*model.Parameter
	return parameters, s.paginate(p).Where("repo_id = ?", repo.ID).OrderBy("display_order").OrderBy("name").Find(&parameters)
}

func (s storage) ParameterUpdate(parameter *model.Parameter) error {
	if err := parameter.Validate(); err != nil {
		return err
	}
	_, err := s.engine.ID(parameter.ID).AllCols().Update(parameter)
	return err
}

func (s storage) ParameterDelete(repo *model.Repo, id int64) error {
	return wrapDelete(s.engine.ID(id).Where("repo_id = ?", repo.ID).Delete(new(model.Parameter)))
}
