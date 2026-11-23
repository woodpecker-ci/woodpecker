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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	log_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/log/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
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

	org := &model.Org{ID: 3, Name: "acme", ForgeID: user.ForgeID, ForgeRemoteID: "5"}

	mockManager.On("ForgeFromUser", user).Return(mockForge, nil)
	mockStore.On("GetRepoForgeID", user.ForgeID, forgeRemoteID).Return(nil, types.ErrRecordNotExist)
	mockForge.On("Repo", mock.Anything, user, forgeRemoteID, "", "").Return(forgeRepo, nil)
	mockForge.On("Org", mock.Anything, user, forgeRepo.Owner).Return(&model.Org{Name: "acme", ForgeRemoteID: "5"}, nil)
	mockStore.On("OrgFindByRemoteID", user.ForgeID, org.ForgeRemoteID).Return(org, nil)
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

func TestPostRepoReturnsConflictOnOrgNameOfGoneForgeAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockStore := store_mocks.NewMockStore(t)
	mockManager := manager_mocks.NewMockManager(t)
	mockForge := forge_mocks.NewMockForge(t)

	server.Config.Services.Manager = mockManager
	server.Config.Permissions.OwnersAllowlist = permissions.NewOwnersAllowlist(nil)
	server.Config.Pipeline.DefaultTimeout = 60
	server.Config.Pipeline.MaxTimeout = 120

	user := &model.User{ID: 10, ForgeID: 7, Login: "alice"}
	forgeRepo := &model.Repo{ForgeRemoteID: "42", Owner: "acme", Name: "rocket", FullName: "acme/rocket", Perm: &model.Perm{Admin: true}}
	// the stored org named acme belongs to a forge account that is gone
	stale := &model.Org{ID: 3, Name: "acme", ForgeID: user.ForgeID, ForgeRemoteID: "5"}

	mockManager.On("ForgeFromUser", user).Return(mockForge, nil)
	mockStore.On("GetRepoForgeID", user.ForgeID, forgeRepo.ForgeRemoteID).Return(nil, types.ErrRecordNotExist)
	mockForge.On("Repo", mock.Anything, user, forgeRepo.ForgeRemoteID, "", "").Return(forgeRepo, nil)
	mockForge.On("Org", mock.Anything, user, "acme").Return(&model.Org{Name: "acme", ForgeRemoteID: "6"}, nil)
	mockStore.On("OrgFindByRemoteID", user.ForgeID, model.ForgeRemoteID("6")).Return(nil, types.ErrRecordNotExist)
	mockStore.On("OrgFindByName", "acme", user.ForgeID).Return(stale, nil)
	mockForge.On("OrgByRemoteID", mock.Anything, user, stale.ForgeRemoteID, false).Return(nil, forge_types.ErrOrgNotFound)
	mockStore.On("OrgUpdate", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("store", mockStore)
	c.Set("user", user)
	c.Request = httptest.NewRequest(http.MethodPost, "/repos?forge_remote_id=42", nil)

	PostRepo(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "an admin has to resolve the conflict")
	assert.EqualValues(t, "6", stale.ConflictForgeRemoteID)
	mockForge.AssertNotCalled(t, "Activate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStore.AssertNotCalled(t, "CreateRepo", mock.Anything)
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
			WebhookSecret: "hash-" + name,
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
		mockForge.EXPECT().Activate(mock.Anything, user, mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, _ *model.User, activated *model.Repo, _ string) error {
				activated.CloneSecret = &model.CloneSecret{ForgeRemoteID: "moved"}
				return nil
			})

		code, body := move(t, user, repo, "newcorp/rocket")
		require.Equal(t, http.StatusNoContent, code, body)

		stored, err := s.GetRepo(repo.ID)
		require.NoError(t, err)
		assert.Equal(t, "newcorp/rocket", stored.FullName)
		assert.Equal(t, &model.CloneSecret{ForgeRemoteID: "moved"}, stored.CloneSecret, "persist the clone secret written by activation")
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
		target := &model.Org{Name: "newinc", ForgeID: 1, ForgeRemoteID: "o-newinc"}
		require.NoError(t, s.OrgCreate(target))

		mockForge := moveRepoForge(t)
		from := &model.Repo{
			ForgeRemoteID: repo.ForgeRemoteID,
			Owner:         "newinc",
			Name:          "probe",
			FullName:      "newinc/probe",
			Perm:          &model.Perm{Admin: true},
		}
		mockForge.On("Repo", mock.Anything, user, model.ForgeRemoteID(""), "newinc", "probe").Return(from, nil)
		mockForge.On("Org", mock.Anything, user, "newinc").Return(&model.Org{Name: "newinc", ForgeRemoteID: "o-newinc"}, nil)
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
			WebhookSecret: "hash-" + name,
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
			mockLogStore.On("LogDelete", mock.Anything, mock.MatchedBy(func(s *model.Step) bool { return s.ID == step.ID })).Return(nil).Once()
		}

		require.Equal(t, http.StatusOK, deleteRepo(t, user, repo))

		_, err := s.GetRepo(repo.ID)
		assert.ErrorIs(t, err, types.ErrRecordNotExist)
	})

	t.Run("a failed log delete keeps the repo", func(t *testing.T) {
		user, repo, _ := seed(t, "probe")
		mockLogStore := setLogStore(t)
		mockLogStore.On("LogDelete", mock.Anything, mock.Anything).Return(assert.AnError)

		assert.Equal(t, http.StatusInternalServerError, deleteRepo(t, user, repo))

		_, err := s.GetRepo(repo.ID)
		assert.NoError(t, err)
	})
}

