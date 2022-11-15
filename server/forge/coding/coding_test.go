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

package coding

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"github.com/woodpecker-ci/woodpecker/server/forge/coding/fixtures"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func Test_coding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	c := &Coding{URL: s.URL}

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("Coding", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a forge", func() {
			g.It("Should return client with specified options", func() {
				forge, _ := New(Opts{
					URL:        "https://coding.net",
					Client:     "KTNF2ALdm3ofbtxLh6IbV95Ro5AKWJUP",
					Secret:     "zVtxJrKhNhBcNyqCz1NggNAAmehAxnRO3Z0fXmCp",
					Scopes:     []string{"user", "project", "project:depot"},
					Username:   "someuser",
					Password:   "password",
					SkipVerify: true,
				})
				g.Assert(forge.(*Coding).URL).Equal("https://coding.net")
				g.Assert(forge.(*Coding).Client).Equal("KTNF2ALdm3ofbtxLh6IbV95Ro5AKWJUP")
				g.Assert(forge.(*Coding).Secret).Equal("zVtxJrKhNhBcNyqCz1NggNAAmehAxnRO3Z0fXmCp")
				g.Assert(forge.(*Coding).Scopes).Equal([]string{"user", "project", "project:depot"})
				g.Assert(forge.(*Coding).Username).Equal("someuser")
				g.Assert(forge.(*Coding).Password).Equal("password")
				g.Assert(forge.(*Coding).SkipVerify).Equal(true)
			})
		})

		g.Describe("Given an authorization request", func() {
			g.It("Should redirect to authorize", func() {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "", nil)
				_, err := c.Login(ctx, w, r)
				g.Assert(err).IsNil()
				g.Assert(w.Code).Equal(http.StatusSeeOther)
			})
			g.It("Should return authenticated user", func() {
				r, _ := http.NewRequest("GET", "?code=code", nil)
				u, err := c.Login(ctx, nil, r)
				g.Assert(err).IsNil()
				g.Assert(u.Login).Equal(fakeUser.Login)
				g.Assert(u.Token).Equal(fakeUser.Token)
				g.Assert(u.Secret).Equal(fakeUser.Secret)
			})
		})

		g.Describe("Given an access token", func() {
			g.It("Should return the authenticated user", func() {
				login, err := c.Auth(ctx, fakeUser.Token, fakeUser.Secret)
				g.Assert(err).IsNil()
				g.Assert(login).Equal(fakeUser.Login)
			})
			g.It("Should handle a failure to resolve user", func() {
				_, err := c.Auth(ctx, fakeUserNotFound.Token, fakeUserNotFound.Secret)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("Given a refresh token", func() {
			g.It("Should return a refresh access token", func() {
				ok, err := c.Refresh(ctx, fakeUserRefresh)
				g.Assert(err).IsNil()
				g.Assert(ok).IsTrue()
				g.Assert(fakeUserRefresh.Token).Equal("VDZupx0usVRV4oOd1FCu4xUxgk8SY0TK")
				g.Assert(fakeUserRefresh.Secret).Equal("BenBQq7TWZ7Cp0aUM47nQjTz2QHNmTWcPctB609n")
			})
			g.It("Should handle an invalid refresh token", func() {
				ok, _ := c.Refresh(ctx, fakeUserRefreshInvalid)
				g.Assert(ok).IsFalse()
			})
		})

		g.Describe("When requesting a repository", func() {
			g.It("Should return the details", func() {
				repo, err := c.Repo(ctx, fakeUser, "", fakeRepo.Owner, fakeRepo.Name)
				g.Assert(err).IsNil()
				g.Assert(repo.FullName).Equal(fakeRepo.FullName)
				g.Assert(repo.Avatar).Equal(s.URL + fakeRepo.Avatar)
				g.Assert(repo.Link).Equal(s.URL + fakeRepo.Link)
				g.Assert(repo.SCMKind).Equal(fakeRepo.SCMKind)
				g.Assert(repo.Clone).Equal(fakeRepo.Clone)
				g.Assert(repo.Branch).Equal(fakeRepo.Branch)
				g.Assert(repo.IsSCMPrivate).Equal(fakeRepo.IsSCMPrivate)
			})
			g.It("Should handle not found errors", func() {
				_, err := c.Repo(ctx, fakeUser, "", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When requesting repository permissions", func() {
			g.It("Should authorize admin access for project owner", func() {
				perm, err := c.Perm(ctx, fakeUser, &model.Repo{Owner: "demo1", Name: "perm_owner"})
				g.Assert(err).IsNil()
				g.Assert(perm.Pull).IsTrue()
				g.Assert(perm.Push).IsTrue()
				g.Assert(perm.Admin).IsTrue()
			})
			g.It("Should authorize admin access for project admin", func() {
				perm, err := c.Perm(ctx, fakeUser, &model.Repo{Owner: "demo1", Name: "perm_admin"})
				g.Assert(err).IsNil()
				g.Assert(perm.Pull).IsTrue()
				g.Assert(perm.Push).IsTrue()
				g.Assert(perm.Admin).IsTrue()
			})
			g.It("Should authorize read access for project member", func() {
				perm, err := c.Perm(ctx, fakeUser, &model.Repo{Owner: "demo1", Name: "perm_member"})
				g.Assert(err).IsNil()
				g.Assert(perm.Pull).IsTrue()
				g.Assert(perm.Push).IsTrue()
				g.Assert(perm.Admin).IsFalse()
			})
			g.It("Should authorize no access for project guest", func() {
				perm, err := c.Perm(ctx, fakeUser, &model.Repo{Owner: "demo1", Name: "perm_guest"})
				g.Assert(err).IsNil()
				g.Assert(perm.Pull).IsFalse()
				g.Assert(perm.Push).IsFalse()
				g.Assert(perm.Admin).IsFalse()
			})
			g.It("Should handle not found errors", func() {
				_, err := c.Perm(ctx, fakeUser, fakeRepoNotFound)
				g.Assert(err).IsNotNil()
			})
		})

		g.Describe("When downloading a file", func() {
			g.It("Should return file for specified pipeline", func() {
				data, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, ".woodpecker.yml")
				g.Assert(err).IsNil()
				g.Assert(string(data)).Equal("pipeline:\n  test:\n    image: golang:1.6\n    commands:\n      - go test\n")
			})
		})

		g.Describe("When requesting a netrc config", func() {
			g.It("Should return the netrc file for global credential", func() {
				forge, _ := New(Opts{
					Username: "someuser",
					Password: "password",
				})
				netrc, err := forge.Netrc(fakeUser, fakeRepo)
				g.Assert(err).IsNil()
				g.Assert(netrc.Login).Equal("someuser")
				g.Assert(netrc.Password).Equal("password")
				g.Assert(netrc.Machine).Equal("git.coding.net")
			})
			g.It("Should return the netrc file for specified user", func() {
				forge, _ := New(Opts{})
				netrc, err := forge.Netrc(fakeUser, fakeRepo)
				g.Assert(err).IsNil()
				g.Assert(netrc.Login).Equal(fakeUser.Token)
				g.Assert(netrc.Password).Equal("x-oauth-basic")
				g.Assert(netrc.Machine).Equal("git.coding.net")
			})
		})

		g.Describe("When activating a repository", func() {
			g.It("Should create the hook", func() {
				err := c.Activate(ctx, fakeUser, fakeRepo, "http://127.0.0.1")
				g.Assert(err).IsNil()
			})
			g.It("Should update the hook when exists", func() {
				err := c.Activate(ctx, fakeUser, fakeRepo, "http://127.0.0.2")
				g.Assert(err).IsNil()
			})
		})

		g.Describe("When deactivating a repository", func() {
			g.It("Should successfully remove hook", func() {
				err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://127.0.0.3")
				g.Assert(err).IsNil()
			})
			g.It("Should successfully deactivate when hook already removed", func() {
				err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://127.0.0.4")
				g.Assert(err).IsNil()
			})
		})

		g.Describe("When parsing post-commit hook body", func() {
			g.It("Should parse the hook", func() {
				buf := bytes.NewBufferString(fixtures.PushHook)
				req, _ := http.NewRequest("POST", "/hook", buf)
				req.Header = http.Header{}
				req.Header.Set(hookEvent, hookPush)

				r, _, err := c.Hook(ctx, req)
				g.Assert(err).IsNil()
				g.Assert(r.FullName).Equal("demo1/test1")
			})
		})
	})
}

var (
	fakeUser = &model.User{
		Login:  "demo1",
		Token:  "KTNF2ALdm3ofbtxLh6IbV95Ro5AKWJUP",
		Secret: "zVtxJrKhNhBcNyqCz1NggNAAmehAxnRO3Z0fXmCp",
	}

	fakeUserNotFound = &model.User{
		Login:  "demo1",
		Token:  "8DpqlE0hI6yr5MLlq8ysAL4p72cKGwT0",
		Secret: "8Em2dkFE8Xsze88Ar8LMG7TF4CO3VCQMgpKa0VCm",
	}

	fakeUserRefresh = &model.User{
		Login:  "demo1",
		Secret: "i9i0HQqNR8bTY4rALYEF2itayFJNbnzC1eMFppwT",
	}

	fakeUserRefreshInvalid = &model.User{
		Login:  "demo1",
		Secret: "invalid_refresh_token",
	}

	fakeRepo = &model.Repo{
		Owner:        "demo1",
		Name:         "test1",
		FullName:     "demo1/test1",
		Avatar:       "/static/project_icon/scenery-5.png",
		Link:         "/u/gilala/p/abp/git",
		SCMKind:      model.RepoGit,
		Clone:        "https://git.coding.net/demo1/test1.git",
		Branch:       "master",
		IsSCMPrivate: true,
	}

	fakeRepoNotFound = &model.Repo{
		Owner: "not_found_owner",
		Name:  "not_found_project",
	}

	fakePipeline = &model.Pipeline{
		Commit: "4504a072cc",
	}
)
