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
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitlab/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
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

func Test_GitLab(t *testing.T) {
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
				assert.Equal(t, "http://10.40.8.5:3200/test/woodpecker/-/tags/v22", pipeline.ForgeURL)
				assert.Len(t, pipeline.ChangedFiles, 0)
				assert.Equal(t, model.EventTag, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
			}
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
				assert.Equal(t, "Update client.go 🎉", pipeline.PullRequest.Title)
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
				assert.Equal(t, "Edit README.md for more text to read", pipeline.PullRequest.Title)
				assert.Len(t, pipeline.ChangedFiles, 0) // see L217
				assert.Equal(t, model.EventPull, pipeline.Event)
				assert.Empty(t, pipeline.EventReason)
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
				assert.Equal(t, "Add new file", pipeline.PullRequest.Title)
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
				assert.Equal(t, "Some ned more AAAA", pipeline.PullRequest.Title)
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
				assert.Equal(t, "Add new file", pipeline.PullRequest.Title)
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
				assert.Equal(t, "Edit README for more text to read", pipeline.PullRequest.Title)
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
				assert.Equal(t, "Edit README for more text to read", pipeline.PullRequest.Title)
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, name, err := extractFromPath(tc.input)

			// Check error expectations
			if tc.errContains != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantOwner, owner)
			assert.EqualValues(t, tc.wantName, name)
		})
	}
}
