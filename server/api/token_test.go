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

//go:build test

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

func TestGetRepoToken(t *testing.T) {
	s := newTestStore(t)

	newRepo := func(t *testing.T, name string) *model.Repo {
		t.Helper()
		repo := &model.Repo{
			ForgeRemoteID: model.ForgeRemoteID(name),
			Owner:         "owner",
			Name:          name,
			FullName:      "owner/" + name,
			IsActive:      true,
		}
		require.NoError(t, s.CreateRepo(repo))
		return repo
	}

	call := func(t *testing.T, s store.Store, repo *model.Repo, query string) *httptest.ResponseRecorder {
		t.Helper()
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{Push: true})(tc)
		tc.Ctx.Request = httptest.NewRequest(http.MethodGet, "/?"+query, nil)
		GetRepoToken(tc.Ctx)
		return tc.Recorder
	}

	t.Run("returns the badge token by default", func(t *testing.T) {
		repo := newRepo(t, "with-token")
		token := model.NewToken(repo.ID, model.TokenTypeBadge)
		require.NoError(t, s.TokenCreate(token))

		rec := call(t, s, repo, "")
		assert.Equal(t, http.StatusOK, rec.Code)

		var got model.Token
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, token.Value, got.Value)
		assert.Equal(t, model.TokenTypeBadge, got.Type)
		assert.Equal(t, repo.ID, got.RepoID)
	})

	t.Run("rejects unknown token types", func(t *testing.T) {
		repo := newRepo(t, "bad-type")
		rec := call(t, s, repo, "type=cron")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	// repos activated before badge tokens existed hold none, the endpoint
	// equips them on first read instead of failing
	t.Run("creates a missing token", func(t *testing.T) {
		repo := newRepo(t, "without-token")

		rec := call(t, s, repo, "")
		assert.Equal(t, http.StatusOK, rec.Code)

		var got model.Token
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.NotEmpty(t, got.Value)

		stored, err := s.TokenFind(repo, model.TokenTypeBadge)
		require.NoError(t, err)
		assert.Equal(t, stored.Value, got.Value)

		// reading again must not hand out a different token
		second := call(t, s, repo, "")
		var again model.Token
		require.NoError(t, json.Unmarshal(second.Body.Bytes(), &again))
		assert.Equal(t, got.Value, again.Value)
	})
}

func TestRotateRepoTokens(t *testing.T) {
	s := newTestStore(t)

	repo := &model.Repo{
		ForgeRemoteID: "rotate",
		Owner:         "owner",
		Name:          "rotate",
		FullName:      "owner/rotate",
		IsActive:      true,
	}
	require.NoError(t, s.CreateRepo(repo))

	old := model.NewToken(repo.ID, model.TokenTypeBadge)
	require.NoError(t, s.TokenCreate(old))

	rotated := rotateTokens(t, s, repo)

	require.Len(t, rotated, 1)
	assert.Equal(t, model.TokenTypeBadge, rotated[0].Type)
	assert.NotEqual(t, old.Value, rotated[0].Value, "the old token has to stop working")

	stored, err := s.TokenFind(repo, model.TokenTypeBadge)
	require.NoError(t, err)
	assert.Equal(t, rotated[0].Value, stored.Value)
}

// Rotating a repo that never had a token must equip it rather than fail.
func TestRotateRepoTokensWithoutExisting(t *testing.T) {
	s := newTestStore(t)

	repo := &model.Repo{
		ForgeRemoteID: "rotate-fresh",
		Owner:         "owner",
		Name:          "rotate-fresh",
		FullName:      "owner/rotate-fresh",
		IsActive:      true,
	}
	require.NoError(t, s.CreateRepo(repo))

	rotated := rotateTokens(t, s, repo)

	require.Len(t, rotated, 1)
	_, err := s.TokenFind(repo, model.TokenTypeBadge)
	assert.NoError(t, err)
}

func rotateTokens(t *testing.T, s store.Store, repo *model.Repo) []*model.Token {
	t.Helper()

	tc := newTestContext(t, s)
	withRepo(repo, nil)(tc)
	tc.Ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	RotateRepoTokens(tc.Ctx)

	require.Equal(t, http.StatusOK, tc.Recorder.Code)

	var rotated []*model.Token
	require.NoError(t, json.Unmarshal(tc.Recorder.Body.Bytes(), &rotated))
	return rotated
}
