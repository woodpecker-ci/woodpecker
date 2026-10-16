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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	log_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/log/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestPostRepoReturnsConflictOnDuplicateRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockStore := store_mocks.NewMockStore(t)
	mockManager := manager_mocks.NewMockManager(t)
	mockForge := forge_mocks.NewMockForge(t)

	server.Config.Services.Manager = mockManager
	server.Config.Permissions.OwnersAllowlist = permissions.NewOwnersAllowlist(nil)
	server.Config.Server.WebhookHost = "https://woodpecker.example"
	server.Config.Pipeline.DefaultApprovalMode = model.RequireApprovalForks
	server.Config.Pipeline.DefaultAllowPullRequests = true
	server.Config.Pipeline.DefaultCancelPreviousPipelineEvents = nil
	server.Config.Pipeline.DefaultTimeout = 60
	server.Config.Pipeline.MaxTimeout = 120

	user := &model.User{ID: 10, ForgeID: 7, Login: "alice"}
	forgeRemoteID := model.ForgeRemoteID("42")

	forgeRepo := &model.Repo{
		ForgeRemoteID: forgeRemoteID,
		Owner:         "acme",
		Name:          "rocket",
		FullName:      "acme/rocket",
		Perm:          &model.Perm{Admin: true},
	}

	org := &model.Org{ID: 3, Name: "acme", ForgeID: user.ForgeID}

	mockManager.On("ForgeFromUser", user).Return(mockForge, nil)
	mockStore.On("GetRepoForgeID", user.ForgeID, forgeRemoteID).Return(nil, types.ErrRecordNotExist)
	mockForge.On("Repo", mock.Anything, user, forgeRemoteID, "", "").Return(forgeRepo, nil)
	mockStore.On("OrgFindByName", forgeRepo.Owner, user.ForgeID).Return(org, nil)
	mockForge.On("Activate", mock.Anything, user, mock.AnythingOfType("*model.Repo"), mock.AnythingOfType("string")).Return(nil)
	mockStore.On("CreateRepo", mock.AnythingOfType("*model.Repo")).Return(types.ErrInsertDuplicateDetected)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("store", mockStore)
	c.Set("user", user)
	c.Request = httptest.NewRequest(http.MethodPost, "/repos?forge_remote_id=42", nil)

	PostRepo(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Remove the stale repository entry")
	mockStore.AssertNotCalled(t, "PermUpsert", mock.Anything)
}

func moveRepoForge(t *testing.T) *forge_mocks.MockForge {
	t.Helper()
	mgr := manager_mocks.NewMockManager(t)
	mockForge := forge_mocks.NewMockForge(t)
	mgr.On("ForgeFromRepo", mock.Anything).Return(mockForge, nil)
	server.Config.Services.Manager = mgr
	server.Config.Server.WebhookHost = "https://woodpecker.example"
	return mockForge
}

