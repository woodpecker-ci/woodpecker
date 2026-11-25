// Copyright 2024 Woodpecker Authors
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

// HookPush is a sample Forgejo push hook.
//
//go:embed HookPush.json
var HookPush []byte

// HookPushMulti push multible commits to a branch.
//
//go:embed HookPushMulti.json
var HookPushMulti []byte

// HookPushBranch is a sample Forgejo push hook where a new branch was created from an existing commit.
//
//go:embed HookPushBranch.json
var HookPushBranch []byte

// HookTag is a sample Forgejo tag hook.
//
//go:embed HookTag.json
var HookTag []byte

// HookPullRequest is a sample pull_request webhook payload.
//
//go:embed HookPullRequest.json
var HookPullRequest []byte

//go:embed HookPullRequestUpdated.json
var HookPullRequestUpdated []byte

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged []byte

//go:embed HookPullRequestClosed.json
var HookPullRequestClosed []byte

//go:embed HookPullRequestEdited.json
var HookPullRequestEdited []byte

//go:embed HookRelease.json
var HookRelease []byte

//go:embed HookPullRequestAssigneesAdded.json
var HookPullRequestAssigneesAdded []byte

//go:embed HookPullRequestMilestoneAdded.json
var HookPullRequestMilestoneAdded []byte

//go:embed HookPullRequestLabelAdded.json
var HookPullRequestLabelAdded []byte

//go:embed HookPullRequestAssigneeCleared.json
var HookPullRequestAssigneeCleared []byte

//go:embed HookPullRequestMilestoneChanged.json
var HookPullRequestMilestoneChanged []byte

//go:embed HookPullRequestLabelsUpdated.json
var HookPullRequestLabelsUpdated []byte

//go:embed HookPullRequestLabelsCleared.json
var HookPullRequestLabelsCleared []byte

//go:embed HookPullRequestMilestoneCleared.json
var HookPullRequestMilestoneCleared []byte

//go:embed HookPullRequestReopened.json
var HookPullRequestReopened []byte
