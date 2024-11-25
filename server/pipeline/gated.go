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

func setApprovalState(repo *model.Repo, pipeline *model.Pipeline) {
	if !needsApproval(repo, pipeline) {
		return
	}

	// set pipeline status to blocked and require approval
	pipeline.Status = model.StatusBlocked
}

func needsApproval(repo *model.Repo, pipeline *model.Pipeline) bool {
	// skip events created by woodpecker itself
	if pipeline.Event == model.EventCron || pipeline.Event == model.EventManual {
		return false
	}

	switch repo.RequireApproval {
	// repository allows all events without approval
	case model.RequireApprovalNone:
		return false

	// repository requires approval for pull requests from forks
	case model.RequireApprovalForks:
		if pipeline.Event == model.EventPull && pipeline.FromFork {
			return true
		}

	// repository requires approval for pull requests
	case model.RequireApprovalPullRequests:
		if pipeline.Event == model.EventPull {
			return true
		}

	case model.RequireApprovalAllEvents:
		return true
	}

	return false
}
