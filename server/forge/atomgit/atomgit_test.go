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

package atomgit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/atomgit/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestNew(t *testing.T) {
	forge, _ := New(1, Opts{
		URL:        "http://localhost:8080",
		SkipVerify: true,
	})

	f, _ := forge.(*AtomGit)
	assert.Equal(t, "http://localhost:8080", f.url)
	assert.True(t, f.skipVerify)
	assert.Equal(t, "atomgit", f.Name())
}

// Test_user_unmarshal verifies AtomGit's /api/v5/user payload decodes
// correctly. AtomGit returns the id as an opaque hex string (e.g.
// "6638af02bbeee41d0fe74c35"), never a number, and the login field is
// "login" rather than "username". This guards the login regression where the
// id could not be unmarshaled.
func Test_user_unmarshal(t *testing.T) {
	var u user
	// AtomGit returns id as a hex string and the login field as "login".
	err := json.Unmarshal([]byte(`{"id":"6638af02bbeee41d0fe74c35","login":"someuser","name":"Some User","email":"a@b.com"}`), &u)
	assert.NoError(t, err)
	assert.Equal(t, "6638af02bbeee41d0fe74c35", u.ID.String())
	assert.Equal(t, "someuser", u.Username)

	w := toUser(&u)
	assert.Equal(t, model.ForgeRemoteID("6638af02bbeee41d0fe74c35"), w.ForgeRemoteID)
	assert.Equal(t, "someuser", w.Login)
}

