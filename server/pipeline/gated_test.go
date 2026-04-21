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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestSetGatedState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		repo          *model.Repo
		pipeline      *model.Pipeline
		expectBlocked bool
	}{
		{
			name: "by-pass for cron",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventCron,
			},
			expectBlocked: false,
		},
		{
			name: "by-pass for manual pipeline",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventManual,
			},
			expectBlocked: false,
		},
		{
			name: "require approval for fork PRs",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalForks,
			},
			pipeline: &model.Pipeline{
				Event:       model.EventPull,
				PullRequest: &model.PullRequest{FromFork: true},
			},
			expectBlocked: true,
		},
		{
			name: "require approval for PRs",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalPullRequests,
			},
			pipeline: &model.Pipeline{
				Event:       model.EventPull,
				PullRequest: &model.PullRequest{FromFork: false},
			},
			expectBlocked: true,
		},
		{
			name: "require approval for edited PRs",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalPullRequests,
			},
			pipeline: &model.Pipeline{
				Event:       model.EventPullMetadata,
				PullRequest: &model.PullRequest{FromFork: false},
			},
			expectBlocked: true,
		},
		{
			name: "require approval for everything",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventPush,
			},
			expectBlocked: true,
		},
		{
			name: "require approval for everything with allowed user",
			repo: &model.Repo{
				RequireApproval:      model.RequireApprovalAllEvents,
				ApprovalAllowedUsers: []string{"user"},
			},
			pipeline: &model.Pipeline{
				Event:  model.EventPush,
				Author: "user",
			},
			expectBlocked: false,
		},
	}

	for _, tc := range testCases {
		setApprovalState(tc.repo, tc.pipeline)
		assert.Equal(t, tc.expectBlocked, tc.pipeline.Status == model.StatusBlocked)
	}
}
