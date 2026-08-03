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

func TestParseQueuePriorityRules(t *testing.T) {
	rules, err := ParseQueuePriorityRules([]string{
		"priority=100 repo=postgis/postgis event=push branch=stable-*",
		"priority=-20 event=pull_request event_reason=synchronized sender=renovate*",
	})
	require.NoError(t, err)
	require.Len(t, rules, 2)

	assert.Equal(t, 100, rules[0].Priority)
	assert.Equal(t, "postgis/postgis", rules[0].Repo)
	assert.Equal(t, EventPush, rules[0].Event)
	assert.Equal(t, "stable-*", rules[0].Branch)

	assert.Equal(t, -20, rules[1].Priority)
	assert.Equal(t, EventPull, rules[1].Event)
	assert.Equal(t, "synchronized", rules[1].EventReason)
	assert.Equal(t, "renovate*", rules[1].Sender)
}

func TestParseQueuePriorityRuleFile(t *testing.T) {
	rules, err := ParseQueuePriorityRuleFile([]byte(`
# default branches first
priority=100 event=push branch=master
priority=-20 event=pull_request event_reason=synchroniz*
`))
	require.NoError(t, err)
	require.Len(t, rules, 2)

	assert.Equal(t, 100, rules[0].Priority)
	assert.Equal(t, "master", rules[0].Branch)
	assert.Equal(t, -20, rules[1].Priority)
	assert.Equal(t, "synchroniz*", rules[1].EventReason)
}

func TestParseQueuePriorityRulesRejectsInvalidRule(t *testing.T) {
	tests := []string{
		"event=push",
		"priority=bad event=push",
		"priority=10 nope=value",
		"priority=10 event=bogus",
		"priority=10 min_rerun_count=nope",
		"priority=10 min_rerun_count=-1",
		"priority=10 branch=[",
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			_, err := ParseQueuePriorityRules([]string{value})
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
	rules, err := ParseQueuePriorityRules([]string{
		"priority=80 repo=postgis/postgis branch=stable-*",
		"priority=30 event=pull_request pr_label=needs-*",
		"priority=-25 sender=renovate*",
		"priority=-10 event_reason=synchronized",
		"priority=5 min_rerun_count=2",
		"priority=900 repo=other/project",
	})
	require.NoError(t, err)

	assert.Equal(t, 80, QueuePriority(repo, pipeline, rules))
}
