// Copyright 2024 Woodpecker Authors
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

package pipeline

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker/mocks"
)

func TestPipelineList(t *testing.T) {
	testtases := []struct {
		name        string
		repoID      int64
		repoErr     error
		pipelines   []*woodpecker.Pipeline
		pipelineErr error
		args        []string
		expected    []*woodpecker.Pipeline
		wantErr     error
	}{
		{
			name:   "success",
			repoID: 1,
			pipelines: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
			args: []string{"ls", "repo/name"},
			expected: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
		},
		{
			name:   "limit results",
			repoID: 1,
			pipelines: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
			args: []string{"ls", "--limit", "2", "repo/name"},
			expected: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
			},
		},
		{
			name:        "pipeline list error",
			repoID:      1,
			pipelineErr: errors.New("pipeline error"),
			args:        []string{"ls", "repo/name"},
			wantErr:     errors.New("pipeline error"),
		},
	}

	for _, tt := range testtases {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			mockClient.On("PipelineList", mock.Anything, mock.Anything).Return(func(_ int64, opt woodpecker.PipelineListOptions) ([]*woodpecker.Pipeline, error) {
				if tt.pipelineErr != nil {
					return nil, tt.pipelineErr
				}
				if opt.Page == 1 {
					return tt.pipelines, nil
				}
				return []*woodpecker.Pipeline{}, nil
			}).Maybe()
			mockClient.On("RepoLookup", mock.Anything).Return(&woodpecker.Repo{ID: tt.repoID}, nil)

			command := buildPipelineListCmd()
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) error {
				pipelines, err := pipelineList(c, mockClient)
				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
					return nil
				}

				assert.NoError(t, err)
				assert.EqualValues(t, tt.expected, pipelines)

				return nil
			}

			_ = command.Run(t.Context(), tt.args)
		})
	}
}
