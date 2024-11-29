package pipeline

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker/mocks"
)

func TestPipelinePurge(t *testing.T) {
	tests := []struct {
		name          string
		repoID        int64
		args          []string
		pipelinesKeep []*woodpecker.Pipeline
		pipelines     []*woodpecker.Pipeline
		wantDelete    int
		wantErr       error
	}{
		{
			name:   "success with no pipelines to purge",
			repoID: 1,
			args:   []string{"purge", "--older-than", "1h", "repo/name"},
			pipelinesKeep: []*woodpecker.Pipeline{
				{ID: 1},
			},
			pipelines: []*woodpecker.Pipeline{},
		},
		{
			name:   "success with pipelines to purge",
			repoID: 1,
			args:   []string{"purge", "--older-than", "1h", "repo/name"},
			pipelinesKeep: []*woodpecker.Pipeline{
				{ID: 1},
			},
			pipelines: []*woodpecker.Pipeline{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			wantDelete: 2,
		},
		{
			name:    "error on invalid duration",
			repoID:  1,
			args:    []string{"purge", "--older-than", "invalid", "repo/name"},
			wantErr: errors.New("time: invalid duration \"invalid\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewClient(t)
			mockClient.On("RepoLookup", mock.Anything).Maybe().Return(&woodpecker.Repo{ID: tt.repoID}, nil)

			mockClient.On("PipelineList", mock.Anything, mock.Anything).Return(func(_ int64, opt woodpecker.PipelineListOptions) ([]*woodpecker.Pipeline, error) {
				// Return keep pipelines for first call without Before/After filters
				if opt.Before.IsZero() && opt.After.IsZero() {
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

			if tt.wantDelete > 0 {
				mockClient.On("PipelineDelete", tt.repoID, mock.Anything).Return(nil).Times(tt.wantDelete)
			}

			command := buildPipelinePurgeCmd()
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) error {
				err := pipelinePurge(c, mockClient)

				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
					return nil
				}

				assert.NoError(t, err)

				return nil
			}

			_ = command.Run(context.Background(), tt.args)
		})
	}
}
