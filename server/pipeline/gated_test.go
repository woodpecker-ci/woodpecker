package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestSetGatedState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		repo          *model.Repo
		pipeline      *model.Pipeline
		expectBlocked bool
	}{
		{
			name: "by-pass for cron",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventCron,
			},
			expectBlocked: false,
		},
		{
			name: "by-pass for manual pipeline",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventManual,
			},
			expectBlocked: false,
		},
		{
			name: "require approval for fork PRs",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalForks,
			},
			pipeline: &model.Pipeline{
				Event:    model.EventPull,
				FromFork: true,
			},
			expectBlocked: true,
		},
		{
			name: "require approval for PRs",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalPullRequests,
			},
			pipeline: &model.Pipeline{
				Event:    model.EventPull,
				FromFork: false,
			},
			expectBlocked: true,
		},
		{
			name: "require approval for everything",
			repo: &model.Repo{
				RequireApproval: model.RequireApprovalAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventPush,
			},
			expectBlocked: true,
		},
	}

	for _, tc := range testCases {
		setApprovalState(tc.repo, tc.pipeline)
		assert.Equal(t, tc.expectBlocked, tc.pipeline.Status == model.StatusBlocked)
	}
}
