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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// parameterFixture seeds one repo + user into the store.
func parameterFixture(t *testing.T, s store.Store) (*model.Repo, *model.User) {
	t.Helper()
	user := &model.User{Login: "owner", ForgeRemoteID: "u1"}
	require.NoError(t, s.CreateUser(user))
	repo := &model.Repo{
		UserID:        user.ID,
		ForgeRemoteID: "r1",
		Owner:         "owner",
		Name:          "repo",
		FullName:      "owner/repo",
		IsActive:      true,
	}
	require.NoError(t, s.CreateRepo(repo))
	return repo, user
}

// seedParameter inserts a parameter directly into the store for read/update/delete tests.
func seedParameter(t *testing.T, s store.Store, repoID int64, name string) *model.Parameter {
	t.Helper()
	parameter := &model.Parameter{
		RepoID:  repoID,
		Name:    name,
		Type:    model.ParameterTypeString,
		Default: "foo",
		Source:  model.ParameterSourceRepoConfig,
	}
	require.NoError(t, s.ParameterCreate(parameter))
	return parameter
}

func TestGetParameter(t *testing.T) {
	s := newTestStore(t)
	repo, _ := parameterFixture(t, s)

	t.Run("happy path returns parameter", func(t *testing.T) {
		parameter := seedParameter(t, s, repo.ID, "SOME_VAR")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", strItoa(parameter.ID))(tc)

		GetParameter(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Parameter
		tc.decodeJSON(t, &got)
		assert.Equal(t, parameter.ID, got.ID)
		assert.Equal(t, "SOME_VAR", got.Name)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "not-a-number")(tc)

		GetParameter(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing parameter returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "9999")(tc)

		GetParameter(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})
}

func TestGetParameterList(t *testing.T) {
	s := newTestStore(t)
	repo, _ := parameterFixture(t, s)

	t.Run("empty repo returns empty list", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)

		GetParameterList(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got []*model.Parameter
		tc.decodeJSON(t, &got)
		assert.Empty(t, got)
	})

	t.Run("returns all parameters for repo", func(t *testing.T) {
		seedParameter(t, s, repo.ID, "LIST_A")
		seedParameter(t, s, repo.ID, "LIST_B")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)

		GetParameterList(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got []*model.Parameter
		tc.decodeJSON(t, &got)
		assert.Len(t, got, 2)
	})
}

func TestPostParameter(t *testing.T) {
	s := newTestStore(t)
	repo, user := parameterFixture(t, s)

	t.Run("happy path creates parameter", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Parameter{
			Name:    "DEPLOY_TARGET",
			Type:    model.ParameterTypeChoice,
			Options: []string{"staging", "production"},
			Default: "staging",
		})(tc)

		PostParameter(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Parameter
		tc.decodeJSON(t, &got)
		assert.Equal(t, "DEPLOY_TARGET", got.Name)
		assert.Equal(t, repo.ID, got.RepoID)
		assert.Equal(t, model.ParameterSourceRepoConfig, got.Source) // enforced server-side
	})

	t.Run("validation fails on invalid name", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Parameter{Name: "not a name", Type: model.ParameterTypeString})(tc)

		PostParameter(tc.Ctx)

		assert.Equal(t, http.StatusUnprocessableEntity, tc.Recorder.Code)
	})

	t.Run("validation fails on default not in options", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Parameter{
			Name:    "SOME_CHOICE",
			Type:    model.ParameterTypeChoice,
			Options: []string{"a", "b"},
			Default: "c",
		})(tc)

		PostParameter(tc.Ctx)

		assert.Equal(t, http.StatusUnprocessableEntity, tc.Recorder.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRawBody(http.MethodPost, "application/json", []byte("{not json"))(tc)

		PostParameter(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("duplicate parameter returns conflict", func(t *testing.T) {
		seedParameter(t, s, repo.ID, "DUP_VAR")
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Parameter{Name: "DUP_VAR", Type: model.ParameterTypeString})(tc)

		PostParameter(tc.Ctx)

		assert.Equal(t, http.StatusConflict, tc.Recorder.Code)
	})
}

func TestPatchParameter(t *testing.T) {
	s := newTestStore(t)
	repo, user := parameterFixture(t, s)

	t.Run("happy path updates fields", func(t *testing.T) {
		parameter := seedParameter(t, s, repo.ID, "PATCH_ME")
		newDefault := "bar"
		required := true
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", strItoa(parameter.ID))(tc)
		withRequest(http.MethodPatch, &model.ParameterPatch{Default: &newDefault, Required: &required})(tc)

		PatchParameter(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Parameter
		tc.decodeJSON(t, &got)
		assert.Equal(t, "bar", got.Default)
		assert.True(t, got.Required)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "nope")(tc)
		withRequest(http.MethodPatch, &model.ParameterPatch{})(tc)

		PatchParameter(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing parameter returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "9999")(tc)
		withRequest(http.MethodPatch, &model.ParameterPatch{})(tc)

		PatchParameter(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})

	t.Run("patch resulting in invalid definition fails validation", func(t *testing.T) {
		parameter := seedParameter(t, s, repo.ID, "PATCH_INVALID")
		invalidType := model.ParameterType("nope")
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", strItoa(parameter.ID))(tc)
		withRequest(http.MethodPatch, &model.ParameterPatch{Type: &invalidType})(tc)

		PatchParameter(tc.Ctx)

		assert.Equal(t, http.StatusUnprocessableEntity, tc.Recorder.Code)
	})
}

func TestDeleteParameter(t *testing.T) {
	s := newTestStore(t)
	repo, _ := parameterFixture(t, s)

	t.Run("happy path deletes", func(t *testing.T) {
		parameter := seedParameter(t, s, repo.ID, "DEL_VAR")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", strItoa(parameter.ID))(tc)

		DeleteParameter(tc.Ctx)

		assert.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status())
		_, err := s.ParameterFind(repo, parameter.ID)
		assert.Error(t, err)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "abc")(tc)

		DeleteParameter(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing parameter returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("parameter", "9999")(tc)

		DeleteParameter(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})
}
