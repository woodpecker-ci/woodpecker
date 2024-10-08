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

func TestPipelineList(t *testing.T) {
	testtases := []struct {
		name        string
		repoID      int64
		repoErr     error
		pipelines   []*woodpecker.Pipeline
		pipelineErr error
		args        []string
		expected    []woodpecker.Pipeline
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
			expected: []woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
		},
		{
			name:   "filter by branch",
			repoID: 1,
			pipelines: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
			args: []string{"ls", "--branch", "main", "repo/name"},
			expected: []woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
		},
		{
			name:   "filter by event",
			repoID: 1,
			pipelines: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
			args: []string{"ls", "--event", "push", "repo/name"},
			expected: []woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
		},
		{
			name:   "filter by status",
			repoID: 1,
			pipelines: []*woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
				{ID: 2, Branch: "develop", Event: "pull_request", Status: "running"},
				{ID: 3, Branch: "main", Event: "push", Status: "failure"},
			},
			args: []string{"ls", "--status", "success", "repo/name"},
			expected: []woodpecker.Pipeline{
				{ID: 1, Branch: "main", Event: "push", Status: "success"},
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
			expected: []woodpecker.Pipeline{
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
			mockClient := mocks.NewClient(t)
			mockClient.On("PipelineList", mock.Anything, mock.Anything).Return(tt.pipelines, tt.pipelineErr)
			mockClient.On("RepoLookup", mock.Anything).Return(&woodpecker.Repo{ID: tt.repoID}, nil)

			command := pipelineListCmd
			command.Writer = io.Discard
			command.Action = func(ctx context.Context, c *cli.Command) error {
				pipelines, err := pipelineList(ctx, c, mockClient)
				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
					return nil
				}

				assert.NoError(t, err)
				assert.EqualValues(t, tt.expected, pipelines)

				return nil
			}

			_ = command.Run(context.Background(), tt.args)
		})
	}
}
