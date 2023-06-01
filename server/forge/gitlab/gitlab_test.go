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
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server/forge/gitlab/testdata"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func load(t *testing.T, config string) *GitLab {
	_url, err := url.Parse(config)
	if err != nil {
		t.FailNow()
	}
	params := _url.Query()
	_url.RawQuery = ""

	gitlab := GitLab{}
	gitlab.url = _url.String()
	gitlab.ClientID = params.Get("client_id")
	gitlab.ClientSecret = params.Get("client_secret")
	gitlab.SkipVerify, _ = strconv.ParseBool(params.Get("skip_verify"))
	gitlab.HideArchives, _ = strconv.ParseBool(params.Get("hide_archives"))

	// this is a temp workaround
	gitlab.Search, _ = strconv.ParseBool(params.Get("search"))

	return &gitlab
}

func Test_GitLab(t *testing.T) {
	// setup a dummy gitlab server
	server := testdata.NewServer(t)
	defer server.Close()

	env := server.URL + "?client_id=test&client_secret=test"

	client := load(t, env)

	user := model.User{
		Login: "test_user",
		Token: "e3b0c44298fc1c149afbf4c8996fb",
	}

	repo := model.Repo{
		Name:  "diaspora-client",
		Owner: "diaspora",
	}

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("GitLab Plugin", func() {
		// Test projects method
		g.Describe("AllProjects", func() {
			g.It("Should return only non-archived projects is hidden", func() {
				client.HideArchives = true
				_projects, err := client.Repos(ctx, &user)
				assert.NoError(t, err)
				assert.Len(t, _projects, 1)
			})

			g.It("Should return all the projects", func() {
				client.HideArchives = false
				_projects, err := client.Repos(ctx, &user)

				g.Assert(err).IsNil()
				g.Assert(len(_projects)).Equal(2)
			})
		})

		// Test repository method
		g.Describe("Repo", func() {
			g.It("Should return valid repo", func() {
				_repo, err := client.Repo(ctx, &user, "0", "diaspora", "diaspora-client")
				assert.NoError(t, err)
				assert.Equal(t, "diaspora-client", _repo.Name)
				assert.Equal(t, "diaspora", _repo.Owner)
				assert.True(t, _repo.IsSCMPrivate)
			})

			g.It("Should return error, when repo not exist", func() {
				_, err := client.Repo(ctx, &user, "0", "not-existed", "not-existed")
				assert.Error(t, err)
			})
		})

		// Test activate method
		g.Describe("Activate", func() {
			g.It("Should be success", func() {
				err := client.Activate(ctx, &user, &repo, "http://example.com/api/hook/test/test?access_token=token")
				assert.NoError(t, err)
			})

			g.It("Should be failed, when token not given", func() {
				err := client.Activate(ctx, &user, &repo, "http://example.com/api/hook/test/test")

				g.Assert(err).IsNotNil()
			})
		})

		// Test deactivate method
		g.Describe("Deactivate", func() {
			g.It("Should be success", func() {
				err := client.Deactivate(ctx, &user, &repo, "http://example.com/api/hook/test/test?access_token=token")

				g.Assert(err).IsNil()
			})
		})

		// Test hook method
		g.Describe("Hook", func() {
			g.Describe("Push hook", func() {
				g.It("Should parse actual push hook", func() {
					req, _ := http.NewRequest(
						testdata.ServiceHookMethod,
						testdata.ServiceHookURL.String(),
						bytes.NewReader(testdata.ServiceHookPushBody),
					)
					req.Header = testdata.ServiceHookHeaders

					hookRepo, pipeline, err := client.Hook(ctx, req)
					assert.NoError(t, err)
					if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
						assert.Equal(t, pipeline.Event, model.EventPush)
						assert.Equal(t, "test", hookRepo.Owner)
						assert.Equal(t, "woodpecker", hookRepo.Name)
						assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
						assert.Equal(t, "develop", hookRepo.Branch)
						assert.Equal(t, "refs/heads/master", pipeline.Ref)
						assert.Equal(t, []string{"cmd/cli/main.go"}, pipeline.ChangedFiles)
					}
				})
			})

			g.Describe("Tag push hook", func() {
				g.It("Should parse tag push hook", func() {
					req, _ := http.NewRequest(
						testdata.ServiceHookMethod,
						testdata.ServiceHookURL.String(),
						bytes.NewReader(testdata.ServiceHookTagPushBody),
					)
					req.Header = testdata.ServiceHookHeaders

					hookRepo, pipeline, err := client.Hook(ctx, req)
					assert.NoError(t, err)
					if assert.NotNil(t, hookRepo) && assert.NotNil(t, pipeline) {
						assert.Equal(t, "test", hookRepo.Owner)
						assert.Equal(t, "woodpecker", hookRepo.Name)
						assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
						assert.Equal(t, "develop", hookRepo.Branch)
						assert.Equal(t, "refs/tags/v22", pipeline.Ref)
						assert.Len(t, pipeline.ChangedFiles, 0)
					}
				})
			})

			g.Describe("Merge request hook", func() {
				g.It("Should parse merge request hook", func() {
					req, _ := http.NewRequest(
						testdata.ServiceHookMethod,
						testdata.ServiceHookURL.String(),
						bytes.NewReader(testdata.WebhookMergeRequestBody),
					)
					req.Header = testdata.ServiceHookHeaders

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
					}
				})
			})
		})
	})
}
