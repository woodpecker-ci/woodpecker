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
