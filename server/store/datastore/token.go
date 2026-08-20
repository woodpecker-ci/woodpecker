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

func (s storage) TokenCreate(token *model.Token) error {
	if err := token.Validate(); err != nil {
		return err
	}
	err := wrapInsert(s.engine.Insert(token))
	if errors.Is(err, types.ErrInsertDuplicateDetected) {
		return fmt.Errorf("create token failed, duplicate detected: %w", err)
	}
	return err
}

func (s storage) TokenFind(repo *model.Repo, tokenType model.TokenType) (*model.Token, error) {
	token := new(model.Token)
	return token, wrapGet(s.engine.Where("repo_id = ? AND type = ?", repo.ID, tokenType).Get(token))
}