func TestRepairRepoRotatesWebhookSecret(t *testing.T) {
	originalManager := server.Config.Services.Manager
	originalHost := server.Config.Server.WebhookHost
	t.Cleanup(func() {
		server.Config.Services.Manager = originalManager
		server.Config.Server.WebhookHost = originalHost
	})

	for _, test := range []struct {
		name          string
		all           bool
		lookupErr     error
		deactivateErr error
		activateErr   error
		remoteID      model.ForgeRemoteID
	}{
		{name: "single repository"},
		{name: "all repositories", all: true},
		{name: "deactivation fails", deactivateErr: errors.New("forge unavailable")},
		{name: "activation fails", activateErr: errors.New("forge unavailable")},
		{name: "repository lookup fails", lookupErr: errors.New("forge unavailable")},
		{name: "repository ID changes", remoteID: "456"},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newTestStore(t)
			user := &model.User{Login: "owner", ForgeID: 7}
			require.NoError(t, s.CreateUser(user))
			repo := &model.Repo{
				Owner: "owner", Name: "repo", FullName: "owner/repo", ForgeID: user.ForgeID,
				ForgeRemoteID: "123", WebhookSecret: "old-secret", UserID: user.ID, IsActive: true,
			}
			require.NoError(t, s.CreateRepo(repo))
			from := &model.Repo{
				Owner: repo.Owner, Name: repo.Name, FullName: repo.FullName,
				ForgeRemoteID: repo.ForgeRemoteID, Perm: &model.Perm{Admin: true},
			}
			if test.remoteID.IsValid() {
				from.ForgeRemoteID = test.remoteID
			}
			driver := forge_mocks.NewMockForge(t)
			manager := manager_mocks.NewMockManager(t)
			manager.On("ForgeFromRepo", mock.Anything).Return(driver, nil)
			server.Config.Services.Manager = manager
			server.Config.Server.WebhookHost = "https://woodpecker.example"
			driver.On("Repo", mock.Anything, mock.Anything, mock.Anything, repo.Owner, repo.Name).Return(from, test.lookupErr)
			driver.EXPECT().Org(mock.Anything, mock.Anything, repo.Owner).Return(&model.Org{Name: repo.Owner, ForgeRemoteID: "o-owner"}, nil).Maybe()

			var oldSecret string
			var activatedSecret *model.CloneSecret
			if test.lookupErr == nil {
				driver.EXPECT().Deactivate(mock.Anything, mock.Anything, mock.Anything, server.Config.Server.WebhookHost).
					RunAndReturn(func(_ context.Context, _ *model.User, repairing *model.Repo, _ string) error {
						stored, err := s.GetRepo(repairing.ID)
						require.NoError(t, err)
						assert.NotEqual(t, oldSecret, stored.WebhookSecret, "persist the new secret before replacing the webhook")
						assert.Equal(t, repairing.WebhookSecret, stored.WebhookSecret)
						return test.deactivateErr
					})
				driver.EXPECT().Activate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ *model.User, activated *model.Repo, hookURL string) error {
						stored, err := s.GetRepo(activated.ID)
						require.NoError(t, err)
						assert.Equal(t, stored.WebhookSecret, activated.WebhookSecret)
						uri, err := url.Parse(hookURL)
						require.NoError(t, err)
						signed := uri.Query().Get("access_token")
						parsed, err := token.Parse([]token.Type{token.HookToken}, signed, func(*token.Token) (string, error) {
							return stored.WebhookSecret, nil
						})
						require.NoError(t, err)
						assert.Equal(t, string(from.ForgeRemoteID), parsed.Get("repo-forge-remote-id"))
						assert.Equal(t, strconv.FormatInt(repo.ForgeID, 10), parsed.Get("forge-id"))
						_, err = token.Parse([]token.Type{token.HookToken}, signed, func(*token.Token) (string, error) {
							return oldSecret, nil
						})
						assert.Error(t, err, "the new URL must not use the old signing key")
						if test.activateErr == nil {
							activatedSecret = &model.CloneSecret{ForgeRemoteID: model.ForgeRemoteID("token-" + activated.WebhookSecret)}
							activated.CloneSecret = activatedSecret
						}
						return test.activateErr
					})
			}

			// Repeat repair so every recreation must revoke the preceding credential.
			for range 2 {
				stored, err := s.GetRepo(repo.ID)
				require.NoError(t, err)
				oldSecret = stored.WebhookSecret
				oldToken := token.New(token.HookToken)
				oldToken.Set("repo-forge-remote-id", string(stored.ForgeRemoteID))
				oldToken.Set("forge-id", strconv.FormatInt(stored.ForgeID, 10))
				oldSigned, err := oldToken.Sign(oldSecret)
				require.NoError(t, err)
				tc := newTestContext(t, s)
				withRepo(stored, &model.Perm{Admin: true})(tc)
				tc.Ctx.Request = httptest.NewRequest(http.MethodPost, "/repos/1/repair", nil)
				if test.all {
					RepairAllRepos(tc.Ctx)
				} else {
					RepairRepo(tc.Ctx)
				}

				expectedStatus := http.StatusNoContent
				if test.lookupErr != nil || test.activateErr != nil {
					expectedStatus = http.StatusInternalServerError
				}
				assert.Equal(t, expectedStatus, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())
				stored, err = s.GetRepo(repo.ID)
				require.NoError(t, err)
				_, err = token.Parse([]token.Type{token.HookToken}, oldSigned, func(*token.Token) (string, error) {
					return stored.WebhookSecret, nil
				})
				if test.lookupErr != nil {
					assert.Equal(t, oldSecret, stored.WebhookSecret)
					assert.NoError(t, err, "failed preflight must leave the existing webhook usable")
				} else {
					assert.NotEqual(t, oldSecret, stored.WebhookSecret)
					assert.Error(t, err, "old URLs must remain revoked even when activation fails")
				}
				if test.lookupErr == nil && test.activateErr == nil {
					assert.Equal(t, activatedSecret, stored.CloneSecret, "persist the clone secret written by activation")
				}
			}
		})
	}
}

