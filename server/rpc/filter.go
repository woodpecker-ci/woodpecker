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

package rpc

import (
	"maps"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
)

func createFilterFunc(agentFilter rpc.Filter) queue.FilterFn {
	return func(task *model.Task) (bool, int) {
		// Create a copy of the labels for filtering to avoid modifying the original task
		labels := maps.Clone(task.Labels)

		if requiredLabelsMissing(labels, agentFilter.Labels) {
			return false, 0
		}

		// ignore internal labels for filtering
		for k := range labels {
			if strings.HasPrefix(k, pipeline.InternalLabelPrefix) {
				delete(labels, k)
			}
		}

		score := 0
		for taskLabel, taskLabelValue := range labels {
			// if a task label is empty it will be ignored
			if taskLabelValue == "" {
				continue
			}

			// all task labels are required to be present for an agent to match
			agentLabelValue, ok := agentFilter.Labels[taskLabel]
			if !ok {
				// Check for required label
				agentLabelValue, ok = agentFilter.Labels["!"+taskLabel]
				if !ok {
					return false, 0
				}
			}

			matched := false
			for _, val := range strings.Split(agentLabelValue, ",") {
				if val == "*" || val == taskLabelValue {
					matched = true
					if val == "*" {
						score++
					} else {
						score += 10
					}
					break
				}
			}

			if !matched {
				return false, 0
			}
		}
		return true, score
	}
}

func requiredLabelsMissing(taskLabels, agentLabels map[string]string) bool {
	for label, value := range agentLabels {
		if len(label) > 0 && label[0] == '!' {
			val, ok := taskLabels[label[1:]]
			if !ok {
				return true
			}

			found := false
			for _, v := range strings.Split(value, ",") {
				if val == v {
					found = true
					break
				}
			}
			if !found {
				return true
			}
		}
	}
	return false
}
