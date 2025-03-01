// Copyright 2024 Woodpecker Authors
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

package forgejo

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/forgejo/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(Opts{
		URL:        "http://localhost:8080",
		SkipVerify: true,
	})

	f, _ := forge.(*Forgejo)
	assert.Equal(t, "http://localhost:8080", f.url)
	assert.True(t, f.SkipVerify)
}

func Test_forgejo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()
	c, _ := New(Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	mockStore := mocks_store.NewStore(t)
	ctx := store.InjectToContext(context.Background(), mockStore)

	t.Run("netrc with user token", func(t *testing.T) {
		forge, _ := New(Opts{})
		netrc, _ := forge.Netrc(fakeUser, fakeRepo)
		assert.Equal(t, "forgejo.org", netrc.Machine)
		assert.Equal(t, fakeUser.Login, netrc.Login)
		assert.Equal(t, fakeUser.AccessToken, netrc.Password)
		assert.Equal(t, model.ForgeTypeForgejo, netrc.Type)
	})
	t.Run("netrc with machine account", func(t *testing.T) {
		forge, _ := New(Opts{})
		netrc, _ := forge.Netrc(nil, fakeRepo)
		assert.Equal(t, "forgejo.org", netrc.Machine)
		assert.Empty(t, netrc.Login)
		assert.Empty(t, netrc.Password)
	})

	t.Run("repository details", func(t *testing.T) {
		repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.Owner, repo.Owner)
		assert.Equal(t, fakeRepo.Name, repo.Name)
		assert.Equal(t, fakeRepo.Owner+"/"+fakeRepo.Name, repo.FullName)
		assert.True(t, repo.IsSCMPrivate)
		assert.Equal(t, "http://localhost/test_name/repo_name.git", repo.Clone)
		assert.Equal(t, "http://localhost/test_name/repo_name", repo.ForgeURL)
	})
	t.Run("repo not found", func(t *testing.T) {
		_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
		assert.Error(t, err)
	})

	t.Run("repository list", func(t *testing.T) {
		repos, err := c.Repos(ctx, fakeUser)
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.ForgeRemoteID, repos[0].ForgeRemoteID)
		assert.Equal(t, fakeRepo.Owner, repos[0].Owner)
		assert.Equal(t, fakeRepo.Name, repos[0].Name)
		assert.Equal(t, fakeRepo.Owner+"/"+fakeRepo.Name, repos[0].FullName)
	})
	t.Run("not found error", func(t *testing.T) {
		_, err := c.Repos(ctx, fakeUserNoRepos)
		assert.Error(t, err)
	})

	t.Run("register repository", func(t *testing.T) {
		err := c.Activate(ctx, fakeUser, fakeRepo, "http://localhost")
		assert.NoError(t, err)
	})

	t.Run("remove hooks", func(t *testing.T) {
		err := c.Deactivate(ctx, fakeUser, fakeRepo, "http://localhost")
		assert.NoError(t, err)
	})

	t.Run("repository file", func(t *testing.T) {
		raw, err := c.File(ctx, fakeUser, fakeRepo, fakePipeline, ".woodpecker.yml")
		assert.NoError(t, err)
		assert.Equal(t, "{ platform: linux/amd64 }", string(raw))
	})

	t.Run("pipeline status", func(t *testing.T) {
		err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
		assert.NoError(t, err)
	})

	t.Run("PR hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookPullRequest)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookPullRequest)
		mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything).Return(fakeRepo, nil)
		mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
		r, b, err := c.Hook(ctx, req)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NoError(t, err)
		assert.Equal(t, model.EventPull, b.Event)
		assert.Equal(t, []string{"README.md"}, b.ChangedFiles)
	})
}

var (
	fakeUser = &model.User{
		Login:       "someuser",
		AccessToken: "cfcd2084",
	}

	fakeUserNoRepos = &model.User{
		Login:       "someuser",
		AccessToken: "repos_not_found",
	}

	fakeRepo = &model.Repo{
		Clone:         "http://forgejo.org/test_name/repo_name.git",
		ForgeRemoteID: "5",
		Owner:         "test_name",
		Name:          "repo_name",
		FullName:      "test_name/repo_name",
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}

	fakePipeline = &model.Pipeline{
		Commit: &model.Commit{SHA: "9ecad50"},
	}

	fakeWorkflow = &model.Workflow{
		Name:  "test",
		State: model.StatusSuccess,
	}
)
