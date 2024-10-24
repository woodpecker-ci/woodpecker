package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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
			name: "no restrictions",
			repo: &model.Repo{
				ApprovalMode: model.ApprovalModeAllOutsideCollaborators,
			},
			pipeline: &model.Pipeline{
				Event: model.EventPull,
			},
			expectBlocked: false,
		},
		{
			name: "require approval for fork PRs",
			repo: &model.Repo{
				ApprovalMode: model.ApprovalModeAllOutsideCollaborators,
			},
			pipeline: &model.Pipeline{
				Event: model.EventPull,
			},
			expectBlocked: true,
		},
		{
			name: "by-pass for cron / manual events",
			repo: &model.Repo{
				ApprovalMode: model.ApprovalModeAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventCron,
			},
			expectBlocked: false,
		},

		{
			name: "require approval for everything",
			repo: &model.Repo{
				ApprovalMode: model.ApprovalModeAllEvents,
			},
			pipeline: &model.Pipeline{
				Event: model.EventPush,
			},
			expectBlocked: true,
		},
	}

	for _, tc := range testCases {
		setGatedState(tc.repo, tc.pipeline)
		assert.Equal(t, tc.expectBlocked, tc.pipeline.Status == model.StatusBlocked)
	}
}
