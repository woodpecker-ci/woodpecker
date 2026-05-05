// Copyright 2025 Woodpecker Authors
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
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker/mocks"
)

func TestRepoShow(t *testing.T) {
	tests := []struct {
		name          string
		repoID        int64
		mockRepo      *woodpecker.Repo
		mockError     error
		expectedError bool
		expected      *woodpecker.Repo
		args          []string
	}{
		{
			name:     "valid repo by ID",
			repoID:   123,
			mockRepo: &woodpecker.Repo{Name: "test-repo"},
			expected: &woodpecker.Repo{Name: "test-repo"},
			args:     []string{"show", "123"},
		},
		{
			name:     "valid repo by full name",
			repoID:   456,
			mockRepo: &woodpecker.Repo{ID: 456, Name: "repo", Owner: "owner"},
			expected: &woodpecker.Repo{ID: 456, Name: "repo", Owner: "owner"},
			args:     []string{"show", "owner/repo"},
		},
		{
			name:          "invalid repo ID",
			repoID:        999,
			expectedError: true,
			args:          []string{"show", "invalid"},
			mockError:     errors.New("repo not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			mockClient.On("Repo", tt.repoID).Return(tt.mockRepo, tt.mockError).Maybe()
			mockClient.On("RepoLookup", "owner/repo").Return(tt.mockRepo, nil).Maybe()

			command := repoShowCmd
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) error {
				output, err := repoShow(c, mockClient)
				if tt.expectedError {
					assert.Error(t, err)
					return nil
				}

				assert.NoError(t, err)
				assert.Equal(t, tt.expected, output)
				return nil
			}

			_ = command.Run(t.Context(), tt.args)
		})
	}
}
