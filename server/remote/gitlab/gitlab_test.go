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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitlab/testdata"
)

func load(config string) *Gitlab {
	url_, err := url.Parse(config)
	if err != nil {
		panic(err)
	}
	params := url_.Query()
	url_.RawQuery = ""

	gitlab := Gitlab{}
	gitlab.URL = url_.String()
	gitlab.ClientID = params.Get("client_id")
	gitlab.ClientSecret = params.Get("client_secret")
	gitlab.SkipVerify, _ = strconv.ParseBool(params.Get("skip_verify"))
	gitlab.HideArchives, _ = strconv.ParseBool(params.Get("hide_archives"))

	// this is a temp workaround
	gitlab.Search, _ = strconv.ParseBool(params.Get("search"))

	return &gitlab
}

func Test_Gitlab(t *testing.T) {
	// setup a dummy github server
	var server = testdata.NewServer(t)
	defer server.Close()

	env := server.URL + "?client_id=test&client_secret=test"

	client := load(env)

	var user = model.User{
		Login: "test_user",
		Token: "e3b0c44298fc1c149afbf4c8996fb",
	}

	var repo = model.Repo{
		Name:  "diaspora-client",
		Owner: "diaspora",
	}

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("Gitlab Plugin", func() {
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
				_repo, err := client.Repo(ctx, &user, "diaspora", "diaspora-client")
				assert.NoError(t, err)
				assert.Equal(t, "diaspora-client", _repo.Name)
				assert.Equal(t, "diaspora", _repo.Owner)
				assert.True(t, _repo.IsPrivate)
			})

			g.It("Should return error, when repo not exist", func() {
				_, err := client.Repo(ctx, &user, "not-existed", "not-existed")
				assert.Error(t, err)
			})
		})

		// Test permissions method
		g.Describe("Perm", func() {
			g.It("Should return repo permissions", func() {
				perm, err := client.Perm(ctx, &user, "diaspora", "diaspora-client")
				assert.NoError(t, err)
				assert.True(t, perm.Admin)
				assert.True(t, perm.Pull)
				assert.True(t, perm.Push)
			})
			g.It("Should return repo permissions when user is admin", func() {
				perm, err := client.Perm(ctx, &user, "brightbox", "puppet")
				assert.NoError(t, err)
				g.Assert(perm.Admin).Equal(true)
				g.Assert(perm.Pull).Equal(true)
				g.Assert(perm.Push).Equal(true)
			})
			g.It("Should return error, when repo is not exist", func() {
				_, err := client.Perm(ctx, &user, "not-existed", "not-existed")

				g.Assert(err).IsNotNil()
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

					hookRepo, build, err := client.Hook(req)
					assert.NoError(t, err)
					if assert.NotNil(t, hookRepo) && assert.NotNil(t, build) {
						assert.Equal(t, build.Event, model.EventPush)
						assert.Equal(t, "test", hookRepo.Owner)
						assert.Equal(t, "woodpecker", hookRepo.Name)
						assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
						assert.Equal(t, "develop", hookRepo.Branch)
						assert.Equal(t, "refs/heads/master", build.Ref)
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

					hookRepo, build, err := client.Hook(req)
					assert.NoError(t, err)
					if assert.NotNil(t, hookRepo) && assert.NotNil(t, build) {
						assert.Equal(t, "test", hookRepo.Owner)
						assert.Equal(t, "woodpecker", hookRepo.Name)
						assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
						assert.Equal(t, "develop", hookRepo.Branch)
						assert.Equal(t, "refs/tags/v22", build.Ref)
					}
				})
			})

			g.Describe("Merge request hook", func() {
				g.It("Should parse merge request hook", func() {
					req, _ := http.NewRequest(
						testdata.ServiceHookMethod,
						testdata.ServiceHookURL.String(),
						bytes.NewReader(testdata.ServiceHookMergeRequestBody),
					)
					req.Header = testdata.ServiceHookHeaders

					hookRepo, build, err := client.Hook(req)
					assert.NoError(t, err)
					if assert.NotNil(t, hookRepo) && assert.NotNil(t, build) {
						assert.Equal(t, "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg", hookRepo.Avatar)
						assert.Equal(t, "develop", hookRepo.Branch)
						assert.Equal(t, "test", hookRepo.Owner)
						assert.Equal(t, "woodpecker", hookRepo.Name)
						assert.Equal(t, "Update client.go 🎉", build.Title)
					}
				})
			})
		})
	})
}
