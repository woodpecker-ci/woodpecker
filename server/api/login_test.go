// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/api"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
)

func TestHandleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user := &model.User{
		ID:            1,
		OrgID:         1,
		ForgeID:       1,
		ForgeRemoteID: "remote-id-1",
		Login:         "test",
		Email:         "test@example.com",
		Admin:         false,
	}
	org := &model.Org{
		ID:   1,
		Name: user.Login,
	}

	server.Config.Server.SessionExpires = time.Hour

	t.Run("should handle errors from the callback", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		query := url.Values{}
		query.Set("error", "invalid_scope")
		query.Set("error_description", "The requested scope is invalid, unknown, or malformed")
		query.Set("error_uri", "https://developer.atlassian.com/cloud/jira/platform/rest/#api-group-OAuth2-ErrorHandling")

		c.Request = &http.Request{
			Header: make(http.Header),
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme:   "https",
				Path:     "/authorize",
				RawQuery: query.Encode(),
			},
		}

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, fmt.Sprintf("/login?%s", query.Encode()), c.Writer.Header().Get("Location"))
	})

	t.Run("should fail if the state is wrong", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)

		query := url.Values{}
		query.Set("code", "assumed_to_be_valid_code")

		wrongToken := token.New(token.OAuthStateToken)
		wrongToken.Set("forge_id", "1")
		signedWrongToken, _ := wrongToken.Sign("wrong_secret")
		query.Set("state", signedWrongToken)

		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme:   "https",
				RawQuery: query.Encode(),
			},
		}

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=invalid_state", c.Writer.Header().Get("Location"))
	})

	t.Run("should redirect to forge login page", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)

		forgeRedirectURL := ""
		_forge.On("Login", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			state, ok := args.Get(1).(*forge_types.OAuthRequest)
			if ok {
				forgeRedirectURL = fmt.Sprintf("https://my-awesome-forge.com/oauth/authorize?client_id=client-id&state=%s", state.State)
			}
		}).Return(nil, func(context.Context, *forge_types.OAuthRequest) string {
			return forgeRedirectURL
		}, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, forgeRedirectURL, c.Writer.Header().Get("Location"))
	})

	t.Run("should register a new user", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(nil, types.ErrRecordNotExist)
		_store.On("GetUserByLogin", user.ForgeID, user.Login).Return(nil, types.ErrRecordNotExist)
		_store.On("CreateUser", mock.Anything).Return(nil)
		_store.On("OrgFindByName", user.Login, user.ForgeID).Return(nil, nil)
		_store.On("OrgCreate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
		_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should login an existing user", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(user, nil)
		_store.On("OrgGet", org.ID).Return(org, nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
		_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should deny a new user if registration is closed", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = false
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(nil, types.ErrRecordNotExist)
		_store.On("GetUserByLogin", user.ForgeID, user.Login).Return(nil, types.ErrRecordNotExist)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=registration_closed", c.Writer.Header().Get("Location"))
		// a rejected login must not persist any (broken) user/org row, see #6769
		_store.AssertNotCalled(t, "CreateUser", mock.Anything)
		_store.AssertNotCalled(t, "OrgCreate", mock.Anything)
	})

	t.Run("should deny a login if the forge cannot be loaded", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs([]string{"org1"})
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		// without the forge we cannot tell which orgs are allowed, so the login must fail
		_store.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=internal_error", c.Writer.Header().Get("Location"))
		_forge.AssertNotCalled(t, "Teams", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("should stop requesting teams once a page is not full", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs([]string{"org1"})
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		teamsCalls := 0
		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		// a single team is less than a full page, so there is nothing more to fetch
		_forge.On("Teams", mock.Anything, user, mock.Anything).Run(func(mock.Arguments) {
			teamsCalls++
		}).Return([]*model.Team{
			{
				Login: "org2",
			},
		}, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=org_access_denied", c.Writer.Header().Get("Location"))
		// without stopping early this would walk through all maxPage pages
		assert.Equal(t, 1, teamsCalls)
	})

	t.Run("should stop requesting teams if the forge does not implement it", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs([]string{"org1"})
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}

		teamsCalls := 0
		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_forge.On("Teams", mock.Anything, user, mock.Anything).Run(func(mock.Arguments) {
			teamsCalls++
		}).Return(nil, forge_types.ErrNotImplemented)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/login?error=org_access_denied", c.Writer.Header().Get("Location"))
		assert.Equal(t, 1, teamsCalls)
	})

	t.Run("should create an user org if it does not exists", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		user.OrgID = 0

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(user, nil)
		_store.On("OrgFindByName", user.Login, user.ForgeID).Return(nil, types.ErrRecordNotExist)
		_store.On("OrgCreate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
		_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should link an user org if it has the same name as the user", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		user.OrgID = 0

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(user, nil)
		_store.On("OrgFindByName", user.Login, user.ForgeID).Return(org, nil)
		_store.On("OrgUpdate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
		_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})

	t.Run("should update an user org if the user name was changed", func(t *testing.T) {
		_manager := manager_mocks.NewMockManager(t)
		_forge := forge_mocks.NewMockForge(t)
		_store := store_mocks.NewMockStore(t)
		server.Config.Services.Manager = _manager
		server.Config.Permissions.Open = true
		server.Config.Permissions.Orgs = permissions.NewOrgs(nil)
		server.Config.Permissions.Admins = permissions.NewAdmins(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("store", _store)
		c.Request = &http.Request{
			Header: make(http.Header),
			URL: &url.URL{
				Scheme: "https",
			},
		}
		org.Name = "not-the-user-name"

		_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)
		_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(user, nil)
		_store.On("OrgGet", user.OrgID).Return(org, nil)
		_store.On("OrgUpdate", mock.Anything).Return(nil)
		_store.On("UpdateUser", mock.Anything).Return(nil)
		_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
		_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		api.HandleAuth(c)

		assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
		assert.Equal(t, "/", c.Writer.Header().Get("Location"))
		assert.NotEmpty(t, c.Writer.Header().Get("Set-Cookie"))
	})
}

// TestHandleAuthAllowedOrgs walks through the combinations of WOODPECKER_ORGS
// and the orgs of the forge a user logs in with. The global list applies to
// every forge, the orgs of a forge are allowed in addition to it. Org names are
// only matched against the memberships reported by that very forge, a name
// allowed for one forge never grants access on another one (#6852).
func TestHandleAuthAllowedOrgs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	user := &model.User{
		ID:            1,
		OrgID:         1,
		ForgeID:       1,
		ForgeRemoteID: "remote-id-1",
		Login:         "test",
		Email:         "test@example.com",
	}
	org := &model.Org{
		ID:   1,
		Name: user.Login,
	}

	server.Config.Server.SessionExpires = time.Hour

	tests := []struct {
		name       string
		globalOrgs []string
		forgeOrgs  []string
		teams      []string
		allow      bool
	}{
		{
			name:  "nothing configured lets everybody in",
			teams: []string{"github-org"},
			allow: true,
		},
		{
			name:       "global org matching a membership of this forge",
			globalOrgs: []string{"github-org"},
			teams:      []string{"github-org"},
			allow:      true,
		},
		{
			name:       "global org matching a membership on another forge only",
			globalOrgs: []string{"github-org"},
			teams:      []string{"gitlab-org"},
			allow:      false,
		},
		{
			name:      "forge org matching a membership of this forge",
			forgeOrgs: []string{"github-org"},
			teams:     []string{"github-org"},
			allow:     true,
		},
		{
			name:  "orgs of another forge do not gate this one",
			teams: []string{"gitlab-org"},
			allow: true,
		},
		{
			name:      "forge org naming a group of another forge",
			forgeOrgs: []string{"gitlab-org"},
			teams:     []string{"github-org"},
			allow:     false,
		},
		{
			name:       "global org while the forge lists other orgs",
			globalOrgs: []string{"github-org"},
			forgeOrgs:  []string{"dummy-org"},
			teams:      []string{"github-org"},
			allow:      true,
		},
		{
			name:       "forge org while the global list holds other orgs",
			globalOrgs: []string{"dummy-org"},
			forgeOrgs:  []string{"gitlab-org"},
			teams:      []string{"gitlab-org"},
			allow:      true,
		},
		{
			name:       "member of neither the global nor the forge orgs",
			globalOrgs: []string{"dummy"},
			forgeOrgs:  []string{"dummy-org"},
			teams:      []string{"github-org"},
			allow:      false,
		},
		{
			name:       "orgs are matched case-insensitively",
			globalOrgs: []string{"GITHUB-ORG"},
			forgeOrgs:  []string{"GITLAB-ORG"},
			teams:      []string{"gitlab-org"},
			allow:      true,
		},
		{
			name:      "membership in a parent group is not membership in a subgroup",
			forgeOrgs: []string{"gitlab-org/some-subgroup"},
			teams:     []string{"gitlab-org"},
			allow:     false,
		},
		{
			name:       "no membership matches although both lists are set",
			globalOrgs: []string{"github-org"},
			forgeOrgs:  []string{"github-org"},
			teams:      []string{"gitlab-org"},
			allow:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_manager := manager_mocks.NewMockManager(t)
			_forge := forge_mocks.NewMockForge(t)
			_store := store_mocks.NewMockStore(t)
			server.Config.Services.Manager = _manager
			server.Config.Permissions.Open = true
			server.Config.Permissions.Orgs = permissions.NewOrgs(tc.globalOrgs)
			server.Config.Permissions.Admins = permissions.NewAdmins(nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("store", _store)
			c.Request = &http.Request{
				Header: make(http.Header),
				URL: &url.URL{
					Scheme: "https",
				},
			}

			_manager.On("ForgeByID", int64(1)).Return(_forge, nil)
			_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1, Orgs: tc.forgeOrgs}, nil)
			_forge.On("Login", mock.Anything, mock.Anything).Return(user, "", nil)

			// memberships are only requested if there is a list to check against
			if len(tc.globalOrgs)+len(tc.forgeOrgs) > 0 {
				teams := make([]*model.Team, 0, len(tc.teams))
				for _, team := range tc.teams {
					teams = append(teams, &model.Team{Login: team})
				}
				_forge.On("Teams", mock.Anything, user, mock.Anything).Return(teams, nil)
			}

			if tc.allow {
				_store.On("GetUserByRemoteID", user.ForgeID, user.ForgeRemoteID).Return(user, nil)
				_store.On("OrgGet", org.ID).Return(org, nil)
				_store.On("UpdateUser", mock.Anything).Return(nil)
				_store.On("PermPrune", mock.Anything, []int64(nil)).Return(nil)
				_store.On("RepoList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				_forge.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			}

			api.HandleAuth(c)

			assert.Equal(t, http.StatusSeeOther, c.Writer.Status())
			if tc.allow {
				assert.Equal(t, "/", c.Writer.Header().Get("Location"))
			} else {
				assert.Equal(t, "/login?error=org_access_denied", c.Writer.Header().Get("Location"))
			}
		})
	}
}
