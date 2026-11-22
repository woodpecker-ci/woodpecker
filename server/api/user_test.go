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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/permissions"
)

func TestGetReposAllShowsVisibilitySyncDefault(t *testing.T) {
	originalManager := server.Config.Services.Manager
	originalAllowlist := server.Config.Permissions.OwnersAllowlist
	t.Cleanup(func() {
		server.Config.Services.Manager = originalManager
		server.Config.Permissions.OwnersAllowlist = originalAllowlist
	})
	server.Config.Permissions.OwnersAllowlist = permissions.NewOwnersAllowlist(nil)

	s := newTestStore(t)
	user := &model.User{Login: "alice", ForgeID: 1}
	require.NoError(t, s.CreateUser(user))
	known := &model.Repo{ForgeID: 1, ForgeRemoteID: "1", Owner: "acme", Name: "known", FullName: "acme/known", Visibility: model.VisibilityPublic}
	require.NoError(t, s.CreateRepo(known))
	require.NoError(t, s.PermUpsert(&model.Perm{UserID: user.ID, RepoID: known.ID, Pull: true, Push: true, Admin: true}))

	mgr := manager_mocks.NewMockManager(t)
	driver := forge_mocks.NewMockForge(t)
	mgr.On("ForgeFromUser", user).Return(driver, nil)
	server.Config.Services.Manager = mgr
	firstPage := mock.MatchedBy(func(o *model.ListOptions) bool { return o.Page == 1 })
	driver.On("Repos", mock.Anything, user, firstPage).Return([]*model.Repo{
		{ForgeRemoteID: "1", Owner: "acme", Name: "known", FullName: "acme/known", Perm: &model.Perm{Pull: true, Push: true, Admin: true}},
		{ForgeRemoteID: "2", Owner: "acme", Name: "new", FullName: "acme/new", Perm: &model.Perm{Pull: true, Push: true, Admin: true}},
	}, nil)
	driver.On("Repos", mock.Anything, user, mock.Anything).Return(nil, nil)

	tc := newTestContext(t, s)
	withUser(user)(tc)
	tc.Ctx.Request = httptest.NewRequest(http.MethodGet, "/user/repos?all=true", nil)
	GetRepos(tc.Ctx)
	require.Equal(t, http.StatusOK, tc.Ctx.Writer.Status(), tc.Recorder.Body.String())

	var repos []*model.Repo
	tc.decodeJSON(t, &repos)
	sync := make(map[string]bool)
	for _, r := range repos {
		sync[r.FullName] = r.VisibilitySync
	}
	assert.Equal(t, map[string]bool{"acme/known": false, "acme/new": true}, sync)
}
