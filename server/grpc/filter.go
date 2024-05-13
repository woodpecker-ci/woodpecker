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
	"github.com/bmatcuk/doublestar/v4"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
)

func createFilterFunc(agentFilter rpc.Filter) queue.FilterFn {
	return func(task *model.Task) bool {
		for taskLabel, taskLabelValue := range task.Labels {
			// if a task label is empty it will be ignored
			if taskLabelValue == "" {
				continue
			}

			agentLabelValue, ok := agentFilter.Labels[taskLabel]

			if !ok {
				return false
			}

			// if agent label has a wildcard
			if agentLabelValue == "*" {
				continue
			}

			match, err := doublestar.Match(taskLabelValue, agentLabelValue)
			if err != nil {
				log.Error().Err(err).Msg("got unexpected error while try to match task and agent lable value")
			}
			return match
		}
		return true
	}
}
