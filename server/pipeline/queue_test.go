// Copyright 2026 Woodpecker Authors
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestQueuePipelineConcurrency(t *testing.T) {
	repo := &model.Repo{ID: 7}
	activePipeline := &model.Pipeline{ID: 42}

	tests := []struct {
		name          string
		item          *builder.Item
		expectedLimit int
		expectedGroup string
	}{
		{
			name: "no limit leaves concurrency unset",
			item: &builder.Item{
				Workflow: &builder.Workflow{ID: 1, Name: "build"},
			},
			expectedLimit: 0,
			expectedGroup: "",
		},
		{
			name: "explicit group is scoped to the repo",
			item: &builder.Item{
				Workflow:         &builder.Workflow{ID: 2, Name: "build"},
				ConcurrencyLimit: 2,
				ConcurrencyGroup: "deploy",
			},
			expectedLimit: 2,
			expectedGroup: "7//deploy",
		},
		{
			name: "empty group defaults to the workflow name",
			item: &builder.Item{
				Workflow:         &builder.Workflow{ID: 3, Name: "test"},
				ConcurrencyLimit: 1,
			},
			expectedLimit: 1,
			expectedGroup: "7/test/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := pipelineTasks(repo, nil, activePipeline, []*builder.Item{tt.item})
			require.NoError(t, err)
			require.Len(t, tasks, 1)

			task := tasks[0]
			assert.Equal(t, tt.expectedLimit, task.ConcurrencyLimit)
			assert.Equal(t, tt.expectedGroup, task.ConcurrencyGroup)
		})
	}
}

func TestQueuePipelineCreated(t *testing.T) {
	repo := &model.Repo{ID: 7}
	item := &builder.Item{Workflow: &builder.Workflow{ID: 1, Name: "build"}}

	runOnce := func(t *testing.T, pipeline *model.Pipeline) *model.Task {
		t.Helper()
		tasks, err := pipelineTasks(repo, nil, pipeline, []*builder.Item{item})
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		return tasks[0]
	}

	t.Run("inherits the pipeline creation time", func(t *testing.T) {
		task := runOnce(t, &model.Pipeline{ID: 42, Created: 1700000000})
		assert.Equal(t, int64(1700000000), task.Created)
	})

	t.Run("falls back to now when the pipeline has no creation time", func(t *testing.T) {
		before := time.Now().Unix()
		task := runOnce(t, &model.Pipeline{ID: 42})
		assert.GreaterOrEqual(t, task.Created, before)
	})
}

func TestQueuePipelineOrgAgentsOnly(t *testing.T) {
	repo := &model.Repo{ID: 7, OrgID: 3}
	activePipeline := &model.Pipeline{ID: 42}

	newItem := func() *builder.Item {
		return &builder.Item{
			Workflow: &builder.Workflow{ID: 1, Name: "build"},
			Labels:   map[string]string{pipeline.LabelOrgAgentsOnly: "true"},
		}
	}

	tests := []struct {
		name     string
		org      *model.Org
		expected bool
	}{
		{name: "no org", org: nil, expected: false},
		{name: "org not requiring org agents", org: &model.Org{ID: 3}, expected: false},
		{name: "org requiring org agents", org: &model.Org{ID: 3, RequireOrgAgents: true}, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := pipelineTasks(repo, tt.org, activePipeline, []*builder.Item{newItem()})
			require.NoError(t, err)
			require.Len(t, tasks, 1)

			value, ok := tasks[0].Labels[pipeline.LabelOrgAgentsOnly]
			assert.Equal(t, tt.expected, ok)
			if tt.expected {
				assert.Equal(t, "true", value)
			}
			assert.Equal(t, "3", tasks[0].Labels[pipeline.LabelFilterOrg])
		})
	}
}
