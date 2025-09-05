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

package fixtures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/neticdk/go-bitbucket/mock"
)

func Server() *httptest.Server {
	return mock.NewMockServer(
		mock.WithRequestMatch(mock.SearchRepositories, bitbucket.RepositoryList{
			ListResponse: bitbucket.ListResponse{
				LastPage: true,
			},
			Repositories: []*bitbucket.Repository{
				{
					ID:   uint64(123),
					Slug: "repo-slug-1",
					Name: "REPO Name 1",
					Project: &bitbucket.Project{
						ID:  uint64(456),
						Key: "PRJ",
					},
				},
				{
					ID:   uint64(1234),
					Slug: "repo-slug-2",
					Name: "REPO Name 2",
					Project: &bitbucket.Project{
						ID:  uint64(456),
						Key: "PRJ",
					},
				},
			},
		}),
		mock.WithRequestMatch(mock.GetRepository, bitbucket.Repository{
			ID:   uint64(123),
			Slug: "repo-slug",
			Name: "REPO Name",
			Project: &bitbucket.Project{
				ID:  uint64(456),
				Key: "PRJ",
			},
		}),
		mock.WithRequestMatch(mock.GetDefaultBranch, bitbucket.Branch{
			ID:        "refs/head/main",
			DisplayID: "main",
			Default:   true,
		}),
	)
}

func ServerWithOrgPermissions() *httptest.Server {
	return mock.NewMockServer(
		mock.WithRequestMatchHandler(mock.SearchRepositories, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectKey := r.URL.Query().Get("projectkey")
			permission := r.URL.Query().Get("permission")

			var response bitbucket.RepositoryList

			switch {
			case projectKey == "PRJ-ADMIN" && bitbucket.Permission(permission) == bitbucket.PermissionRepoAdmin:
				response = bitbucket.RepositoryList{
					ListResponse: bitbucket.ListResponse{
						LastPage: true,
					},
					Repositories: []*bitbucket.Repository{
						{
							ID:   uint64(123),
							Slug: "admin-repo",
							Name: "Admin Repo",
							Project: &bitbucket.Project{
								ID:  uint64(456),
								Key: "PRJ-ADMIN",
							},
						},
					},
				}
			case projectKey == "PRJ-WRITE" && bitbucket.Permission(permission) == bitbucket.PermissionRepoAdmin:
				response = bitbucket.RepositoryList{
					ListResponse: bitbucket.ListResponse{
						LastPage: true,
					},
					Repositories: []*bitbucket.Repository{},
				}
			case projectKey == "PRJ-WRITE" && bitbucket.Permission(permission) == bitbucket.PermissionRepoWrite:
				response = bitbucket.RepositoryList{
					ListResponse: bitbucket.ListResponse{
						LastPage: true,
					},
					Repositories: []*bitbucket.Repository{
						{
							ID:   uint64(124),
							Slug: "write-repo",
							Name: "Write Repo",
							Project: &bitbucket.Project{
								ID:  uint64(457),
								Key: "PRJ-WRITE",
							},
						},
					},
				}
			default:
				response = bitbucket.RepositoryList{
					ListResponse: bitbucket.ListResponse{
						LastPage: true,
					},
					Repositories: []*bitbucket.Repository{},
				}
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		})),
	)
}
