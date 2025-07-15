// Copyright 2022 Woodpecker Authors
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

package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// all agent labels need to be present in task
// tasks without labels can also be picked (also pick untagged mode)
// tasks with additional labels (not set on agent) wont be picked

func TestCreateFilterFunc(t *testing.T) {
	tests := []struct {
		name        string
		agentFilter rpc.Filter
		task        *model.Task
		wantMatched bool
		wantScore   int
	}{
		{
			name: "Two exact matches",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			wantMatched: true,
			wantScore:   20,
		},
		{
			name: "Wildcard and exact match",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "*", "platform": "linux"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			wantMatched: true,
			wantScore:   11,
		},
		{
			name: "Partial match",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "windows"},
			},
			wantMatched: false,
			wantScore:   0,
		},
		{
			name: "No match",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "456", "platform": "linux"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "windows"},
			},
			wantMatched: false,
			wantScore:   0,
		},
		{
			name: "Missing label",
			agentFilter: rpc.Filter{
				Labels: map[string]string{},
			},
			task: &model.Task{
				Labels: map[string]string{"needed": "some"},
			},
			wantMatched: false,
			wantScore:   0,
		},
		{
			name: "Empty task labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			task: &model.Task{
				Labels: map[string]string{},
			},
			wantMatched: true,
			wantScore:   0,
		},
		{
			name: "Agent with additional label",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "123", "platform": "linux", "extra": "value"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "linux", "empty": ""},
			},
			wantMatched: true,
			wantScore:   20,
		},
		{
			name: "Two wildcard matches",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"org-id": "*", "platform": "*"},
			},
			task: &model.Task{
				Labels: map[string]string{"org-id": "123", "platform": "linux"},
			},
			wantMatched: true,
			wantScore:   2,
		},
		{
			name: "Two different labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{"hello": "true"},
			},
			wantMatched: false,
			wantScore:   0,
		},
		{
			name: "Exact match",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{"docker": "true"},
			},
			wantMatched: true,
			wantScore:   10,
		},
		// TODO
		{
			name: "Agent without labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{},
			},
			task: &model.Task{
				Labels: map[string]string{"docker": "true"},
			},
			wantMatched: false,
			wantScore:   0,
		},
		{
			name: "Task without labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{},
			},
			wantMatched: true,
			wantScore:   0, // ???
		},
		// ??
		{
			name: "Agent and task without labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{},
			},
			task: &model.Task{
				Labels: map[string]string{},
			},
			wantMatched: true,
			wantScore:   0,
		},
		{
			name: "Multiple matching labels",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true", "shell": "true", "gpu": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{"docker": "true", "shell": "true", "gpu": "true"},
			},
			wantMatched: true,
			wantScore:   30,
		},
		{
			name: "Additional label in agent",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true", "shell": "true", "gpu": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{"docker": "true", "shell": "true"},
			},
			wantMatched: true,
			wantScore:   20,
		},
		{
			name: "Additional label in task",
			agentFilter: rpc.Filter{
				Labels: map[string]string{"docker": "true", "shell": "true"},
			},
			task: &model.Task{
				Labels: map[string]string{"docker": "true", "shell": "true", "gpu": "true"},
			},
			wantMatched: false,
			wantScore:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterFunc := createFilterFunc(tt.agentFilter)
			gotMatched, gotScore := filterFunc(tt.task)

			assert.Equal(t, tt.wantMatched, gotMatched, "Matched result")
			assert.Equal(t, tt.wantScore, gotScore, "Score")
		})
	}
}
