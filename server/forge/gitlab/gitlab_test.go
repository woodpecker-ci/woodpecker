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

package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitlab/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func load(config string) *GitLab {
	_url, _ := url.Parse(config)
	params := _url.Query()
	_url.RawQuery = ""

	gitlab := GitLab{}
	gitlab.url = _url.String()
	gitlab.oAuthClientID = params.Get("client_id")
	gitlab.oAuthClientSecret = params.Get("client_secret")
	gitlab.skipVerify, _ = strconv.ParseBool(params.Get("skip_verify"))
	gitlab.hideArchives, _ = strconv.ParseBool(params.Get("hide_archives"))

	// this is a temp workaround
	gitlab.search, _ = strconv.ParseBool(params.Get("search"))

	return &gitlab
}

func TestGitLab(t *testing.T) {
	// setup a dummy gitlab server
	server := fixtures.NewServer(t)
	defer server.Close()

	env := server.URL + "?client_id=test&client_secret=test"

	client := load(env)

	user := model.User{
		Login:         "test_user",
		AccessToken:   "e3b0c44298fc1c149afbf4c8996fb",
		ForgeRemoteID: "3",
	}

	repo := model.Repo{
		Name:  "diaspora-client",
		Owner: "diaspora",
	}

	ctx := t.Context()
	// Test projects method
	t.Run("Should return only non-archived projects is hidden", func(t *testing.T) {
		client.hideArchives = true
		_projects, err := client.Repos(ctx, &user, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		assert.Len(t, _projects, 1)
	})

	t.Run("Should return all the projects", func(t *testing.T) {
		client.hideArchives = false
		_projects, err := client.Repos(ctx, &user, &model.ListOptions{Page: 1, PerPage: 10})

		assert.NoError(t, err)
		assert.Len(t, _projects, 2)
	})

	// Test repository method
	t.Run("Should return valid repo", func(t *testing.T) {
		_repo, err := client.Repo(ctx, &user, "0", "diaspora", "diaspora-client")
		assert.NoError(t, err)
		assert.Equal(t, "diaspora-client", _repo.Name)
		assert.Equal(t, "diaspora", _repo.Owner)
		assert.True(t, _repo.IsSCMPrivate)
	})

	t.Run("Should return error, when repo not exist", func(t *testing.T) {
		_, err := client.Repo(ctx, &user, "0", "not-existed", "not-existed")
		assert.Error(t, err)
	})

	t.Run("Should return repo with push access, when user inherits membership from namespace", func(t *testing.T) {
		_repo, err := client.Repo(ctx, &user, "6", "brightbox", "puppet")
		assert.NoError(t, err)
		assert.True(t, _repo.Perm.Push)
	})

	// Test teams membership method
	t.Run("Should identify groups by full path, not display name", func(t *testing.T) {
		teams, err := client.Teams(ctx, &user, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)

		logins := make([]string, 0, len(teams))
		for _, team := range teams {
			logins = append(logins, team.Login)
		}
		// both groups are named "woodpecker", only the full path tells them apart
		assert.Equal(t, []string{"woodpecker", "eve/woodpecker"}, logins)
	})

	// Test pull requests method
	t.Run("pull request list", func(t *testing.T) {
		prs, err := client.PullRequests(ctx, &user, &repo, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		assert.Len(t, prs, 1)
		// the project scoped IID is used, as in the pipeline ref "refs/merge-requests/<iid>/head"
		assert.Equal(t, model.ForgeRemoteID("7"), prs[0].Index)
		assert.Equal(t, "Add feature X", prs[0].Title)
	})

	// Test activate method
	t.Run("Activate, success", func(t *testing.T) {
		err := client.Activate(ctx, &user, &repo, "http://example.com/api/hook?access_token=token")
		assert.NoError(t, err)
	})

	t.Run("Activate, failed no token", func(t *testing.T) {
		err := client.Activate(ctx, &user, &repo, "http://example.com/api/hook")

		assert.Error(t, err)
	})

	// Test deactivate method
	t.Run("Deactivate", func(t *testing.T) {
		err := client.Deactivate(ctx, &user, &repo, "http://example.com/api/hook?access_token=token")
		assert.NoError(t, err)
	})

	t.Run("test parse webhook", func(t *testing.T) {
		// Test hook method
		t.Run("parse push", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPush),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, pipeline.Event, model.EventPush)
				assert.Equal(t, "test", hookRepo.Owner)
				assert.Equal(t, "woodpecker", hookRepo.Name)
				assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
				assert.Equal(t, "develop", hookRepo.Branch)
				assert.Equal(t, "refs/heads/main", pipeline.Ref)
				assert.Equal(t, []string{"cmd/cli/main.go"}, pipeline.ChangedFiles)
				assert.Equal(t, model.EventPush, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
			}
		})

		t.Run("tag push", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookTag),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "test", hookRepo.Owner)
				assert.Equal(t, "woodpecker", hookRepo.Name)
				assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
				assert.Equal(t, "develop", hookRepo.Branch)
				assert.Equal(t, "refs/tags/v22", pipeline.Ref)
				assert.Equal(t, "v22", pipeline.TagTitle)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventTag, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
			}
		})

		t.Run("tag deleted", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookTagDeleted),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
			assert.Nil(t, hookRepo)
			assert.Nil(t, pipeline)
		})

		t.Run("merge request", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestUpdated),
			)
			req.Header = fixtures.ServiceHookHeaders

			// TODO: insert fake store into context to retrieve user & repo, this will activate fetching of ChangedFiles
			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "anbraten", hookRepo.Owner)
				assert.Equal(t, "woodpecker", hookRepo.Name)
				assert.Equal(t, "Update client.go 🎉", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPull, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
			}
		})

		t.Run("merge request new opened", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestOpened),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			// TODO: insert fake store into context to retrieve user & repo, this will activate fetching of ChangedFiles
			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Equal(t, "Edit README.md for more text to read", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPull, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
				assert.False(t, pipeline.PullRequestDraft)
			}
		})

		t.Run("merge request work in progress", func(t *testing.T) {
			payload := bytes.ReplaceAll(
				fixtures.HookPullRequestOpened,
				[]byte(`"work_in_progress": false`),
				[]byte(`"work_in_progress": true`),
			)
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(payload),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			_, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, pipeline) {
				assert.True(t, pipeline.PullRequestDraft)
			}
		})

		t.Run("ignore merge request hook without changes", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestWithoutChanges),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.Nil(t, hookRepo)
			assert.Nil(t, pipeline)
			if assert.ErrorIs(t, err, &types.ErrIgnoreEvent{}) {
				assert.EqualValues(t,
					"explicit ignored event 'Merge Request Hook', reason: Action 'update' no supported changes detected",
					err.Error())
			}
		})

		t.Run("ignore unsupported action", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestUnsupportedAction),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.Nil(t, hookRepo)
			assert.Nil(t, pipeline)
			if assert.ErrorIs(t, err, &types.ErrIgnoreEvent{}) {
				assert.EqualValues(t,
					"explicit ignored event 'Merge Request Hook', reason: Action 'action_we_do_not_support' not supported",
					err.Error())
			}
		})

		t.Run("parse merge request closed", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestClosed),
			)
			req.Header = fixtures.ServiceHookHeaders

			// TODO: insert fake store into context to retrieve user & repo, this will activate fetching of ChangedFiles
			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "anbraten", hookRepo.Owner)
				assert.Equal(t, "woodpecker-test", hookRepo.Name)
				assert.Equal(t, "Add new file", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPullClosed, pipeline.Event)
			}
		})

		t.Run("merge request reopened", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestReopened),
			)
			req.Header = fixtures.ServiceHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Equal(t, "Some ned more AAAA", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPull, pipeline.Event)
			}
		})

		t.Run("parse merge request merged", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestMerged),
			)
			req.Header = fixtures.ServiceHookHeaders

			// TODO: insert fake store into context to retrieve user & repo, this will activate fetching of ChangedFiles
			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "anbraten", hookRepo.Owner)
				assert.Equal(t, "woodpecker-test", hookRepo.Name)
				assert.Equal(t, "Add new file", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPullClosed, pipeline.Event)
			}
		})

		t.Run("merge request title and description edited", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestEdited),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			// TODO: insert fake store into context to retrieve user & repo, this will activate fetching of ChangedFiles
			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Equal(t, "Edit README for more text to read", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"title_edited", "description_edited"}, pipeline.EventReason)
			}
		})

		t.Run("release", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.WebhookReleaseBody),
			)
			req.Header = fixtures.ReleaseHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "refs/tags/0.0.2", pipeline.Ref)
				assert.Equal(t, "ci", hookRepo.Name)
				assert.Equal(t, "Awesome version 0.0.2", pipeline.Release.Title)
				assert.Equal(t, model.EventRelease, pipeline.Event)
			}
		})

		t.Run("merge request review approved", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestApproved),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
			assert.Nil(t, hookRepo)
			assert.Nil(t, pipeline)
		})

		t.Run("merge request review requested", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestReviewRequested),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Equal(t, "Edit README for more text to read", pipeline.Title)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"review_requested"}, pipeline.EventReason)
			}
		})

		t.Run("merge request assigned", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestAssigned),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"assigned"}, pipeline.EventReason)
			}
		})

		t.Run("merge request unassigned", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestUnassigned),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"assigned", "unassigned"}, pipeline.EventReason)
			}
		})

		t.Run("merge request milestoned", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestMilestoned),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"milestoned"}, pipeline.EventReason)
			}
		})

		t.Run("merge request demilestoned", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestDemilestoned),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"demilestoned"}, pipeline.EventReason)
			}
		})

		t.Run("merge request labels added", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestLabelsAdded),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"label_added"}, pipeline.EventReason)
			}
		})

		t.Run("merge request labels cleared", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestLabelsCleared),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"label_cleared"}, pipeline.EventReason)
			}
		})

		t.Run("merge request labels updated", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestLabelsUpdated),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.NoError(t, err)
			if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
				assert.Equal(t, "main", hookRepo.Branch)
				assert.Equal(t, "demoaccount2-commits-group", hookRepo.Owner)
				assert.Equal(t, "test_ci_tmp", hookRepo.Name)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventPullMetadata, pipeline.Event)
				assert.Equal(t, []string{"label_updated"}, pipeline.EventReason)
			}
		})

		t.Run("merge request unapproved", func(t *testing.T) {
			req, _ := http.NewRequest(
				fixtures.ServiceHookMethod,
				fixtures.ServiceHookURL.String(),
				bytes.NewReader(fixtures.HookPullRequestUnapproved),
			)
			req.Header = fixtures.MergeRequestHookHeaders

			hookRepo, pipeline, err := client.Hook(ctx, req)
			assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
			assert.Nil(t, hookRepo)
			assert.Nil(t, pipeline)
		})
	})
}

