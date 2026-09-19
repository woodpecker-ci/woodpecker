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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestCancelWorkflowRoute(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		code     int
	}{
		{"running", "9", 204},
		{"pending", "9", 204},
		{"anonymous", "9", 401},
		{"read only", "9", 404},
		{"malformed", "no", 400},
		{"zero", "0", 400},
		{"negative", "-1", 400},
		{"missing", "9", 404},
		{"cross pipeline", "9", 404},
		{"cross repository", "9", 404},
		{"blocked", "9", 400},
		{"finished", "9", 400},
		{"terminal parent", "9", 400},
		{"queue error", "9", 500},
		{"storage error", "9", 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			s := store_mocks.NewMockStore(t)
			f := forge_mocks.NewMockForge(t)
			manager := manager_mocks.NewMockManager(t)
			sched := mocks.NewMockScheduler(t)
			oldManager, oldSched := server.Config.Services.Manager, server.Config.Services.Scheduler
			server.Config.Services.Manager = manager
			server.Config.Services.Scheduler = sched
			t.Cleanup(func() { server.Config.Services.Manager = oldManager; server.Config.Services.Scheduler = oldSched })
			repo := &model.Repo{ID: 1, UserID: 2}
			user := &model.User{ID: 2}
			pl := &model.Pipeline{ID: 3, Number: 7, RepoID: 1, Status: model.StatusRunning}
			w := &model.Workflow{ID: 9, PipelineID: 3, State: model.StatusRunning}
			denied := tc.name == "anonymous" || tc.name == "read only"
			malformed := tc.name == "malformed" || tc.name == "zero" || tc.name == "negative"
			if !denied {
				s.On("GetPipelineNumber", repo, int64(7)).Return(pl, nil)
			}
			if !denied && !malformed {
				manager.On("ForgeFromRepo", repo).Return(f, nil)
				if tc.name == "missing" {
					s.On("WorkflowLoad", int64(9)).Return(nil, types.ErrRecordNotExist)
				} else {
					if tc.name == "cross pipeline" || tc.name == "cross repository" {
						w.PipelineID = 4
					}
					if tc.name == "blocked" {
						w.State = model.StatusBlocked
					}
					if tc.name == "finished" {
						w.State = model.StatusSuccess
					}
					if tc.name == "terminal parent" {
						pl.Status = model.StatusSuccess
					}
					if tc.name == "pending" || tc.name == "storage error" {
						w.State = model.StatusPending
					}
					s.On("WorkflowLoad", int64(9)).Return(w, nil)
					if w.PipelineID == pl.ID {
						s.On("GetPipeline", pl.ID).Return(pl, nil)
					}
					if tc.code == 204 || tc.name == "queue error" || tc.name == "storage error" {
						var queueErr error
						if tc.name == "queue error" {
							queueErr = errors.New("queue failed")
						}
						sched.On("CancelWorkflows", mock.Anything, []string{"9"}).Return(queueErr).Once()
						if w.State == model.StatusPending {
							var storeErr error
							if tc.name == "storage error" {
								storeErr = errors.New("database failed")
							}
							s.On("WorkflowCancelPending", int64(9), mock.Anything).Return(storeErr == nil, storeErr)
							if storeErr == nil {
								tree := []*model.Workflow{{ID: 9, State: model.StatusCanceled}, {ID: 10, State: model.StatusRunning}}
								s.On("WorkflowGetTree", mock.Anything).Return(tree, nil)
								s.On("GetUser", int64(2)).Return(user, nil)
								f.On("Status", mock.Anything, user, repo, mock.Anything, tree[0]).Return(nil)
								sched.On("PublishPipelineEvent", mock.Anything, repo, mock.Anything).Return(nil)
							}
						}
					}
				}
			}
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("store", s)
				c.Set("repo", repo)
				c.Set("perm", &model.Perm{Push: !denied})
				if tc.name != "anonymous" {
					c.Set("user", user)
				}
			})
			router.POST("/repos/:repo_id/pipelines/:pipeline_number/workflows/:workflow_id/cancel", session.MustPush, session.SetPipeline(), CancelWorkflow)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/repos/1/pipelines/7/workflows/"+tc.id+"/cancel", nil))
			assert.Equal(t, tc.code, rec.Code)
		})
	}
}
