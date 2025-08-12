// Copyright 2018 Drone.IO Inc.
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

package fixtures

import _ "embed"

const HookPushEmptyHash = `
{
  "push": {
    "changes": [
      {
        "new": {
          "type": "branch",
          "target": { "hash": "" }
        }
      }
    ]
  }
}
`

//go:embed HookPush.json
var HookPush string

var HookPushHeaders = map[string][]string{
	"X-Event-Key":    {"repo:push"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"a9a1e3d4-de91-4e45-9fc3-941976118544"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullRequestRejected.json
var HookPullRequestRejected string

var HookPullRequestRejectedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:rejected"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"cfadee1f-282d-4dc5-9abb-a8067960868b"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullRequestCreated.json
var HookPullRequestCreated string

var HookPullRequestCreatedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:created"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"b6a7936d-956e-4d6d-a5ad-3e179249add6"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullPush.json
var HookPullPush string

var HookPullPushHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:updated"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"77429825-4efc-4326-916e-bd4c1a18546a"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullToDraft.json
var HookPullToDraft string

var HookPullToDraftHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:updated"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"f8916d54-02f3-41f7-a38e-8eee3d72a00d"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullReadyFromDraft.json
var HookPullReadyFromDraft string

var HookPullReadyFromDraftHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:updated"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"1b49a8cb-e36d-49c4-99ec-ee2c800832ba"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullCommentCreated.json
var HookPullCommentCreated string

var HookPullCommentCreatedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:comment_created"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"325ce2a9-4d86-4fae-ab59-7d1376fe0438"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullChangesRequestCreated.json
var HookPullChangesRequestCreated string

var HookPullChangesRequestCreatedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:changes_request_created"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"5cb41f6b-531a-469a-8fb0-9489f4530dbd"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullChangesRequestRemoved.json
var HookPullChangesRequestRemoved string

var HookPullChangesRequestRemovedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:changes_request_removed"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"78ff767e-741d-4137-a546-b1fc526ffb79"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullApproved.json
var HookPullApproved string

var HookPullApprovedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:approved"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"23ac9e91-8346-4e2a-8238-1d244d6b0138"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullUnapproved.json
var HookPullUnapproved string

var HookPullUnapprovedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:unapproved"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"b612e0dc-6fa5-4075-8440-1c6a0d67262c"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullMerged.json
var HookPullMerged string

var HookPullMergedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:fulfilled"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"a7ff3cfe-b917-4c96-b9b9-64b6ad58e893"},
	"Content-Type":   {"application/json"},
}

//go:embed HookPullReviewRequested.json
var HookPullReviewRequested string

var HookPullReviewRequestedHeaders = map[string][]string{
	"X-Event-Key":    {"pullrequest:updated"},
	"X-Hook-UUID":    {"81424cfd-6ea3-47d0-bb73-ae76bb3fb3a0"},
	"User-Agent":     {"Bitbucket-Webhooks/2.0"},
	"X-Request-UUID": {"407b8a5f-a397-45ea-83ab-d08971f0cf03"},
	"Content-Type":   {"application/json"},
}
