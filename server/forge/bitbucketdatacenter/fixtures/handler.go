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
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/neticdk/go-bitbucket/mock"
	"github.com/stretchr/testify/assert"
)

var (
	//go:embed expected/*
	embeddedFixtures embed.FS
	PostBuildStatus  = mock.EndpointPattern{Pattern: "/api/latest/projects/:projectKey/repos/:repositorySlug/commits/:commitId/builds", Method: "POST"}
)

type ResponseContent map[string]any

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

		mock.WithRequestMatchHandler(PostBuildStatus, ExpectedContentHandler(
			"PostBuildStatus.json",
			http.StatusNoContent, nil,
			http.StatusBadRequest, ResponseContent{
				"errors": []ResponseContent{
					{
						"context":       "",
						"exceptionName": "",
						"message":       "invalid branch was provided",
					},
				},
			},
		)),
	)
}

func ExpectedContentHandler(expectedFileName string, successCode int, successContent ResponseContent, failCode int, failContent ResponseContent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expectedContent, err := loadExpectedContent(expectedFileName)
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, ResponseContent{"error": "Internal Server Error"})
			return
		}

		var requestBody ResponseContent
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			writeResponse(w, failCode, failContent)
			return
		}

		if !assert.ObjectsAreEqual(requestBody, expectedContent) {
			writeResponse(w, failCode, failContent)
			return
		}

		writeResponse(w, successCode, successContent)
	}
}

func loadExpectedContent(fileName string) (ResponseContent, error) {
	file, err := embeddedFixtures.Open(filepath.Join("expected", fileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var content ResponseContent
	err = json.NewDecoder(file).Decode(&content)
	return content, err
}

func writeResponse(w http.ResponseWriter, statusCode int, content ResponseContent) {
	w.WriteHeader(statusCode)
	if content != nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(content); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
