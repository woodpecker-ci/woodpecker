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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestTokenCreate(t *testing.T) {
	store, closer := newTestStore(t, new(model.Token))
	defer closer()

	repo := &model.Repo{ID: 1, Name: "repo"}
	token := model.NewToken(repo.ID, model.TokenTypeBadge)
	assert.NoError(t, store.TokenCreate(token))
	assert.NotEqualValues(t, 0, token.ID)

	// a repo can only hold one token per type
	duplicate := model.NewToken(repo.ID, model.TokenTypeBadge)
	assert.ErrorIs(t, store.TokenCreate(duplicate), types.ErrInsertDuplicateDetected)

	// invalid tokens are rejected
	assert.Error(t, store.TokenCreate(&model.Token{RepoID: repo.ID, Type: model.TokenType("unknown"), Value: "x"}))
}

func TestTokenFind(t *testing.T) {
	store, closer := newTestStore(t, new(model.Token))
	defer closer()

	repo1 := &model.Repo{ID: 1, Name: "repo1"}
	repo2 := &model.Repo{ID: 2, Name: "repo2"}
	token := model.NewToken(repo1.ID, model.TokenTypeBadge)
	assert.NoError(t, store.TokenCreate(token))

	found, err := store.TokenFind(repo1, model.TokenTypeBadge)
	assert.NoError(t, err)
	assert.Equal(t, token.Value, found.Value)

	// tokens are not shared between repos
	_, err = store.TokenFind(repo2, model.TokenTypeBadge)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)
}

func TestTokenDeletedWithRepo(t *testing.T) {
	store, closer := newTestStore(t,
		new(model.Token),
		new(model.Repo),
		new(model.Perm),
		new(model.Pipeline),
		new(model.PipelineConfig),
		new(model.LogEntry),
		new(model.Step),
		new(model.Secret),
		new(model.Registry),
		new(model.Config),
		new(model.Redirection),
		new(model.Workflow))
	defer closer()

	repo := &model.Repo{Name: "repo", Owner: "owner", FullName: "owner/repo", ForgeRemoteID: "1"}
	assert.NoError(t, store.CreateRepo(repo))
	assert.NoError(t, store.TokenCreate(model.NewToken(repo.ID, model.TokenTypeBadge)))

	assert.NoError(t, store.DeleteRepo(repo))

	_, err := store.TokenFind(repo, model.TokenTypeBadge)
	assert.ErrorIs(t, err, types.ErrRecordNotExist)
}
