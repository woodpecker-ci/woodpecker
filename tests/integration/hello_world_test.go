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

//go:build test
// +build test

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v2/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestHelloWorldPipeline(t *testing.T) {
	mockForge := forge_mocks.NewForge(t)
	ctx := context.Background()

	// Set up the mock forge expectations
	repo := &model.Repo{
		ID:     1,
		UserID: 1,
		Owner:  "test-owner",
		Name:   "test-repo",
		Config: ".woodpecker.yml",
	}
	user := &model.User{ID: 1, Login: "test-user"}

	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Status: model.StatusPending,
	}

	config := `
pipeline:
  hello:
    image: alpine
    commands:
      - echo "Hello, World!"
`

	mockForge.On("Repo", ctx, user, model.ForgeRemoteID(""), repo.Owner, repo.Name).Return(repo, nil)
	mockForge.On("File", ctx, user, repo, pipeline, ".woodpecker.yml").Return([]byte(config), nil)
	mockForge.On("Hook", ctx, mock.Anything).Return(repo, pipeline, nil)

	// Use the fake forge with our mock
	WithForge(t, mockForge, func() {
		/* TODOs:
		- login as user
		- activate an repo
		- emulate webhook of forge
		- check if pipeline was created
		- check of pipeline run result
		- cleanup
		*/
	})

	// Assert that all expectations were met
	mockForge.AssertExpectations(t)
}
