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

package metadata

type Event string

// Event types corresponding to forge hooks.
const (
	EventPush         Event = "push"
	EventPull         Event = "pull_request"
	EventPullClosed   Event = "pull_request_closed"
	EventPullMetadata Event = "pull_request_metadata"
	EventTag          Event = "tag"
	EventRelease      Event = "release"
	EventDeploy       Event = "deployment"
	EventCron         Event = "cron"
	EventManual       Event = "manual"
)

func (event Event) IsPull() bool {
	switch event {
	case EventPull,
		EventPullClosed,
		EventPullMetadata:
		return true
	}
	return false
}

type Failure string

// Different ways to handle failure states.
const (
	FailureIgnore Failure = "ignore"
	FailureFail   Failure = "fail"
	FailureCancel Failure = "cancel"
)
