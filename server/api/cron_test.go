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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// cronFixture seeds one repo + user into the store. It is created once per
// top-level test and shared across subtests; subtests use unique cron names so
// the per-repo unique constraint never collides.
func cronFixture(t *testing.T, s store.Store) (*model.Repo, *model.User) {
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

// seedCron inserts a cron directly into the store for read/update/delete tests.
func seedCron(t *testing.T, s store.Store, repoID int64, name string) *model.Cron {
	t.Helper()
	cron := &model.Cron{
		RepoID:   repoID,
		Name:     name,
		Schedule: "@every 1h",
		Timezone: "UTC",
		Branch:   "main",
		Enabled:  true,
	}
	require.NoError(t, s.CronCreate(cron))
	return cron
}

// cronForgeManager installs a mock manager returning a fresh mock forge and
// returns the forge so tests can add expectations (e.g. BranchHead).
func cronForgeManager(t *testing.T) *forge_mocks.MockForge {
	t.Helper()
	mgr := manager_mocks.NewMockManager(t)
	forge := forge_mocks.NewMockForge(t)
	mgr.On("ForgeFromRepo", mock.Anything).Return(forge, nil)
	server.Config.Services.Manager = mgr
	return forge
}

func TestGetCron(t *testing.T) {
	s := newTestStore(t)
	repo, _ := cronFixture(t, s)

	t.Run("happy path returns cron", func(t *testing.T) {
		cron := seedCron(t, s, repo.ID, "nightly")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", strItoa(cron.ID))(tc)

		GetCron(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Cron
		tc.decodeJSON(t, &got)
		assert.Equal(t, cron.ID, got.ID)
		assert.Equal(t, "nightly", got.Name)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "not-a-number")(tc)

		GetCron(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing cron returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "9999")(tc)

		GetCron(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})
}

func TestGetCronList(t *testing.T) {
	s := newTestStore(t)
	repo, _ := cronFixture(t, s)

	t.Run("empty repo returns empty list", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)

		GetCronList(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got []*model.Cron
		tc.decodeJSON(t, &got)
		assert.Empty(t, got)
	})

	t.Run("returns all crons for repo", func(t *testing.T) {
		seedCron(t, s, repo.ID, "list-a")
		seedCron(t, s, repo.ID, "list-b")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)

		GetCronList(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got []*model.Cron
		tc.decodeJSON(t, &got)
		assert.Len(t, got, 2)
	})
}

func TestDeleteCron(t *testing.T) {
	s := newTestStore(t)
	repo, _ := cronFixture(t, s)

	t.Run("happy path deletes", func(t *testing.T) {
		cron := seedCron(t, s, repo.ID, "del")
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", strItoa(cron.ID))(tc)

		DeleteCron(tc.Ctx)

		assert.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status())
		_, err := s.CronFind(repo, cron.ID)
		assert.Error(t, err)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "abc")(tc)

		DeleteCron(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing cron returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "9999")(tc)

		DeleteCron(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})
}

func TestPostCron(t *testing.T) {
	s := newTestStore(t)
	repo, user := cronFixture(t, s)

	t.Run("happy path creates cron", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Cron{Name: "build", Schedule: "@every 1h"})(tc)

		PostCron(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Cron
		tc.decodeJSON(t, &got)
		assert.Equal(t, "build", got.Name)
		assert.Equal(t, "UTC", got.Timezone) // defaulted
		assert.Positive(t, got.NextExec)
	})

	t.Run("validation fails on empty name", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Cron{Schedule: "@every 1h"})(tc)

		PostCron(tc.Ctx)

		assert.Equal(t, http.StatusUnprocessableEntity, tc.Recorder.Code)
	})

	t.Run("malformed body returns bad request", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRawBody(http.MethodPost, "application/json", []byte("{not json"))(tc)

		PostCron(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("unparseable schedule fails validation", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Cron{Name: "x", Schedule: "totally-invalid"})(tc)

		PostCron(tc.Ctx)

		assert.Equal(t, http.StatusUnprocessableEntity, tc.Recorder.Code)
	})

	t.Run("branch checked on forge when set", func(t *testing.T) {
		forge := cronForgeManager(t)
		forge.On("BranchHead", mock.Anything, user, mock.Anything, "feature").
			Return(&model.Commit{SHA: "abc"}, nil)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Cron{Name: "br", Schedule: "@every 1h", Branch: "feature"})(tc)

		PostCron(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		forge.AssertCalled(t, "BranchHead", mock.Anything, user, mock.Anything, "feature")
	})

	t.Run("duplicate cron returns conflict", func(t *testing.T) {
		seedCron(t, s, repo.ID, "dup")
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withRequest(http.MethodPost, &model.Cron{Name: "dup", Schedule: "@every 1h"})(tc)

		PostCron(tc.Ctx)

		assert.Equal(t, http.StatusConflict, tc.Recorder.Code)
	})
}

func TestRunCron(t *testing.T) {
	s := newTestStore(t)
	repo, _ := cronFixture(t, s)

	t.Run("invalid id returns bad request", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "abc")(tc)

		RunCron(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing cron returns not found", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "9999")(tc)

		RunCron(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})

	t.Run("pipeline creation failure returns internal error", func(t *testing.T) {
		cron := seedCron(t, s, repo.ID, "run-fail")
		// CreatePipeline resolves repo+user from the store, then asks the forge
		// for the branch head. Make that fail so RunCron hits its 500 branch
		// without exercising the full pipeline.Create machinery.
		forge := cronForgeManager(t)
		forge.On("BranchHead", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		tc := newTestContext(t, s)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", strItoa(cron.ID))(tc)

		RunCron(tc.Ctx)

		assert.Equal(t, http.StatusInternalServerError, tc.Recorder.Code)
	})
}

func TestPatchCron(t *testing.T) {
	s := newTestStore(t)
	repo, user := cronFixture(t, s)

	t.Run("happy path updates name and schedule", func(t *testing.T) {
		cron := seedCron(t, s, repo.ID, "orig")
		cronForgeManager(t)

		newName := "renamed"
		newSchedule := "@every 2h"
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", strItoa(cron.ID))(tc)
		withRequest(http.MethodPatch, &model.CronPatch{Name: &newName, Schedule: &newSchedule})(tc)

		PatchCron(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Cron
		tc.decodeJSON(t, &got)
		assert.Equal(t, "renamed", got.Name)
		assert.Equal(t, "@every 2h", got.Schedule)
	})

	t.Run("invalid id returns bad request", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "nope")(tc)
		withRequest(http.MethodPatch, &model.CronPatch{})(tc)

		PatchCron(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})

	t.Run("missing cron returns not found", func(t *testing.T) {
		cronForgeManager(t)
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", "9999")(tc)
		withRequest(http.MethodPatch, &model.CronPatch{})(tc)

		PatchCron(tc.Ctx)

		assert.Equal(t, http.StatusNotFound, tc.Recorder.Code)
	})

	t.Run("branch checked on forge when patched", func(t *testing.T) {
		cron := seedCron(t, s, repo.ID, "brpatch")
		forge := cronForgeManager(t)
		forge.On("BranchHead", mock.Anything, user, mock.Anything, "develop").
			Return(&model.Commit{SHA: "abc"}, nil)

		branch := "develop"
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{})(tc)
		withParam("cron", strItoa(cron.ID))(tc)
		withRequest(http.MethodPatch, &model.CronPatch{Branch: &branch})(tc)

		PatchCron(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		var got model.Cron
		tc.decodeJSON(t, &got)
		assert.Equal(t, "develop", got.Branch)
	})
}
