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

	"go.woodpecker-ci.org/woodpecker/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/server/model"
)

func TestCreateFilterFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		agentLabels map[string]string
		task        model.Task
		exp         bool
	}{
		{
			name:        "agent with missing labels",
			agentLabels: map[string]string{"repo": "test/woodpecker"},
			task: model.Task{
				Labels: map[string]string{"platform": "linux/amd64", "repo": "test/woodpecker"},
			},
			exp: false,
		},
		{
			name:        "agent with wrong labels",
			agentLabels: map[string]string{"platform": "linux/arm64"},
			task: model.Task{
				Labels: map[string]string{"platform": "linux/amd64"},
			},
			exp: false,
		},
		{
			name:        "agent with correct labels",
			agentLabels: map[string]string{"platform": "linux/amd64", "location": "europe"},
			task: model.Task{
				Labels: map[string]string{"platform": "linux/amd64", "location": "europe"},
			},
			exp: true,
		},
		{
			name:        "agent with additional labels",
			agentLabels: map[string]string{"platform": "linux/amd64", "location": "europe"},
			task: model.Task{
				Labels: map[string]string{"platform": "linux/amd64"},
			},
			exp: true,
		},
		{
			name:        "agent with wildcard label",
			agentLabels: map[string]string{"platform": "linux/amd64", "location": "*"},
			task: model.Task{
				Labels: map[string]string{"platform": "linux/amd64", "location": "america"},
			},
			exp: true,
		},
		{
			name:        "agent with platform label and task without",
			agentLabels: map[string]string{"platform": "linux/amd64"},
			task: model.Task{
				Labels: map[string]string{"platform": ""},
			},
			exp: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fn, err := createFilterFunc(rpc.Filter{Labels: test.agentLabels})
			if !assert.NoError(t, err) {
				t.Fail()
			}

			assert.EqualValues(t, test.exp, fn(&test.task))
		})
	}
}
