// Copyright 2023 Woodpecker Authors
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

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

func setGatedState(repo *model.Repo, pipeline *model.Pipeline) {
	// skip events created by woodpecker itself
	if pipeline.Event == model.EventCron || pipeline.Event == model.EventManual {
		return
	}

	// always require approval for pull requests from forks
	if pipeline.Event == model.EventPull && pipeline.FromFork {
		pipeline.Status = model.StatusBlocked
		return
	}

	// repository requires approval for pull requests
	if pipeline.Event == model.EventPull && repo.ApprovalMode == model.ApprovalModePullRequests {
		pipeline.Status = model.StatusBlocked
		return
	}

	// repository requires approval for all events
	if repo.ApprovalMode == model.ApprovalModeAllEvents {
		pipeline.Status = model.StatusBlocked
		return
	}
}
