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

			mockClient.On("PipelineList", tt.repoID, mock.MatchedBy(func(opt woodpecker.PipelineListOptions) bool {
				return opt.PerPage > 0
			})).Maybe().Return(tt.pipelinesKeep, nil)

			mockClient.On("PipelineList", tt.repoID, mock.MatchedBy(func(opt woodpecker.PipelineListOptions) bool {
				return !opt.Before.IsZero() || !opt.After.IsZero()
			})).Maybe().Return(tt.pipelines, nil)

			if tt.wantDelete > 0 {
				mockClient.On("PipelineDelete", tt.repoID, mock.Anything).Return(nil)
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

				mockClient.AssertNumberOfCalls(t, "PipelineDelete", tt.wantDelete)

				return nil
			}

			_ = command.Run(context.Background(), tt.args)
		})
	}
}
