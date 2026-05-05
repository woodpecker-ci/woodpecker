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

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
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