func TestMoveRepoUpdatesOrg(t *testing.T) {
	s := newTestStore(t)

	seed := func(t *testing.T, forgeID int64, owner, name string) (*model.User, *model.Org, *model.Repo) {
		t.Helper()
		user := &model.User{Login: "alice-" + name, ForgeID: forgeID, ForgeRemoteID: model.ForgeRemoteID("u-" + name), Hash: "userhash-" + name}
		require.NoError(t, s.CreateUser(user))
		oldOrg := &model.Org{Name: owner, ForgeID: forgeID}
		require.NoError(t, s.OrgCreate(oldOrg))
		repo := &model.Repo{
			ForgeID:       forgeID,
			ForgeRemoteID: model.ForgeRemoteID("r-" + name),
			Owner:         owner,
			Name:          name,
			FullName:      owner + "/" + name,
			OrgID:         oldOrg.ID,
			UserID:        user.ID,
			Hash:          "hash-" + name,
			IsActive:      true,
		}
		require.NoError(t, s.CreateRepo(repo))
		return user, oldOrg, repo
	}

	move := func(t *testing.T, user *model.User, repo *model.Repo, to string) (int, string) {
		t.Helper()
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{Admin: true})(tc)
		tc.Ctx.Request = httptest.NewRequest(http.MethodPost, "/repos/1/move?to="+to, nil)
		MoveRepo(tc.Ctx)
		return tc.Ctx.Writer.Status(), tc.Recorder.Body.String()
	}

	t.Run("move to unknown owner creates the org and relinks the repo", func(t *testing.T) {
		user, oldOrg, repo := seed(t, 1, "oldcorp", "rocket")
		mockForge := moveRepoForge(t)

		from := &model.Repo{
			ForgeRemoteID: repo.ForgeRemoteID,
			Owner:         "newcorp",
			Name:          "rocket",
			FullName:      "newcorp/rocket",
			Perm:          &model.Perm{Admin: true},
		}
		mockForge.On("Repo", mock.Anything, user, model.ForgeRemoteID(""), "newcorp", "rocket").Return(from, nil)
		mockForge.On("Org", mock.Anything, user, "newcorp").Return(&model.Org{Name: "newcorp"}, nil)
		mockForge.On("Deactivate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)
		mockForge.On("Activate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)

		code, body := move(t, user, repo, "newcorp/rocket")
		require.Equal(t, http.StatusNoContent, code, body)

		stored, err := s.GetRepo(repo.ID)
		require.NoError(t, err)
		assert.Equal(t, "newcorp/rocket", stored.FullName)
		assert.NotEqual(t, oldOrg.ID, stored.OrgID, "repo must not stay linked to the old org")

		newOrg, err := s.OrgFindByName("newcorp", repo.ForgeID)
		require.NoError(t, err)
		assert.Equal(t, newOrg.ID, stored.OrgID)
		assert.EqualValues(t, repo.ForgeID, newOrg.ForgeID)
	})

	t.Run("move to known owner links the existing org of the repo's forge", func(t *testing.T) {
		user, oldOrg, repo := seed(t, 1, "oldinc", "probe")
		// same-named org on ANOTHER forge that must not be picked
		require.NoError(t, s.OrgCreate(&model.Org{Name: "newinc", ForgeID: 2}))
		target := &model.Org{Name: "newinc", ForgeID: 1}
		require.NoError(t, s.OrgCreate(target))

		mockForge := moveRepoForge(t)
		from := &model.Repo{
			ForgeRemoteID: repo.ForgeRemoteID,
			Owner:         "newinc",
			Name:          "probe",
			FullName:      "newinc/probe",
			Perm:          &model.Perm{Admin: true},
		}
		// no Org() expectation: forge must not be asked for a known org
		mockForge.On("Repo", mock.Anything, user, model.ForgeRemoteID(""), "newinc", "probe").Return(from, nil)
		mockForge.On("Deactivate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)
		mockForge.On("Activate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)

		code, body := move(t, user, repo, "newinc/probe")
		require.Equal(t, http.StatusNoContent, code, body)

		stored, err := s.GetRepo(repo.ID)
		require.NoError(t, err)
		assert.Equal(t, target.ID, stored.OrgID)
		assert.NotEqual(t, oldOrg.ID, stored.OrgID)
	})
}

func TestDeleteRepoDeletesStoredLogs(t *testing.T) {
	s := newTestStore(t)

	seed := func(t *testing.T, name string) (*model.User, *model.Repo, []*model.Step) {
		t.Helper()
		user := &model.User{Login: "bob-" + name, ForgeID: 1, ForgeRemoteID: model.ForgeRemoteID("u-" + name), Hash: "userhash-" + name}
		require.NoError(t, s.CreateUser(user))
		repo := &model.Repo{
			ForgeID:       1,
			ForgeRemoteID: model.ForgeRemoteID("r-" + name),
			Owner:         "acme",
			Name:          name,
			FullName:      "acme/" + name,
			UserID:        user.ID,
			Hash:          "hash-" + name,
			IsActive:      true,
		}
		require.NoError(t, s.CreateRepo(repo))

		var steps []*model.Step
		for i := range 2 {
			pipelineSteps := []*model.Step{{PID: 1, Name: "build"}, {PID: 2, Name: "test"}}
			require.NoError(t, s.CreatePipeline(&model.Pipeline{RepoID: repo.ID, Number: int64(i + 1), Status: model.StatusSuccess}, pipelineSteps...))
			steps = append(steps, pipelineSteps...)
		}
		return user, repo, steps
	}

	deleteRepo := func(t *testing.T, user *model.User, repo *model.Repo) int {
		t.Helper()
		mockForge := moveRepoForge(t)
		mockForge.On("Deactivate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)

		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{Admin: true})(tc)
		tc.Ctx.Request = httptest.NewRequest(http.MethodDelete, "/repos/1?remove=true", nil)
		DeleteRepo(tc.Ctx)
		return tc.Ctx.Writer.Status()
	}

	setLogStore := func(t *testing.T) *log_mocks.MockService {
		t.Helper()
		mockLogStore := log_mocks.NewMockService(t)
		originalLogStore := server.Config.Services.LogStore
		server.Config.Services.LogStore = mockLogStore
		t.Cleanup(func() {
			server.Config.Services.LogStore = originalLogStore
		})
		return mockLogStore
	}

	t.Run("removing a repo deletes the logs of every step", func(t *testing.T) {
		user, repo, steps := seed(t, "rocket")
		mockLogStore := setLogStore(t)
		for _, step := range steps {
			mockLogStore.On("LogDelete", mock.MatchedBy(func(s *model.Step) bool { return s.ID == step.ID })).Return(nil).Once()
		}

		require.Equal(t, http.StatusOK, deleteRepo(t, user, repo))

		_, err := s.GetRepo(repo.ID)
		assert.ErrorIs(t, err, types.ErrRecordNotExist)
	})

	t.Run("a failed log delete keeps the repo", func(t *testing.T) {
		user, repo, _ := seed(t, "probe")
		mockLogStore := setLogStore(t)
		mockLogStore.On("LogDelete", mock.Anything).Return(assert.AnError)

		assert.Equal(t, http.StatusInternalServerError, deleteRepo(t, user, repo))

		_, err := s.GetRepo(repo.ID)
		assert.NoError(t, err)
	})
}
