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