func Test_atomgit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := httptest.NewServer(fixtures.Handler())
	defer s.Close()
	c, _ := New(1, Opts{
		URL:        s.URL,
		SkipVerify: true,
	})

	mockStore := store_mocks.NewMockStore(t)
	ctx := store.InjectToContext(t.Context(), mockStore)

	t.Run("repository details", func(t *testing.T) {
		repo, err := c.Repo(ctx, fakeUser, fakeRepo.ForgeRemoteID, fakeRepo.Owner, fakeRepo.Name)
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.Owner, repo.Owner)
		assert.Equal(t, fakeRepo.Name, repo.Name)
		assert.Equal(t, fakeRepo.Owner+"/"+fakeRepo.Name, repo.FullName)
		assert.Equal(t, "http://localhost/test_name/repo_name.git", repo.Clone)
		assert.Equal(t, "http://localhost/test_name/repo_name", repo.ForgeURL)
	})
	t.Run("repo not found", func(t *testing.T) {
		_, err := c.Repo(ctx, fakeUser, "0", fakeRepoNotFound.Owner, fakeRepoNotFound.Name)
		assert.Error(t, err)
	})

	t.Run("repository list", func(t *testing.T) {
		repos, err := c.Repos(ctx, fakeUser, &model.ListOptions{Page: 1, PerPage: 10})
		assert.NoError(t, err)
		assert.Equal(t, fakeRepo.ForgeRemoteID, repos[0].ForgeRemoteID)
		assert.Equal(t, fakeRepo.Owner, repos[0].Owner)
		assert.Equal(t, fakeRepo.Name, repos[0].Name)
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

	t.Run("pipeline status is no-op", func(t *testing.T) {
		err := c.Status(ctx, fakeUser, fakeRepo, fakePipeline, fakeWorkflow)
		assert.NoError(t, err)
	})

	t.Run("PR hook", func(t *testing.T) {
		buf := bytes.NewBufferString(fixtures.HookMergeRequest)
		req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
		req.Header = http.Header{}
		req.Header.Set(hookEvent, hookMergeRequest)
		mockStore.On("GetRepoNameFallback", mock.Anything, mock.Anything, mock.Anything).Return(fakeRepo, nil)
		mockStore.On("GetUser", mock.Anything).Return(fakeUser, nil)
		r, b, err := c.Hook(ctx, req)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NoError(t, err)
		assert.Equal(t, model.EventPull, b.Event)
	})

	t.Run("netrc", func(t *testing.T) {
		netrc, err := c.Netrc(fakeUser, fakeRepo)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", netrc.Machine)
	assert.Equal(t, fakeUser.Login, netrc.Login)
	assert.Equal(t, model.ForgeTypeAtomGit, netrc.Type)
})
}

// Test_atomgit_repoByIDScan_noNamespace reproduces the bug where the
// /user/repos summary omits the namespace object. Repo() must still upgrade
// the matched summary to the full repository (via GET /repos/:owner/:name) so
// that forge_url (html_url) and pr_enabled (has_pull_requests) are populated.
func Test_atomgit_repoByIDScan_noNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mux := http.NewServeMux()
	// Force the primary GET /repositories/:id path to 404 so Repo() falls
	// back to getRepoByIDScan.
	mux.HandleFunc("/api/v5/repositories/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found","code":404}`))
	})
	// /user/repos summary WITHOUT a namespace object, but with the fields
	// needed to derive owner/name (path_with_namespace / full_name).
	mux.HandleFunc("/api/v5/user/repos", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{
			"id": 5,
			"name": "repo_name",
			"path_with_namespace": "test_name/repo_name",
			"full_name": "test_name/repo_name",
			"html_url": "http://localhost/test_name/repo_name",
			"has_pull_requests": true
		}]`))
	})
	// Full repo fetch returns html_url + has_pull_requests.
	mux.HandleFunc("/api/v5/repos/test_name/repo_name", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"id": 5,
			"name": "repo_name",
			"path_with_namespace": "test_name/repo_name",
			"full_name": "test_name/repo_name",
			"html_url": "http://localhost/test_name/repo_name",
			"http_url_to_repo": "http://localhost/test_name/repo_name.git",
			"ssh_url_to_repo": "git@localhost:test_name/repo_name.git",
			"default_branch": "master",
			"public": true,
			"has_pull_requests": true
		}`))
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	c, _ := New(1, Opts{URL: s.URL, SkipVerify: true})

	mockStore := store_mocks.NewMockStore(t)
	ctx := store.InjectToContext(t.Context(), mockStore)

	// Call Repo() with only the remoteID (no owner/name) so the by-id scan
	// path is exercised.
	repo, err := c.Repo(ctx, fakeUser, "5", "", "")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/test_name/repo_name", repo.ForgeURL)
	assert.True(t, repo.PREnabled)
}

// Test_atomgit_repo_byIDIncomplete verifies that when GET /repositories/:id
// succeeds but returns an incomplete payload (no html_url), Repo() does NOT
// trust it and instead upgrades to the full repository via GET /repos/:owner/:name
// (or the /user/repos scan), so forge_url is still populated. This is the exact
// bug that produced an empty forge_url -> "<button>" on the repo page.
func Test_atomgit_repo_byIDIncomplete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mux := http.NewServeMux()
	// /repositories/:id returns 200 but WITHOUT html_url / has_pull_requests.
	mux.HandleFunc("/api/v5/repositories/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id": 5, "name": "repo_name", "full_name": "test_name/repo_name"}`))
	})
	// /user/repos scan (used by getRepoByIDScan fallback) returns the summary.
	mux.HandleFunc("/api/v5/user/repos", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id": 5, "path_with_namespace": "test_name/repo_name", "full_name": "test_name/repo_name"}]`))
	})
	// Full repo fetch carries html_url + has_pull_requests.
	mux.HandleFunc("/api/v5/repos/test_name/repo_name", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"id": 5,
			"name": "repo_name",
			"path_with_namespace": "test_name/repo_name",
			"full_name": "test_name/repo_name",
			"html_url": "http://localhost/test_name/repo_name",
			"http_url_to_repo": "http://localhost/test_name/repo_name.git",
			"ssh_url_to_repo": "git@localhost:test_name/repo_name.git",
			"default_branch": "master",
			"public": true,
			"has_pull_requests": true
		}`))
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	c, _ := New(1, Opts{URL: s.URL, SkipVerify: true})

	mockStore := store_mocks.NewMockStore(t)
	ctx := store.InjectToContext(t.Context(), mockStore)

	// Call with a valid remoteID AND owner/name; the by-id path is incomplete
	// and must be upgraded to the full repo.
	repo, err := c.Repo(ctx, fakeUser, "5", "test_name", "repo_name")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/test_name/repo_name", repo.ForgeURL)
	assert.True(t, repo.PREnabled)
}

var (
	fakeUser = &model.User{
		Login:       "someuser",
		AccessToken: "cfcd2084",
	}

	fakeRepo = &model.Repo{
		Clone:         "http://localhost/test_name/repo_name.git",
		ForgeRemoteID: "5",
		Owner:         "test_name",
		Name:          "repo_name",
		FullName:      "test_name/repo_name",
		Hash:          "secret",
	}

	fakeRepoNotFound = &model.Repo{
		Owner:    "test_name",
		Name:     "repo_not_found",
		FullName: "test_name/repo_not_found",
	}

	fakePipeline = &model.Pipeline{
		Commit: "9ecad50",
	}

	fakeWorkflow = &model.Workflow{
		Name:  "test",
		State: model.StatusSuccess,
	}
)

// Test_sshURLFromClone_derivesFromHTTPHost verifies the SSH clone URL is built
// from the HTTP clone URL's host.
func Test_sshURLFromClone_derivesFromHTTPHost(t *testing.T) {
	cases := []struct {
		name  string
		clone string
		owner string
		repo  string
		want  string
	}{
		{"atomgit https", "https://atomgit.com/jetsung/testci.git", "jetsung", "testci", "git@atomgit.com:jetsung/testci.git"},
		{"atomgit http", "http://atomgit.com/jetsung/testci.git", "jetsung", "testci", "git@atomgit.com:jetsung/testci.git"},
		{"same as http host", "https://atomgit.com/jetsung/testci.git", "jetsung", "testci", "git@atomgit.com:jetsung/testci.git"},
		{"empty clone", "", "jetsung", "testci", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := sshURLFromClone(c.clone, c.owner, c.repo)
			assert.Equal(t, c.want, got)
		})
	}
}