func TestExtractFromPath(t *testing.T) {
	type testCase struct {
		name        string
		input       string
		wantOwner   string
		wantName    string
		errContains string
	}

	tests := []testCase{
		{
			name:      "basic two components",
			input:     "owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "three components",
			input:     "owner/group/repo",
			wantOwner: "owner/group",
			wantName:  "repo",
		},
		{
			name:      "many components",
			input:     "owner/group/subgroup/deep/repo",
			wantOwner: "owner/group/subgroup/deep",
			wantName:  "repo",
		},
		{
			name:        "empty string",
			input:       "",
			errContains: "minimum match not found",
		},
		{
			name:        "single component",
			input:       "onlyrepo",
			errContains: "minimum match not found",
		},
		{
			name:      "trailing slash",
			input:     "owner/repo/",
			wantOwner: "owner/repo",
			wantName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := extractFromPath(tt.input)

			// Check error expectations
			if tt.errContains != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, tt.wantOwner, owner)
			assert.EqualValues(t, tt.wantName, name)
		})
	}
}

func TestGitLabReposUsesEmbeddedPermissions(t *testing.T) {
	var memberCalls atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, _ *http.Request) {
		// project 4 reports the current user's permissions inline,
		// project 6 does not (e.g. access inherited from a nested parent group)
		_, _ = w.Write([]byte(`[` +
			`{"id":4,"path_with_namespace":"diaspora/diaspora-client","visibility":"private","permissions":` +
			`{"project_access":{"access_level":10},"group_access":{"access_level":50}}},` +
			`{"id":6,"path_with_namespace":"brightbox/puppet","visibility":"private","permissions":null}` +
			`]`))
	})
	mux.HandleFunc("/api/v4/projects/4/members/all/3", func(w http.ResponseWriter, _ *http.Request) {
		memberCalls.Add(1)
		http.Error(w, "unexpected membership lookup for project with embedded permissions", http.StatusInternalServerError)
	})
	mux.HandleFunc("/api/v4/projects/6/members/all/3", func(w http.ResponseWriter, _ *http.Request) {
		memberCalls.Add(1)
		_, _ = w.Write([]byte(`{"id":3,"username":"test_user","access_level":30}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := load(server.URL + "?client_id=test&client_secret=test")
	user := model.User{Login: "test_user", AccessToken: "token", ForgeRemoteID: "3"}

	repos, err := client.Repos(t.Context(), &user, &model.ListOptions{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, repos, 2)

	// only the project without embedded permissions should trigger a
	// membership lookup, i.e. no N+1 requests
	assert.Equal(t, int64(1), memberCalls.Load())

	// project 4: owner via group access -> read, write and admin
	assert.True(t, repos[0].Perm.Pull)
	assert.True(t, repos[0].Perm.Push)
	assert.True(t, repos[0].Perm.Admin)

	// project 6: permissions missing -> fallback lookup reports developer -> read and write, but no admin
	assert.True(t, repos[1].Perm.Pull)
	assert.True(t, repos[1].Perm.Push)
	assert.False(t, repos[1].Perm.Admin)
}

func TestGitLabReposResolvesGroupInheritedProjectOnMembershipLookupNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[` +
			`{"id":4,"path_with_namespace":"diaspora/diaspora-client","visibility":"private","permissions":` +
			`{"project_access":{"access_level":40},"group_access":null}},` +
			`{"id":7,"path_with_namespace":"other/personal-project","visibility":"private","permissions":null}` +
			`]`))
	})
	mux.HandleFunc("/api/v4/projects/7/members/all/3", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	})
	// project 7 has no direct membership, but the full-project endpoint
	// reports the access inherited from a parent group (Developer).
	mux.HandleFunc("/api/v4/projects/7", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":7,"path_with_namespace":"other/personal-project","visibility":"private","permissions":` +
			`{"project_access":null,"group_access":{"access_level":30}}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := load(server.URL + "?client_id=test&client_secret=test")
	user := model.User{Login: "test_user", AccessToken: "token", ForgeRemoteID: "3"}

	repos, err := client.Repos(t.Context(), &user, &model.ListOptions{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, repos, 2)

	assert.Equal(t, "diaspora/diaspora-client", repos[0].FullName)
	assert.Equal(t, "other/personal-project", repos[1].FullName)
	// project 7: group-inherited Developer -> read and write, no admin
	assert.True(t, repos[1].Perm.Pull)
	assert.True(t, repos[1].Perm.Push)
	assert.False(t, repos[1].Perm.Admin)
}

func TestGitLabReposKeepsProjectWhenFullProjectLookupFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[` +
			`{"id":4,"path_with_namespace":"diaspora/diaspora-client","visibility":"private","permissions":` +
			`{"project_access":{"access_level":40},"group_access":null}},` +
			`{"id":7,"path_with_namespace":"other/personal-project","visibility":"private","permissions":null}` +
			`]`))
	})
	mux.HandleFunc("/api/v4/projects/7/members/all/3", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	})
	// no handler for /api/v4/projects/7 -> GetProject 404s; the project must
	// still be returned (with zero access) instead of failing the whole page.

	server := httptest.NewServer(mux)
	defer server.Close()

	client := load(server.URL + "?client_id=test&client_secret=test")
	user := model.User{Login: "test_user", AccessToken: "token", ForgeRemoteID: "3"}

	repos, err := client.Repos(t.Context(), &user, &model.ListOptions{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, repos, 2)

	assert.Equal(t, "other/personal-project", repos[1].FullName)
	assert.False(t, repos[1].Perm.Push)
}

