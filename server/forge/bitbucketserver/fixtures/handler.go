package fixtures

import (
	"net/http/httptest"

	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/neticdk/go-bitbucket/mock"
)

func Server() *httptest.Server {
	return mock.NewMockServer(
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
