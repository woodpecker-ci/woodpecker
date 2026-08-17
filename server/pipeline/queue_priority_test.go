// Copyright 2026 Woodpecker Authors
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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestLoadQueuePriorityRulesUsesDefaultBranchHead(t *testing.T) {
	forge := mocks.NewMockForge(t)
	user := &model.User{}
	repo := &model.Repo{FullName: "postgis/postgis", Branch: "master"}
	pipeline := &model.Pipeline{
		Commit:      "pull-request-head",
		Event:       model.EventPull,
		EventReason: []string{"synchronized"},
	}

	forge.On("BranchHead", mock.Anything, user, repo, "master").
		Return(&model.Commit{SHA: "master-head"}, nil)
	forge.On("File", mock.Anything, user, repo, mock.MatchedBy(func(p *model.Pipeline) bool {
		return p.Commit == "master-head"
	}), queuePriorityConfigPath).
		Return([]byte("rules:\n  - priority: -20\n    event: pull_request\n    event_reason: synchroniz*\n"), nil)

	rules, err := loadQueuePriorityRules(t.Context(), forge, user, repo, pipeline)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, -20, model.QueuePriority(repo, pipeline, rules))
}