func TestGitLabStatusIgnoresRejectedTransition(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		body    string
		wantErr bool
	}{
		{
			name: "transition from running",
			code: http.StatusBadRequest,
			body: `{"message":"Cannot transition status via :enqueue from :running (Reason(s): Status cannot transition via \"enqueue\")"}`,
		},
		{
			name: "transition from pending",
			code: http.StatusBadRequest,
			body: `{"message":"Cannot transition status via :enqueue from :pending (Reason(s): Status cannot transition via \"enqueue\")"}`,
		},
		{
			name:    "other bad request",
			code:    http.StatusBadRequest,
			body:    `{"message":"state is missing"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v4/projects/4", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"id":4,"path_with_namespace":"diaspora/diaspora-client"}`))
			})
			mux.HandleFunc("/api/v4/projects/4/statuses/abc123", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.code)
				_, _ = w.Write([]byte(tt.body))
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			client := load(server.URL + "?client_id=test&client_secret=test")
			user := &model.User{Login: "test_user", AccessToken: "token"}
			repo := &model.Repo{ForgeRemoteID: "4", Owner: "diaspora", Name: "diaspora-client", FullName: "diaspora/diaspora-client"}
			pipeline := &model.Pipeline{Number: 1, Commit: "abc123", Event: model.EventPush}
			workflow := &model.Workflow{Name: "build", State: model.StatusPending}

			err := client.Status(t.Context(), user, repo, pipeline, workflow)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newTestServer(t *testing.T, routes map[string]string) *GitLab {
	t.Helper()
	mux := http.NewServeMux()
	for route, body := range routes {
		mux.HandleFunc(route, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		})
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return load(server.URL + "?client_id=test&client_secret=test")
}

