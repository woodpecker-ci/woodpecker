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

//go:embed HookPush.json
var HookPush string

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

//go:embed HookPull.json
var HookPull string

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged string

//go:embed HookPullRequestDeclined.json
var HookPullRequestDeclined string
