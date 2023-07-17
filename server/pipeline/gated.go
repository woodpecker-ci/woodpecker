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

import "github.com/woodpecker-ci/woodpecker/server/model"

func setGatedState(repo *model.Repo, pipe *model.Pipeline) {
	// TODO(336): extend gated feature with an allow/block List
	if repo.IsGated &&
		// events created by woodpecker itself should run right away
		pipe.Event != model.EventCron && pipe.Event != model.EventManual {
		pipe.Status = model.StatusBlocked
	}
}