// storeContext returns a context with a store that resolves every hook repo
// to project 4 of a user with a valid token.
func storeContext(t *testing.T) context.Context {
	t.Helper()
	s := store_mocks.NewMockStore(t)
	s.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(&model.Repo{
		ID:            1,
		UserID:        1,
		ForgeRemoteID: "4",
		Owner:         "diaspora",
		Name:          "diaspora-client",
	}, nil)
	s.On("GetUser", int64(1)).Return(&model.User{
		ID:          1,
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour).Unix(),
	}, nil)
	return store.InjectToContext(t.Context(), s)
}

func hookRequest(t *testing.T, body []byte, header http.Header) *http.Request {
	t.Helper()
	req, err := http.NewRequest(fixtures.ServiceHookMethod, fixtures.ServiceHookURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header = header
	return req
}

func TestHookLoadsMetadataFromForge(t *testing.T) {
	t.Run("merge request gets changed files and milestone", func(t *testing.T) {
		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4": `{"id":4}`,
			// a renamed file reports both paths, an edited one the same path twice
			"/api/v4/projects/4/merge_requests/3/diffs": `[` +
				`{"old_path":"old.go","new_path":"new.go"},` +
				`{"old_path":"main.go","new_path":"main.go"}]`,
			"/api/v4/projects/4/milestones/6088906": `{"id":6088906,"title":"v1.0"}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, fixtures.HookPullRequestMilestoned, fixtures.ServiceHookHeaders))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"new.go", "old.go", "main.go"}, pipeline.ChangedFiles)
		assert.Equal(t, "v1.0", pipeline.PullRequestMilestone)
	})

	t.Run("tag without message loads the tagged commit", func(t *testing.T) {
		var payload map[string]any
		require.NoError(t, json.Unmarshal(fixtures.HookTag, &payload))
		payload["message"] = ""
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		sha, ok := payload["after"].(string)
		require.True(t, ok)

		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4": `{"id":4}`,
			"/api/v4/projects/4/repository/commits/" + sha: `{"id":"` + sha + `","message":"Release v22",` +
				`"author_name":"Jane","author_email":"jane@example.com","committed_date":"2026-01-02T03:04:05Z"}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, body, fixtures.ServiceHookHeaders))
		require.NoError(t, err)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "Release v22", pipeline.Message)
		// the author stays the authenticated webhook username, not the git
		// commit author name (CVE-2026-58370 approval-bypass fix)
		assert.Equal(t, "test", pipeline.Author)
		assert.Equal(t, "jane@example.com", pipeline.Email)
		assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix(), pipeline.Timestamp)
		assert.NotEmpty(t, pipeline.Avatar)
	})

	t.Run("annotated tag records the tagged commit, not the tag object", func(t *testing.T) {
		// an annotated tag's `after` is the tag object's SHA; the commits
		// endpoint resolves it to the tagged commit, whose ID differs
		tagObjectSHA := "187b97cf74a1c2d3e4f5a6b7c8d9e0f1a2b3c4d5"
		commitSHA := "afd225e8b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5"

		var payload map[string]any
		require.NoError(t, json.Unmarshal(fixtures.HookTag, &payload))
		payload["message"] = ""
		payload["after"] = tagObjectSHA
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4": `{"id":4}`,
			"/api/v4/projects/4/repository/commits/" + tagObjectSHA: `{"id":"` + commitSHA + `","message":"Release v22",` +
				`"author_name":"Jane","author_email":"jane@example.com","committed_date":"2026-01-02T03:04:05Z"}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, body, fixtures.ServiceHookHeaders))
		require.NoError(t, err)
		assert.Equal(t, model.EventTag, pipeline.Event)
		assert.Equal(t, "refs/tags/v22", pipeline.Ref)
		assert.Equal(t, commitSHA, pipeline.Commit)
		assert.Equal(t, "Release v22", pipeline.Message)
	})

	t.Run("release gets the release author", func(t *testing.T) {
		client := newTestServer(t, map[string]string{
			"/api/v4/projects/4":                `{"id":4}`,
			"/api/v4/projects/4/releases/0.0.2": `{"tag_name":"0.0.2","author":{"username":"releaser","avatar_url":"https://avatar/releaser"}}`,
		})

		_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, fixtures.WebhookReleaseBody, fixtures.ReleaseHookHeaders))
		require.NoError(t, err)
		assert.Equal(t, model.EventRelease, pipeline.Event)
		assert.Equal(t, "releaser", pipeline.Author)
		assert.Equal(t, "https://avatar/releaser", pipeline.Avatar)
	})

	t.Run("forge errors fail the hook", func(t *testing.T) {
		client := newTestServer(t, map[string]string{"/api/v4/projects/4": `{"id":4}`})

		_, _, err := client.Hook(storeContext(t), hookRequest(t, fixtures.HookPullRequestMilestoned, fixtures.ServiceHookHeaders))
		assert.Error(t, err)
		_, _, err = client.Hook(storeContext(t), hookRequest(t, fixtures.WebhookReleaseBody, fixtures.ReleaseHookHeaders))
		assert.Error(t, err)
	})
}

func TestOrgMembership(t *testing.T) {
	user := &model.User{Login: "test_user", AccessToken: "token"}

	// the first page is full, so the member on the second page must be found
	members := func(page int) string {
		if page == 1 {
			list := make([]string, defaultPerPage)
			for i := range list {
				list[i] = fmt.Sprintf(`{"id":%d,"username":"user%d","access_level":30}`, i+100, i)
			}
			return "[" + strings.Join(list, ",") + "]"
		}
		return `[{"id":3,"username":"test_user","access_level":50}]`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, _ *http.Request) {
		// the search also matches groups whose path only contains the name
		_, _ = w.Write([]byte(`[{"id":1,"path":"group-other","full_path":"group-other"},{"id":2,"path":"group","full_path":"group"}]`))
	})
	mux.HandleFunc("/api/v4/groups/2/members/all", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if r.URL.Query().Get("page") == "2" {
			page = 2
		}
		_, _ = w.Write([]byte(members(page)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")

	t.Run("owner on a later member page", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), user, "group")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true, Admin: true}, perm)
	})

	t.Run("not a member", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), &model.User{Login: "stranger", AccessToken: "token"}, "group")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{}, perm)
	})

	t.Run("no group with that path", func(t *testing.T) {
		perm, err := client.OrgMembership(t.Context(), user, "grou")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{}, perm)
	})
}

func TestOrg(t *testing.T) {
	groups := map[string]string{
		// display names differ from the full path, only the full path counts
		"parent/sub": `[{"id":1,"path":"sub","name":"sub","full_path":"other/sub","visibility":"public"},` +
			`{"id":2,"path":"sub","name":"sub","full_path":"parent/sub","visibility":"internal"}]`,
		"public": `[{"id":3,"path":"public","full_path":"public","visibility":"public"}]`,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/users", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("username") == "jane" {
			_, _ = w.Write([]byte(`[{"id":5,"username":"jane","private_profile":true}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, r *http.Request) {
		body, ok := groups[r.URL.Query().Get("search")]
		if !ok {
			body = `[]`
		}
		_, _ = w.Write([]byte(body))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")
	user := &model.User{AccessToken: "token"}

	org, err := client.Org(t.Context(), user, "jane")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "jane", IsUser: true, Private: true}, org)

	org, err = client.Org(t.Context(), user, "parent/sub")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "parent/sub", Private: true}, org)

	org, err = client.Org(t.Context(), user, "public")
	require.NoError(t, err)
	assert.Equal(t, &model.Org{Name: "public", Private: false}, org)

	_, err = client.Org(t.Context(), user, "missing")
	assert.ErrorContains(t, err, "could not find org missing")
}

