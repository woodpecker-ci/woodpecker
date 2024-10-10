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
	isPullRequestEvent := pipeline.Event == model.EventPull
	isFork := true // TODO: implement this

	// events created by woodpecker itself should run right away
	if pipeline.Event == model.EventCron || pipeline.Event == model.EventManual {
		return
	}

	if repo.SecurityMode == model.SecurityModeNoRestrictions {
		return
	}

	if isPullRequestEvent && isFork && repo.SecurityMode == model.SecurityModeApproveForkPRs {
		pipeline.Status = model.StatusBlocked
	}

	if repo.SecurityMode == model.SecurityModeApproveEverything {
		pipeline.Status = model.StatusBlocked
	}
}