func TestRepairRepoKeepsVisibility(t *testing.T) {
	originalManager := server.Config.Services.Manager
	originalHost := server.Config.Server.WebhookHost
	t.Cleanup(func() {
		server.Config.Services.Manager = originalManager
		server.Config.Server.WebhookHost = originalHost
	})

	for _, all := range []bool{false, true} {
		t.Run("all="+strconv.FormatBool(all), func(t *testing.T) {
			s := newTestStore(t)
			user := &model.User{Login: "owner", ForgeID: 7}
			require.NoError(t, s.CreateUser(user))
			repo := &model.Repo{
				Owner: "owner", Name: "repo", FullName: "owner/repo", ForgeID: user.ForgeID, ForgeRemoteID: "123",
				UserID: user.ID, IsActive: true, Visibility: model.VisibilityPrivate,
			}
			require.NoError(t, s.CreateRepo(repo))

			driver := moveRepoForge(t)
			driver.On("Repo", mock.Anything, mock.Anything, repo.ForgeRemoteID, repo.Owner, repo.Name).Return(&model.Repo{
				Owner: repo.Owner, Name: repo.Name, FullName: repo.FullName, ForgeRemoteID: repo.ForgeRemoteID,
				Visibility: model.VisibilityPublic, Perm: &model.Perm{Admin: true},
			}, nil)
			driver.On("Org", mock.Anything, mock.Anything, repo.Owner).Return(&model.Org{Name: repo.Owner, ForgeRemoteID: "o-owner"}, nil)
			driver.On("Deactivate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			driver.On("Activate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			tc := newTestContext(t, s)
			withRepo(repo, &model.Perm{Admin: true})(tc)
			if all {
				RepairAllRepos(tc.Ctx)
			} else {
				RepairRepo(tc.Ctx)
			}
			require.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())

			stored, err := s.GetRepo(repo.ID)
			require.NoError(t, err)
			assert.False(t, stored.VisibilitySync)
			assert.Equal(t, model.VisibilityPrivate, stored.Visibility)
		})
	}
}

