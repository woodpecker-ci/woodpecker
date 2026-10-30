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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestGetWorkflowAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &model.Repo{ID: 1, OrgID: 5}
	pipeline := &model.Pipeline{ID: 10, RepoID: 1, Number: 3}
	user := &model.User{ID: 2, Login: "alice", ForgeID: 1}

	globalAgent := &model.AgentSnapshot{ID: 101, OrgID: model.IDNotSet, Name: "global-builder", CustomLabels: make(map[string]string)}
	orgAgent := &model.AgentSnapshot{ID: 102, OrgID: 5, Name: "acme-builder", CustomLabels: map[string]string{"zone": "eu"}}
	otherOrgAgent := &model.AgentSnapshot{ID: 103, OrgID: 6, Name: "other-builder"}
	snapshots := map[int64]*model.AgentSnapshot{101: globalAgent, 102: orgAgent, 103: otherOrgAgent}
	// a reference to a snapshot that is gone, which should not happen
	const danglingSnapshotID = 104

	tests := []struct {
		name       string
		workflowID string
		workflow   *model.Workflow
		loadErr    error
		wantStatus int
		wantAgent  *model.AgentSnapshot
	}{
		{
			name:       "user reads a global agent",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7, AgentSnapshotID: globalAgent.ID},
			wantStatus: http.StatusOK,
			wantAgent:  globalAgent,
		},
		{
			name:       "the snapshot is shared, the agent id comes from the workflow",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 70, AgentSnapshotID: globalAgent.ID},
			wantStatus: http.StatusOK,
			wantAgent:  globalAgent,
		},
		{
			name:       "user reads an agent of the repo's org",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 8, AgentSnapshotID: orgAgent.ID},
			wantStatus: http.StatusOK,
			wantAgent:  orgAgent,
		},
		{
			name:       "user reads an agent of another org",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 9, AgentSnapshotID: otherOrgAgent.ID},
			wantStatus: http.StatusOK,
			wantAgent:  otherOrgAgent,
		},
		{
			name:       "not found without a snapshot",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "not found for a dangling snapshot reference",
			workflow:   &model.Workflow{ID: 20, PipelineID: 10, AgentID: 7, AgentSnapshotID: danglingSnapshotID},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "workflow of another pipeline is not found",
			workflow:   &model.Workflow{ID: 20, PipelineID: 11, AgentID: 7, AgentSnapshotID: globalAgent.ID},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown workflow is not found",
			loadErr:    types.ErrRecordNotExist,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid workflow id",
			workflowID: "abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store_mocks.NewMockStore(t)
			if tt.workflowID == "" {
				tt.workflowID = "20"
				mockStore.On("WorkflowLoad", int64(20)).Return(tt.workflow, tt.loadErr)
			}
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
			c.Set("user", user)
			c.Params = gin.Params{{Key: "workflow_id", Value: tt.workflowID}}

			GetWorkflowAgent(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			if tt.wantAgent != nil {
				var got model.AgentSnapshot
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
				want := *tt.wantAgent
				// the snapshot's own id and hash are internal, the agent id is the workflow's
				want.ID = 0
				want.AgentID = tt.workflow.AgentID
				assert.Equal(t, want, got)
				assert.NotContains(t, w.Body.String(), `"hash"`)
				assert.Zero(t, tt.wantAgent.AgentID, "the loaded snapshot must not be changed")
			} else {
				assert.NotContains(t, w.Body.String(), "builder")
			}
		})
	}
}
