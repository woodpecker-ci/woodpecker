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
	"strings"

	pipelineConsts "go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
)

func createFilterFunc(agentFilter rpc.Filter) queue.FilterFn {
	return func(task *model.Task) (bool, int) {
		// ignore internal labels for filtering
		for k := range task.Labels {
			if strings.HasPrefix(k, pipelineConsts.InternalLabelPrefix) {
				delete(task.Labels, k)
			}
		}

		score := 0
		for taskLabel, taskLabelValue := range task.Labels {
			// if a task label is empty it will be ignored
			if taskLabelValue == "" {
				continue
			}

			// all task labels are required to be present for an agent to match
			agentLabelValue, ok := agentFilter.Labels[taskLabel]
			if !ok {
				return false, 0
			}

			switch agentLabelValue {
			// if agent label has a wildcard
			case "*":
				score++
			// if agent label has an exact match
			case taskLabelValue:
				score += 10
			// agent doesn't match
			default:
				return false, 0
			}
		}
		return true, score
	}
}
