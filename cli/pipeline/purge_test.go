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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker/mocks"
)

func TestPipelinePurge(t *testing.T) {
	tests := []struct {
		name            string
		repoID          int64
		args            []string
		pipelinesKeep   []*woodpecker.Pipeline
		pipelines       []*woodpecker.Pipeline
		mockDeleteError error
		wantDelete      int
		wantErr         error
	}{
		{
			name:   "success with no pipelines to purge",
			repoID: 1,
			args:   []string{"purge", "--older-than", "1h", "repo/name"},
			pipelinesKeep: []*woodpecker.Pipeline{
				{Number: 1},
			},
			pipelines: []*woodpecker.Pipeline{},
		},
		{
			name:   "success with pipelines to purge",
			repoID: 1,
			args:   []string{"purge", "--older-than", "1h", "repo/name"},
			pipelinesKeep: []*woodpecker.Pipeline{
				{Number: 1},
			},
			pipelines: []*woodpecker.Pipeline{
				{Number: 1},
				{Number: 2},
				{Number: 3},
			},
			wantDelete: 2,
		},
		{
			name:   "continue on 422 error",
			repoID: 1,
			args:   []string{"purge", "--older-than", "1h", "repo/name"},
			pipelinesKeep: []*woodpecker.Pipeline{
				{Number: 1},
			},
			pipelines: []*woodpecker.Pipeline{
				{Number: 1},
				{Number: 2},
				{Number: 3},
			},
			wantDelete: 2,
			mockDeleteError: &woodpecker.ClientError{
				StatusCode: 422,
				Message:    "test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			mockClient.On("RepoLookup", mock.Anything).Maybe().Return(&woodpecker.Repo{ID: tt.repoID}, nil)

			mockClient.On("PipelineList", mock.Anything, mock.Anything).Return(func(_ int64, opt woodpecker.PipelineListOptions) ([]*woodpecker.Pipeline, error) {
				// Return keep pipelines for first call
				if opt.Before.IsZero() {
					if opt.Page == 1 {
						return tt.pipelinesKeep, nil
					}
					return []*woodpecker.Pipeline{}, nil
				}

				// Return pipelines to purge for calls with Before filter
				if !opt.Before.IsZero() {
					if opt.Page == 1 {
						return tt.pipelines, nil
					}
					return []*woodpecker.Pipeline{}, nil
				}

				return []*woodpecker.Pipeline{}, nil
			}).Maybe()

			if tt.mockDeleteError != nil {
				mockClient.On("PipelineDelete", tt.repoID, mock.Anything).Return(tt.mockDeleteError)
			} else if tt.wantDelete > 0 {
				mockClient.On("PipelineDelete", tt.repoID, mock.Anything).Return(nil).Times(tt.wantDelete)
			}

			command := pipelinePurgeCmd
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) error {
				err := pipelinePurge(c, mockClient, time.Unix(1, 1))

				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
					return nil
				}

				assert.NoError(t, err)

				return nil
			}

			_ = command.Run(t.Context(), tt.args)
		})
	}
}
