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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestGetWorkflowAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &model.Repo{ID: 1, OrgID: 5}
	org := &model.Org{ID: 5, Name: "acme", ForgeID: 1}
	pipeline := &model.Pipeline{ID: 10, RepoID: 1, Number: 3}
	admin := &model.User{ID: 1, Login: "root", ForgeID: 1, Admin: true}
	user := &model.User{ID: 2, Login: "alice", ForgeID: 1}

	globalAgent := &model.AgentSnapshot{ID: 101, AgentID: 7, OrgID: model.IDNotSet, Name: "global-builder", CustomLabels: map[string]string{}}
	orgAgent := &model.AgentSnapshot{ID: 102, AgentID: 8, OrgID: 5, Name: "acme-builder", CustomLabels: map[string]string{"zone": "eu"}}
	otherOrgAgent := &model.AgentSnapshot{ID: 103, AgentID: 9, OrgID: 6, Name: "other-builder"}
	snapshots := map[int64]*model.AgentSnapshot{101: globalAgent, 102: orgAgent, 103: otherOrgAgent}
	// a reference to a snapshot that is gone, which should not happen
	const danglingSnapshotID = 104

	tests := []struct {
		name       string
		user       *model.User
		workflowID string
		workflow   *model.Workflow
		loadErr    error
		orgPerm    *model.OrgPerm
		// server.Config.Agent.DisableUserRegisteredAgentRegistration
		disableOrgAPI bool
		wantStatus    int
		wantAgent     *model.AgentSnapshot
		// org admin membership must only be checked for agents of the repo's org
		wantMembershipCall bool
	}{
		{
			name:       "instance admin reads a global agent",
			user:       admin,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7, AgentSnapshotID: globalAgent.ID},
			wantStatus: http.StatusOK,
			wantAgent:  globalAgent,
		},
		{
			name:       "instance admin gets not found without a snapshot",
			user:       admin,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7},
			wantStatus: http.StatusNotFound,
		},
		{
			name:               "org admin reads an agent of the repo's org",
			user:               user,
			workflow:           &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8, AgentSnapshotID: orgAgent.ID},
			orgPerm:            &model.OrgPerm{Member: true, Admin: true},
			wantStatus:         http.StatusOK,
			wantAgent:          orgAgent,
			wantMembershipCall: true,
		},
		{
			name:               "org member without admin is forbidden",
			user:               user,
			workflow:           &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8, AgentSnapshotID: orgAgent.ID},
			orgPerm:            &model.OrgPerm{Member: true},
			wantStatus:         http.StatusForbidden,
			wantMembershipCall: true,
		},
		{
			name:          "org admin is forbidden when user agent registration is disabled, like the org agent api",
			user:          user,
			disableOrgAPI: true,
			workflow:      &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8, AgentSnapshotID: orgAgent.ID},
			orgPerm:       &model.OrgPerm{Member: true, Admin: true},
			wantStatus:    http.StatusForbidden,
		},
		{
			name:       "org admin is forbidden to read a global agent",
			user:       user,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7, AgentSnapshotID: globalAgent.ID},
			orgPerm:    &model.OrgPerm{Member: true, Admin: true},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "org admin is forbidden to read an agent of another org",
			user:       user,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 9, AgentSnapshotID: otherOrgAgent.ID},
			orgPerm:    &model.OrgPerm{Member: true, Admin: true},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non admin cannot tell a missing snapshot from a forbidden one",
			user:       user,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8},
			orgPerm:    &model.OrgPerm{Member: true, Admin: true},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "instance admin gets not found for a dangling snapshot reference",
			user:       admin,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7, AgentSnapshotID: danglingSnapshotID},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "non admin gets forbidden for a dangling snapshot reference",
			user:       user,
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8, AgentSnapshotID: danglingSnapshotID},
			orgPerm:    &model.OrgPerm{Member: true, Admin: true},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "workflow of another pipeline is not found, even for admins",
			user:       admin,
			workflow:   &model.Workflow{ID: 20, PipelineID: 11, AgentID: 7, AgentSnapshotID: globalAgent.ID},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown workflow is not found",
			user:       admin,
			loadErr:    types.ErrRecordNotExist,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid workflow id",
			user:       admin,
			workflowID: "abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installOrgForgeManager(t)
			membership := &fakeMembership{perm: tt.orgPerm}
			server.Config.Services.Membership = membership
			server.Config.Agent.DisableUserRegisteredAgentRegistration = tt.disableOrgAPI
			t.Cleanup(func() { server.Config.Agent.DisableUserRegisteredAgentRegistration = false })

			mockStore := store_mocks.NewMockStore(t)
			if tt.workflowID == "" {
				tt.workflowID = "20"
				mockStore.On("WorkflowLoad", int64(20)).Return(tt.workflow, tt.loadErr)
			}
			mockStore.On("OrgGet", int64(5)).Return(org, nil).Maybe()
			for id, snapshot := range snapshots {
				mockStore.On("AgentSnapshotFind", id).Return(snapshot, nil).Maybe()
			}
			mockStore.On("AgentSnapshotFind", int64(danglingSnapshotID)).Return(nil, types.ErrRecordNotExist).Maybe()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Set("store", mockStore)
			c.Set("repo", repo)
			c.Set("pipeline", pipeline)
			c.Set("user", tt.user)
			c.Params = gin.Params{{Key: "workflow_id", Value: tt.workflowID}}

			GetWorkflowAgent(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			assert.Equal(t, tt.wantMembershipCall, membership.called)
			if tt.wantAgent != nil {
				var got model.AgentSnapshot
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
				want := *tt.wantAgent
				// the snapshot's own id and hash are internal
				want.ID = 0
				assert.Equal(t, want, got)
				assert.NotContains(t, w.Body.String(), `"hash"`)
			} else {
				assert.NotContains(t, w.Body.String(), "builder")
			}
		})
	}
}