func TestNetrc(t *testing.T) {
	client := load("https://gitlab.example.com?client_id=test&client_secret=test")
	repo := &model.Repo{Clone: "https://gitlab.example.com/group/repo.git"}

	netrc, err := client.Netrc(&model.User{AccessToken: "token"}, repo)
	require.NoError(t, err)
	assert.Equal(t, &model.Netrc{Machine: "gitlab.example.com", Login: "oauth2", Password: "token", Type: model.ForgeTypeGitlab}, netrc)

	// without a user only the machine is known
	netrc, err = client.Netrc(nil, repo)
	require.NoError(t, err)
	assert.Equal(t, &model.Netrc{Machine: "gitlab.example.com", Type: model.ForgeTypeGitlab}, netrc)
}

func TestFileAndDir(t *testing.T) {
	user := &model.User{AccessToken: "token"}
	repo := &model.Repo{ForgeRemoteID: "4"}
	pipeline := &model.Pipeline{Commit: "abc"}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/4", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":4}`))
	})
	mux.HandleFunc("/api/v4/projects/4/repository/files/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "abc", r.URL.Query().Get("ref"))
		switch strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v4/projects/4/repository/files/"), "/raw") {
		case ".woodpecker/a.yaml":
			_, _ = w.Write([]byte("steps: a"))
		case ".woodpecker/b.yaml":
			_, _ = w.Write([]byte("steps: b"))
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/api/v4/projects/4/repository/tree", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, ".woodpecker", r.URL.Query().Get("path"))
		_, _ = w.Write([]byte(`[` +
			`{"type":"blob","path":".woodpecker/a.yaml"},` +
			`{"type":"tree","path":".woodpecker/sub"},` +
			`{"type":"blob","path":".woodpecker/b.yaml"}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := load(server.URL + "?client_id=test&client_secret=test")

	data, err := client.File(t.Context(), user, repo, pipeline, ".woodpecker/a.yaml")
	require.NoError(t, err)
	assert.Equal(t, "steps: a", string(data))

	_, err = client.File(t.Context(), user, repo, pipeline, ".woodpecker.yaml")
	assert.ErrorIs(t, err, &types.ErrConfigNotFound{})

	// sub directories are skipped, files are loaded with their content
	metas, err := client.Dir(t.Context(), user, repo, pipeline, ".woodpecker")
	require.NoError(t, err)
	assert.Equal(t, []*types.FileMeta{
		{Name: ".woodpecker/a.yaml", Data: []byte("steps: a")},
		{Name: ".woodpecker/b.yaml", Data: []byte("steps: b")},
	}, metas)
}

