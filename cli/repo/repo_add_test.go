// Copyright 2026 Woodpecker Authors
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

package repo

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker/mocks"
)

func TestRepoAdd(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		setup       func(*mocks.MockClient)
		wantErr     bool
		errContains string
		wantOutput  string
	}{
		{
			name: "activates raw forge remote id",
			arg:  "10",
			setup: func(client *mocks.MockClient) {
				client.On("RepoPost", woodpecker.RepoPostOptions{ForgeRemoteID: "10"}).Return(&woodpecker.Repo{FullName: "owner/repo"}, nil)
			},
			wantOutput: "Successfully activated repository owner/repo\n",
		},
		{
			name: "activates raw string forge remote id",
			arg:  "{workspace-repo}",
			setup: func(client *mocks.MockClient) {
				client.On("RepoPost", woodpecker.RepoPostOptions{ForgeRemoteID: "{workspace-repo}"}).Return(&woodpecker.Repo{FullName: "owner/repo"}, nil)
			},
			wantOutput: "Successfully activated repository owner/repo\n",
		},
		{
			name: "activates repo full name",
			arg:  "owner/repo",
			setup: func(client *mocks.MockClient) {
				client.On("RepoList", woodpecker.RepoListOptions{All: true, Name: "repo"}).Return([]*woodpecker.Repo{
					{FullName: "other/repo", ForgeRemoteID: "20"},
					{FullName: "owner/repo", ForgeRemoteID: "10"},
				}, nil)
				client.On("RepoPost", woodpecker.RepoPostOptions{ForgeRemoteID: "10"}).Return(&woodpecker.Repo{FullName: "owner/repo"}, nil)
			},
			wantOutput: "Successfully activated repository owner/repo\n",
		},
		{
			name: "activates nested repo full name",
			arg:  "group/subgroup/repo",
			setup: func(client *mocks.MockClient) {
				client.On("RepoList", woodpecker.RepoListOptions{All: true, Name: "repo"}).Return([]*woodpecker.Repo{
					{FullName: "group/subgroup/repo", ForgeRemoteID: "nested-id"},
				}, nil)
				client.On("RepoPost", woodpecker.RepoPostOptions{ForgeRemoteID: "nested-id"}).Return(&woodpecker.Repo{FullName: "group/subgroup/repo"}, nil)
			},
			wantOutput: "Successfully activated repository group/subgroup/repo\n",
		},
		{
			name:        "requires argument",
			arg:         "",
			setup:       func(*mocks.MockClient) {},
			wantErr:     true,
			errContains: "repository or forge remote id required",
		},
		{
			name: "returns lookup error",
			arg:  "owner/repo",
			setup: func(client *mocks.MockClient) {
				client.On("RepoList", woodpecker.RepoListOptions{All: true, Name: "repo"}).Return(nil, errors.New("boom"))
			},
			wantErr:     true,
			errContains: "lookup repository \"owner/repo\"",
		},
		{
			name: "returns not found error",
			arg:  "owner/repo",
			setup: func(client *mocks.MockClient) {
				client.On("RepoList", woodpecker.RepoListOptions{All: true, Name: "repo"}).Return([]*woodpecker.Repo{
					{FullName: "other/repo", ForgeRemoteID: "20"},
				}, nil)
			},
			wantErr:     true,
			errContains: "repository \"owner/repo\" not found",
		},
		{
			name: "requires forge remote id in lookup result",
			arg:  "owner/repo",
			setup: func(client *mocks.MockClient) {
				client.On("RepoList", woodpecker.RepoListOptions{All: true, Name: "repo"}).Return([]*woodpecker.Repo{
					{FullName: "owner/repo"},
				}, nil)
			},
			wantErr:     true,
			errContains: "repository \"owner/repo\" has no forge remote id",
		},
		{
			name: "returns post error",
			arg:  "10",
			setup: func(client *mocks.MockClient) {
				client.On("RepoPost", woodpecker.RepoPostOptions{ForgeRemoteID: "10"}).Return(nil, errors.New("boom"))
			},
			wantErr:     true,
			errContains: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockClient(t)
			tt.setup(client)

			var out bytes.Buffer
			err := repoAddWithClient(tt.arg, client, &out)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, out.String())
		})
	}
}

func TestRepoAddCommandActivatesRepoFullName(t *testing.T) {
	requests := make([]string, 0, 2)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI()))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		switch r.URL.RequestURI() {
		case "/api/user/repos?all=true&name=repo":
			assert.Equal(t, http.MethodGet, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprint(w, `[{"full_name":"owner/repo","forge_remote_id":"10"}]`)
			assert.NoError(t, err)
		case "/api/repos?forge_remote_id=10":
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprint(w, `{"id":1,"full_name":"owner/repo","forge_remote_id":"10"}`)
			assert.NoError(t, err)
		default:
			assert.Failf(t, "unexpected request", "%s %s", r.Method, r.URL.RequestURI())
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	var out bytes.Buffer
	command := &cli.Command{
		Name:   "add",
		Flags:  common.GlobalFlags,
		Action: repoAdd,
		Writer: &out,
	}

	err := command.Run(t.Context(), []string{"add", "--server", ts.URL, "--token", "test-token", "owner/repo"})

	assert.NoError(t, err)
	assert.Equal(t, "Successfully activated repository owner/repo\n", out.String())
	assert.Equal(t, []string{
		"GET /api/user/repos?all=true&name=repo",
		"POST /api/repos?forge_remote_id=10",
	}, requests)
}