func TestPostRepoVisibilitySync(t *testing.T) {
	originalManager := server.Config.Services.Manager
	originalHost := server.Config.Server.WebhookHost
	originalAllowlist := server.Config.Permissions.OwnersAllowlist
	t.Cleanup(func() {
		server.Config.Services.Manager = originalManager
		server.Config.Server.WebhookHost = originalHost
		server.Config.Permissions.OwnersAllowlist = originalAllowlist
	})
	server.Config.Server.WebhookHost = "https://woodpecker.example"
	server.Config.Permissions.OwnersAllowlist = permissions.NewOwnersAllowlist(nil)

	activate := func(t *testing.T, s store.Store, user *model.User, remoteID model.ForgeRemoteID) {
		t.Helper()
		mgr := manager_mocks.NewMockManager(t)
		driver := forge_mocks.NewMockForge(t)
		mgr.On("ForgeFromUser", user).Return(driver, nil)
		server.Config.Services.Manager = mgr
		driver.On("Repo", mock.Anything, user, remoteID, "", "").Return(&model.Repo{
			ForgeRemoteID: remoteID, Owner: "acme", Name: "rocket", FullName: "acme/rocket",
			Visibility: model.VisibilityPrivate, IsSCMPrivate: true, Perm: &model.Perm{Admin: true},
		}, nil)
		driver.On("Org", mock.Anything, user, "acme").Return(&model.Org{Name: "acme", ForgeRemoteID: "o-acme"}, nil)
		driver.On("Activate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)

		tc := newTestContext(t, s)
		withUser(user)(tc)
		tc.Ctx.Request = httptest.NewRequest(http.MethodPost, "/repos?forge_remote_id="+string(remoteID), nil)
		PostRepo(tc.Ctx)
		require.Equal(t, http.StatusOK, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())
	}

	t.Run("new repos sync their visibility", func(t *testing.T) {
		s := newTestStore(t)
		user := &model.User{Login: "alice", ForgeID: 1}
		require.NoError(t, s.CreateUser(user))
		require.NoError(t, s.OrgCreate(&model.Org{Name: "acme", ForgeID: 1}))

		activate(t, s, user, "42")

		stored, err := s.GetRepoForgeID(1, "42")
		require.NoError(t, err)
		assert.True(t, stored.VisibilitySync)
		assert.Equal(t, model.VisibilityPrivate, stored.Visibility)
	})

	t.Run("re-activation keeps the sync setting", func(t *testing.T) {
		for sync, visibility := range map[bool]model.RepoVisibility{false: model.VisibilityInternal, true: model.VisibilityPrivate} {
			s := newTestStore(t)
			user := &model.User{Login: "alice", ForgeID: 1}
			require.NoError(t, s.CreateUser(user))
			require.NoError(t, s.OrgCreate(&model.Org{Name: "acme", ForgeID: 1}))
			require.NoError(t, s.CreateRepo(&model.Repo{
				ForgeID: 1, ForgeRemoteID: "42", Owner: "acme", Name: "rocket", FullName: "acme/rocket",
				Visibility: model.VisibilityInternal, VisibilitySync: sync, WebhookSecret: "secret",
			}))

			activate(t, s, user, "42")

			stored, err := s.GetRepoForgeID(1, "42")
			require.NoError(t, err)
			assert.Equal(t, sync, stored.VisibilitySync)
			assert.Equal(t, visibility, stored.Visibility)
			assert.True(t, stored.IsSCMPrivate)
		}
	})
}

func TestPatchRepoVisibilitySync(t *testing.T) {
	s := newTestStore(t)
	user := &model.User{Login: "alice", ForgeID: 1}
	require.NoError(t, s.CreateUser(user))
	repo := &model.Repo{ForgeID: 1, ForgeRemoteID: "42", Owner: "acme", Name: "rocket", FullName: "acme/rocket", Visibility: model.VisibilityPrivate}
	require.NoError(t, s.CreateRepo(repo))

	patch := func(t *testing.T, in *model.RepoPatch) *model.Repo {
		t.Helper()
		tc := newTestContext(t, s)
		withUser(user)(tc)
		withRepo(repo, &model.Perm{Admin: true})(tc)
		withRequest(http.MethodPatch, in)(tc)
		PatchRepo(tc.Ctx)
		require.Equal(t, http.StatusOK, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())
		stored, err := s.GetRepo(repo.ID)
		require.NoError(t, err)
		return stored
	}

	stored := patch(t, &model.RepoPatch{VisibilitySync: new(true), Visibility: new(string(model.VisibilityPublic))})
	assert.True(t, stored.VisibilitySync)
	assert.Equal(t, model.VisibilityPrivate, stored.Visibility, "the sync overwrites the visibility anyway")

	stored = patch(t, &model.RepoPatch{VisibilitySync: new(false), Visibility: new(string(model.VisibilityPublic))})
	assert.False(t, stored.VisibilitySync)
	assert.Equal(t, model.VisibilityPublic, stored.Visibility)
}

func TestDeactivateRepoDropsCloneSecret(t *testing.T) {
	s := newTestStore(t)
	user := &model.User{Login: "carol", ForgeID: 1, ForgeRemoteID: "u-carol", Hash: "userhash-carol"}
	require.NoError(t, s.CreateUser(user))
	repo := &model.Repo{
		ForgeID:       1,
		ForgeRemoteID: "r-carol",
		Owner:         "acme",
		Name:          "carol",
		FullName:      "acme/carol",
		UserID:        user.ID,
		WebhookSecret: "hash-carol",
		IsActive:      true,
		CloneSecret:   &model.CloneSecret{Login: "deploy", Token: "secret", ForgeRemoteID: "1"},
	}
	require.NoError(t, s.CreateRepo(repo))

	mockForge := moveRepoForge(t)
	mockForge.On("Deactivate", mock.Anything, user, mock.Anything, mock.Anything).Return(nil)

	tc := newTestContext(t, s)
	withUser(user)(tc)
	withRepo(repo, &model.Perm{Admin: true})(tc)
	tc.Ctx.Request = httptest.NewRequest(http.MethodDelete, "/repos/1", nil)
	DeleteRepo(tc.Ctx)
	require.Equal(t, http.StatusOK, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())

	stored, err := s.GetRepo(repo.ID)
	require.NoError(t, err)
	assert.False(t, stored.IsActive)
	assert.Nil(t, stored.CloneSecret, "an inactive repo must not keep clone credentials")
}

// TestSyncFollowsRepoOwnerChanges moves a repo between owners on the forge and
// checks after every change that a sync links the repo to the right org.
// Orgs are told apart by their forge remote id, not by their name.
func TestSyncFollowsRepoOwnerChanges(t *testing.T) {
	originalManager := server.Config.Services.Manager
	originalHost := server.Config.Server.WebhookHost
	t.Cleanup(func() {
		server.Config.Services.Manager = originalManager
		server.Config.Server.WebhookHost = originalHost
	})

	syncs := map[string]func(t *testing.T, s store.Store, user *model.User, repoID int64){
		"hook": func(t *testing.T, s store.Store, _ *model.User, repoID int64) {
			repo, err := s.GetRepo(repoID)
			require.NoError(t, err)
			hookToken := token.New(token.HookToken)
			hookToken.Set("repo-forge-remote-id", string(repo.ForgeRemoteID))
			hookToken.Set("forge-id", strconv.FormatInt(repo.ForgeID, 10))
			signed, err := hookToken.Sign(repo.WebhookSecret)
			require.NoError(t, err)

			tc := newTestContext(t, s)
			tc.Ctx.Request = httptest.NewRequest(http.MethodPost, "/api/hook", nil)
			tc.Ctx.Request.Header.Set("Authorization", "Bearer "+signed)
			PostHook(tc.Ctx)
			require.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())
		},
		"repair": func(t *testing.T, s store.Store, user *model.User, repoID int64) {
			repo, err := s.GetRepo(repoID)
			require.NoError(t, err)

			tc := newTestContext(t, s)
			withUser(user)(tc)
			withRepo(repo, &model.Perm{Admin: true})(tc)
			RepairRepo(tc.Ctx)
			require.Equal(t, http.StatusNoContent, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())
		},
	}

	for name, sync := range syncs {
		t.Run(name, func(t *testing.T) {
			// the forge: the owner of the repo and the current name of each org by its remote id
			repoOwner := "alpha"
			orgs := map[model.ForgeRemoteID]string{"o1": "alpha"}
			forgeRepo := func() *model.Repo {
				return &model.Repo{
					ForgeRemoteID: "42", Owner: repoOwner, Name: "rocket", FullName: repoOwner + "/rocket",
					Perm: &model.Perm{Admin: true},
				}
			}

			driver := moveRepoForge(t)
			driver.EXPECT().Repo(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				RunAndReturn(func(context.Context, *model.User, model.ForgeRemoteID, string, string) (*model.Repo, error) {
					return forgeRepo(), nil
				}).Maybe()
			driver.EXPECT().Hook(mock.Anything, mock.Anything, mock.Anything).
				RunAndReturn(func(context.Context, *http.Request, string) (*model.Repo, *model.Pipeline, error) {
					// a pull request event ends the hook after the sync, as pull requests are not allowed
					return forgeRepo(), &model.Pipeline{Event: model.EventPull}, nil
				}).Maybe()
			driver.EXPECT().Org(mock.Anything, mock.Anything, mock.Anything).
				RunAndReturn(func(_ context.Context, _ *model.User, name string) (*model.Org, error) {
					for id, orgName := range orgs {
						if orgName == name {
							return &model.Org{Name: orgName, ForgeRemoteID: id}, nil
						}
					}
					return nil, forge_types.ErrOrgNotFound
				})
			driver.EXPECT().OrgByRemoteID(mock.Anything, mock.Anything, mock.Anything, false).
				RunAndReturn(func(_ context.Context, _ *model.User, id model.ForgeRemoteID, _ bool) (*model.Org, error) {
					if orgName, ok := orgs[id]; ok {
						return &model.Org{Name: orgName, ForgeRemoteID: id}, nil
					}
					return nil, forge_types.ErrOrgNotFound
				}).Maybe()
			driver.EXPECT().Deactivate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			driver.EXPECT().Activate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

			s := newTestStore(t)
			user := &model.User{Login: "alice", ForgeID: 1, ForgeRemoteID: "u-alice", Hash: "alice"}
			require.NoError(t, s.CreateUser(user))
			alpha := &model.Org{Name: "alpha", ForgeID: 1, ForgeRemoteID: "o1"}
			require.NoError(t, s.OrgCreate(alpha))
			repo := &model.Repo{
				ForgeID: 1, ForgeRemoteID: "42", Owner: "alpha", Name: "rocket", FullName: "alpha/rocket",
				UserID: user.ID, IsActive: true, WebhookSecret: "secret", OrgID: alpha.ID,
			}
			require.NoError(t, s.CreateRepo(repo))

			for _, step := range []struct {
				name   string
				change func()
				org    model.ForgeRemoteID
			}{
				{
					name:   "repo moved to a new owner",
					change: func() { orgs["o2"] = "bravo"; repoOwner = "bravo" },
					org:    "o2",
				},
				{
					name:   "new owner renamed itself",
					change: func() { orgs["o2"] = "charlie"; repoOwner = "charlie" },
					org:    "o2",
				},
				{
					name:   "owner renamed again and a new org with its old name got the repo",
					change: func() { orgs["o2"] = "delta"; orgs["o3"] = "charlie"; repoOwner = "charlie" },
					org:    "o3",
				},
			} {
				step.change()
				sync(t, s, user, repo.ID)

				stored, err := s.GetRepo(repo.ID)
				require.NoError(t, err)
				assert.Equal(t, repoOwner+"/rocket", stored.FullName, step.name)
				org, err := s.OrgGet(stored.OrgID)
				require.NoError(t, err)
				assert.Equal(t, step.org, org.ForgeRemoteID, "%s: the repo must be linked to the org that owns it now", step.name)
				// stored orgs keep their forge remote id and follow renames
				for id, orgName := range orgs {
					org, err := s.OrgFindByRemoteID(1, id)
					require.NoError(t, err, step.name)
					assert.Equal(t, orgName, org.Name, step.name)
				}
			}
		})
	}
}
