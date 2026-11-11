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

package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// A status the forge refuses for one workflow must not keep the other
// workflows of the pipeline from getting theirs.
func TestUpdatePipelineStatusSendsEveryWorkflow(t *testing.T) {
	repo := &model.Repo{ID: 1, FullName: "octocat/hello-world"}
	user := &model.User{ID: 1}
	failing := &model.Workflow{ID: 1, State: model.StatusSuccess}
	next := &model.Workflow{ID: 2, State: model.StatusFailure}
	pipeline := &model.Pipeline{ID: 1, Workflows: []*model.Workflow{failing, next}}

	forge := forge_mocks.NewMockForge(t)
	forge.EXPECT().Status(mock.Anything, user, repo, pipeline, failing).Return(errors.New("forge unavailable")).Once()
	forge.EXPECT().Status(mock.Anything, user, repo, pipeline, next).Return(nil).Once()

	updatePipelineStatus(t.Context(), forge, pipeline, repo, user)
}
