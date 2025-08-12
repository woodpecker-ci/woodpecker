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
