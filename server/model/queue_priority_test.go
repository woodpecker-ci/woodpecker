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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueuePriorityRuleFile(t *testing.T) {
	rules, err := ParseQueuePriorityRuleFile([]byte(`
# default branches first
rules:
  - priority: 100
    event: push
    branch: master
  - priority: -20
    event: pull_request
    event_reason: synchroniz*
`))
	require.NoError(t, err)
	require.Len(t, rules, 2)

	assert.Equal(t, 100, rules[0].Priority)
	assert.Equal(t, "master", rules[0].Branch)
	assert.Equal(t, -20, rules[1].Priority)
	assert.Equal(t, "synchroniz*", rules[1].EventReason)
}

func TestParseQueuePriorityRuleFileRejectsInvalidRule(t *testing.T) {
	tests := []string{
		"rules:\n  - event: push\n",
		"rules:\n  - event: bogus\n    priority: 10\n",
		"rules:\n  - priority: 10\n    min_rerun_count: -1\n",
		"rules:\n  - priority: 10\n    branch: '['\n",
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			_, err := ParseQueuePriorityRuleFile([]byte(value))
			assert.Error(t, err)
		})
	}
}

func TestQueuePriority(t *testing.T) {
	repo := &Repo{FullName: "postgis/postgis"}
	pipeline := &Pipeline{
		Event:             EventPull,
		Branch:            "stable-3.6",
		Ref:               "refs/pull/42/head",
		Author:            "contributor",
		Sender:            "renovate-postgis",
		PullRequestLabels: []string{"needs-ci"},
		EventReason:       []string{"synchronized"},
		RerunCount:        2,
	}
	rules := []QueuePriorityRule{
		{Priority: 80, Repo: "postgis/postgis", Branch: "stable-*"},
		{Priority: 30, Event: EventPull, PullLabel: "needs-*"},
		{Priority: -25, Sender: "renovate*"},
		{Priority: -10, EventReason: "synchronized"},
		{Priority: 5, MinRerunCount: 2},
		{Priority: 900, Repo: "other/project"},
	}

	assert.Equal(t, 80, QueuePriority(repo, pipeline, rules))
}
