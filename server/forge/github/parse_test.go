// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package github

import (
	"bytes"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookEvent   = "X-GitHub-Event"
	hookDeploy  = "deployment"
	hookPush    = "push"
	hookPull    = "pull_request"
	hookRelease = "release"
)

func testHookRequest(payload []byte, event string) *http.Request {
	buf := bytes.NewBuffer(payload)
	req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
	req.Header = http.Header{}
	req.Header.Set(hookEvent, event)
	return req
}

func Test_parseHook(t *testing.T) {
	t.Run("ignore unsupported hook events", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequest), "issues")
		p, r, b, err := parseHook(req, false)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("skip skip push hook when action is deleted", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPushDeleted), hookPush)
		p, r, b, err := parseHook(req, false)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.NoError(t, err)
		assert.Nil(t, p)
	})
	t.Run("push hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPush), hookPush)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.Nil(t, p)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPush, b.Event)
		sort.Strings(b.ChangedFiles)
		assert.Equal(t, []string{"pipeline/shared/replace_secrets.go", "pipeline/shared/replace_secrets_test.go"}, b.ChangedFiles)
	})

	t.Run("PR hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequest), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPull, b.Event)
	})
	t.Run("PR closed hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestClosed), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPullClosed, b.Event)
	})

	t.Run("reopen a pull", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestReopened), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPull, b.Event)
	})

	t.Run("PR merged hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestMerged), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPullClosed, b.Event)
	})

	t.Run("PR edited hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestEdited), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "edited", b.EventReason)
	})

	t.Run("deploy hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookDeploy), hookDeploy)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Nil(t, p)
		assert.Equal(t, model.EventDeploy, b.Event)
		assert.Equal(t, "production", b.DeployTo)
		assert.Equal(t, "deploy", b.DeployTask)
	})

	t.Run("release hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookRelease), hookRelease)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Nil(t, p)
		assert.Equal(t, model.EventRelease, b.Event)
		assert.Len(t, strings.Split(b.Ref, "/"), 3)
		assert.True(t, strings.HasPrefix(b.Ref, "refs/tags/"))
	})

	t.Run("pull review requested", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestReviewRequested), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "review_requested", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			assert.Equal(t, "yeaaa", *p.Body)
			assert.Equal(t, false, *p.Draft)
			assert.Equal(t, false, *p.Merged)
			assert.Equal(t, true, *p.Mergeable)
			assert.Equal(t, "unstable", *p.MergeableState)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.Len(t, p.RequestedReviewers, 1) {
				assert.Equal(t, "demoaccount2-commits", *p.RequestedReviewers[0].Login)
				assert.Equal(t, int64(223550959), *p.RequestedReviewers[0].ID)
			}
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	t.Run("pull milestoned", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestMilestoneAdded), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "milestoned", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			assert.Equal(t, "yeaaa", *p.Body)
			assert.Equal(t, false, *p.Draft)
			assert.Equal(t, false, *p.Merged)
			assert.Equal(t, true, *p.Mergeable)
			assert.Equal(t, "unstable", *p.MergeableState)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.NotNil(t, p.Milestone) {
				assert.Equal(t, int64(13392101), *p.Milestone.ID)
				assert.Equal(t, 2, *p.Milestone.Number)
				assert.Equal(t, "open mile", *p.Milestone.Title)
				assert.Equal(t, "ongoing", *p.Milestone.Description)
				assert.Equal(t, "open", *p.Milestone.State)
				if assert.NotNil(t, p.Milestone.Creator) {
					assert.Equal(t, "demoaccount2-commits", *p.Milestone.Creator.Login)
					assert.Equal(t, int64(223550959), *p.Milestone.Creator.ID)
				}
			}
			assert.Empty(t, p.RequestedReviewers)
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	// milestone change will result two webhooks an demilestoned and milestoned

	t.Run("pull request demilestoned", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestMilestoneRemoved), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "demilestoned", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.Len(t, p.Labels, 1) {
				assert.Equal(t, int64(9024465370), *p.Labels[0].ID)
				assert.Equal(t, "bug", *p.Labels[0].Name)
				assert.Equal(t, "d73a4a", *p.Labels[0].Color)
				assert.Equal(t, "Something isn't working", *p.Labels[0].Description)
			}
			assert.Nil(t, p.Milestone)
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	t.Run("pull request labeled", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestLabelAdded), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "labeled", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			assert.Equal(t, "yeaaa", *p.Body)
			assert.Equal(t, false, *p.Draft)
			assert.Equal(t, false, *p.Merged)
			assert.Equal(t, true, *p.Mergeable)
			assert.Equal(t, "unstable", *p.MergeableState)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.Len(t, p.Labels, 2) {
				assert.Equal(t, int64(9024465376), *p.Labels[0].ID)
				assert.Equal(t, "documentation", *p.Labels[0].Name)
				assert.Equal(t, "0075ca", *p.Labels[0].Color)
				assert.Equal(t, "Improvements or additions to documentation", *p.Labels[0].Description)
				assert.Equal(t, int64(9024465382), *p.Labels[1].ID)
				assert.Equal(t, "enhancement", *p.Labels[1].Name)
				assert.Equal(t, "a2eeef", *p.Labels[1].Color)
				assert.Equal(t, "New feature or request", *p.Labels[1].Description)
			}
			if assert.NotNil(t, p.Milestone) {
				assert.Equal(t, int64(13392101), *p.Milestone.ID)
				assert.Equal(t, "open mile", *p.Milestone.Title)
			}
			assert.Empty(t, p.RequestedReviewers)
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	// lable change will result two webhooks an unlable and labelled

	t.Run("pull request unlabeled", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestLabelRemoved), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "unlabeled", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.Len(t, p.Labels, 1) {
				assert.Equal(t, int64(9024465370), *p.Labels[0].ID)
				assert.Equal(t, "bug", *p.Labels[0].Name)
				assert.Equal(t, "d73a4a", *p.Labels[0].Color)
				assert.Equal(t, "Something isn't working", *p.Labels[0].Description)
			}
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	t.Run("pull request assigned", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestAssigneeAdded), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "assigned", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			if assert.NotNil(t, p.Assignee) {
				assert.Equal(t, "demoaccount2-commits", *p.Assignee.Login)
				assert.Equal(t, int64(223550959), *p.Assignee.ID)
			}
			if assert.Len(t, p.Assignees, 1) {
				assert.Equal(t, "demoaccount2-commits", *p.Assignees[0].Login)
				assert.Equal(t, int64(223550959), *p.Assignees[0].ID)
			}
			if assert.Len(t, p.Labels, 1) {
				assert.Equal(t, int64(9024465370), *p.Labels[0].ID)
				assert.Equal(t, "bug", *p.Labels[0].Name)
			}
			assert.Nil(t, p.Milestone)
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})

	// assigne change will result two webhooks an assigned and unassigned

	t.Run("pull request unassigned", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestAssigneeRemoved), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPullMetadata, b.Event)
		assert.Equal(t, "unassigned", b.EventReason)
		if assert.NotNil(t, p) {
			assert.Equal(t, int64(2705176047), *p.ID)
			assert.Equal(t, 1, *p.Number)
			assert.Equal(t, "open", *p.State)
			assert.Equal(t, "Some ned more AAAA", *p.Title)
			if assert.NotNil(t, p.User) {
				assert.Equal(t, "6543", *p.User.Login)
				assert.Equal(t, int64(24977596), *p.User.ID)
			}
			assert.Nil(t, p.Assignee)
			assert.Empty(t, p.Assignees)
			if assert.Len(t, p.Labels, 1) {
				assert.Equal(t, int64(9024465370), *p.Labels[0].ID)
				assert.Equal(t, "bug", *p.Labels[0].Name)
			}
			assert.Nil(t, p.Milestone)
			if assert.NotNil(t, p.Head) {
				assert.Equal(t, "6543-patch-1", *p.Head.Ref)
				assert.Equal(t, "36b5813240a9d2daa29b05046d56a53e18f39a3e", *p.Head.SHA)
			}
			if assert.NotNil(t, p.Base) {
				assert.Equal(t, "main", *p.Base.Ref)
				assert.Equal(t, "67012991d6c69b1c58378346fca366b864d8d1a1", *p.Base.SHA)
			}
		}
	})
}
