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

import "go.woodpecker-ci.org/woodpecker/v3/server/model"

// list of statuses by their priority. Most important is on top.
var statusPriorityOrder = []model.StatusValue{
	// blocked, declined and created cannot appear in the
	// same workflow/pipeline at the same time
	model.StatusDeclined,
	model.StatusBlocked,
	model.StatusCreated,

	// errors have highest priority.
	model.StatusError,

	// skipped and killed cannot appear together with running/pending.
	model.StatusKilled,
	model.StatusSkipped,

	// running states
	model.StatusRunning,
	model.StatusPending,

	// finished states
	model.StatusFailure,
	model.StatusSuccess,
}

var priorityMap map[model.StatusValue]int = buildPriorityMap()

func buildPriorityMap() map[model.StatusValue]int {
	m := map[model.StatusValue]int{}
	for i, s := range statusPriorityOrder {
		m[s] = i
	}
	return m
}

func MergeStatusValues(s, t model.StatusValue) model.StatusValue {
	return statusPriorityOrder[min(priorityMap[s], priorityMap[t])]
}
