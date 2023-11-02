// Copyright 2021 Woodpecker Authors
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

package gitlab

import (
	"github.com/xanzy/go-gitlab"

	"go.woodpecker-ci.org/woodpecker/server/model"
)

// getStatus is a helper that converts a Woodpecker status to a Gitlab status.
func getStatus(status model.StatusValue) gitlab.BuildStateValue {
	switch status {
	case model.StatusPending, model.StatusBlocked:
		return gitlab.Pending
	case model.StatusRunning:
		return gitlab.Running
	case model.StatusSuccess:
		return gitlab.Success
	case model.StatusFailure, model.StatusError:
		return gitlab.Failed
	case model.StatusKilled:
		return gitlab.Canceled
	default:
		return gitlab.Failed
	}
}