func TestDirPagination(t *testing.T) {
	user := &model.User{AccessToken: "token"}
	repo := &model.Repo{ForgeRemoteID: "4"}
	pipeline := &model.Pipeline{Commit: "abc"}

	// fullPage returns a page with defaultPerPage sub folders and the given blob,
	// so the client has to ask for the next page.
	fullPage := func(blob string) string {
		entries := make([]string, 0, defaultPerPage)
		for i := range defaultPerPage - 1 {
			entries = append(entries, fmt.Sprintf(`{"type":"tree","path":".woodpecker/sub%d"}`, i))
		}
		entries = append(entries, `{"type":"blob","path":"`+blob+`"}`)
		return "[" + strings.Join(entries, ",") + "]"
	}

	newServer := func(t *testing.T, tree http.HandlerFunc) *GitLab {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v4/projects/4", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"id":4}`))
		})
		mux.HandleFunc("/api/v4/projects/4/repository/files/", func(w http.ResponseWriter, r *http.Request) {
			name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v4/projects/4/repository/files/"), "/raw")
			_, _ = w.Write([]byte("steps: " + name))
		})
		mux.HandleFunc("/api/v4/projects/4/repository/tree", tree)
		server := httptest.NewServer(mux)
		t.Cleanup(server.Close)
		return load(server.URL + "?client_id=test&client_secret=test")
	}

	t.Run("follows pages until a short page", func(t *testing.T) {
		var requests []string
		client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			if slices.Contains(requests, page) {
				http.Error(w, "page "+page+" requested twice", http.StatusBadRequest)
				return
			}
			requests = append(requests, page)
			switch page {
			case "1":
				_, _ = w.Write([]byte(fullPage(".woodpecker/a.yaml")))
			case "2":
				_, _ = w.Write([]byte(`[{"type":"blob","path":".woodpecker/b.yaml"}]`))
			default:
				http.Error(w, "unexpected page "+page, http.StatusBadRequest)
			}
		})

		metas, err := client.Dir(t.Context(), user, repo, pipeline, ".woodpecker")
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "2"}, requests)
		assert.Equal(t, []*types.FileMeta{
			{Name: ".woodpecker/a.yaml", Data: []byte("steps: .woodpecker/a.yaml")},
			{Name: ".woodpecker/b.yaml", Data: []byte("steps: .woodpecker/b.yaml")},
		}, metas)
	})
}

