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

// HookPush is a sample Gitea push hook.
//
//go:embed HookPush.json
var HookPush string

// HookPushMulti push multible commits to a branch.
//
//go:embed HookPushMulti.json
var HookPushMulti string

// HookPushBranch is a sample Gitea push hook where a new branch was created from an existing commit.
//
//go:embed HookPushBranch.json
var HookPushBranch string

// HookTag is a sample Gitea tag hook.
//
//go:embed HookTag.json
var HookTag string

// HookPullRequest is a sample pull_request webhook payload.
//
//go:embed HookPullRequest.json
var HookPullRequest string

//go:embed HookPullRequestUpdated.json
var HookPullRequestUpdated string

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged string

//go:embed HookPullRequestClosed.json
var HookPullRequestClosed string

const HookPullRequestChangeTitleHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request
`

//go:embed HookPullRequestChangeTitle.json
var HookPullRequestChangeTitle string

const HookPullRequestChangeBodyHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request
`

//go:embed HookPullRequestChangeBody.json
var HookPullRequestChangeBody string

const HookPullRequestAddReviewRequestHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_review_request
`

//go:embed HookPullRequestAddReviewRequest.json
var HookPullRequestAddReviewRequest string

const HookPullRequestAddLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`

//go:embed HookPullRequestAddLable.json
var HookPullRequestAddLable string

const HookPullRequestChangeLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`

//go:embed HookPullRequestChangeLable.json
var HookPullRequestChangeLable string

const HookPullRequestRemoveLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`

//go:embed HookPullRequestRemoveLable.json
var HookPullRequestRemoveLable string

const HookPullRequestAddMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`

//go:embed HookPullRequestAddMile.json
var HookPullRequestAddMile string

const HookPullRequestChangeMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`

//go:embed HookPullRequestChangeMile.json
var HookPullRequestChangeMile string

const HookPullRequestRemoveMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`

//go:embed HookPullRequestRemoveMile.json
var HookPullRequestRemoveMile string

//go:embed HookRelease.json
var HookRelease string
