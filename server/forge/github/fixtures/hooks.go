// Copyright 2018 Drone.IO Inc.
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

package fixtures

import _ "embed"

// HookPush is a sample push hook.
// https://developer.github.com/v3/activity/events/types/#pushevent
//
//go:embed HookPush.json
var HookPush string

// HookPushDeleted is a sample push hook that is marked as deleted, and is expected to be ignored.
const HookPushDeleted = `
{
  "deleted": true
}
`

// HookPullRequest is a sample hook pull request
// https://developer.github.com/v3/activity/events/types/#pullrequestevent
//
//go:embed HookPullRequest.json
var HookPullRequest string

// HookPullRequestInvalidAction is a sample hook pull request that has an
// action not equal to synchronize or opened, and is expected to be ignored.
const HookPullRequestInvalidAction = `
{
  "action": "reopened",
  "number": 1
}
`

// HookPullRequestInvalidState is a sample hook pull request that has a state
// not equal to open, and is expected to be ignored.
const HookPullRequestInvalidState = `
{
  "action": "synchronize",
  "pull_request": {
    "number": 1,
    "state": "closed"
  }
}
`

// HookPush is a sample deployment hook.
// https://developer.github.com/v3/activity/events/types/#deploymentevent
//
//go:embed HookDeploy.json
var HookDeploy string

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged string

// HookPullRequest is a sample hook pull request
// https://developer.github.com/v3/activity/events/types/#pullrequestevent
//
//go:embed HookPullRequestClosed.json
var HookPullRequestClosed string

//go:embed HookPullRequestEdited.json
var HookPullRequestEdited string

//go:embed HookRelease.json
var HookRelease string