func TestBranchHead(t *testing.T) {
	client := newTestServer(t, map[string]string{
		"/api/v4/projects/4":                          `{"id":4}`,
		"/api/v4/projects/4/repository/branches/main": `{"name":"main","commit":{"id":"abc","web_url":"https://gitlab.example.com/commit/abc"}}`,
	})

	commit, err := client.BranchHead(t.Context(), &model.User{AccessToken: "token"}, &model.Repo{ForgeRemoteID: "4"}, "main")
	require.NoError(t, err)
	assert.Equal(t, &model.Commit{SHA: "abc", ForgeURL: "https://gitlab.example.com/commit/abc"}, commit)
}

func TestRefreshExchangesStillValidToken(t *testing.T) {
	var tokenRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		tokenRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"new-access","token_type":"bearer","expires_in":3600,"refresh_token":"new-refresh"}`)
	}))
	defer server.Close()

	forge, err := New(1, Opts{URL: server.URL, OAuthClientID: "id", OAuthClientSecret: "secret"})
	assert.NoError(t, err)

	user := &model.User{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		Expiry:       time.Now().Add(5 * time.Minute).UTC().Unix(),
	}

	f, ok := forge.(*GitLab)
	require.True(t, ok)

	updated, err := f.Refresh(t.Context(), user)
	assert.NoError(t, err)
	assert.True(t, updated)

	assert.Equal(t, int32(1), tokenRequests.Load(), "token endpoint must be called exactly once")
	assert.Equal(t, "new-access", user.AccessToken)
	assert.Equal(t, "new-refresh", user.RefreshToken)
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(user.Expiry, 0), 2*time.Minute)
}

// TestTagAuthorNotSpoofedFromCommit is a regression for the incomplete fix of
// CVE-2026-58370: a GitLab tag event whose commit message is not in the webhook
// payload falls back to loadCommitFromSHA, which must not overwrite
// pipeline.Author (the authenticated webhook username) with the git commit
// author name, which a pusher controls and could set to an
// ApprovalAllowedUsers entry to skip approval.
func TestTagAuthorNotSpoofedFromCommit(t *testing.T) {
	var payload map[string]any
	require.NoError(t, json.Unmarshal(fixtures.HookTag, &payload))
	payload["message"] = ""
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	sha, ok := payload["after"].(string)
	require.True(t, ok)

	client := newTestServer(t, map[string]string{
		"/api/v4/projects/4": `{"id":4}`,
		"/api/v4/projects/4/repository/commits/" + sha: `{"id":"` + sha + `","message":"Release v22",` +
			`"author_name":"trusted-admin","author_email":"x@example.com","committed_date":"2026-01-02T03:04:05Z"}`,
	})

	_, pipeline, err := client.Hook(storeContext(t), hookRequest(t, body, fixtures.ServiceHookHeaders))
	require.NoError(t, err)
	assert.Equal(t, model.EventTag, pipeline.Event)
	// "test" is user_username in the fixture; "trusted-admin" is the spoofable git author name.
	assert.Equal(t, "test", pipeline.Author)
}
