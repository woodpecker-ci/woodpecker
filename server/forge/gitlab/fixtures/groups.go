// Copyright 2026 Woodpecker Authors
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

// Two groups both named "woodpecker": the real top-level one and one an attacker
// created below their own namespace. They are only distinguishable by full_path.
var groupsPayload = []byte(`
[
	{
		"id": 1,
		"name": "woodpecker",
		"path": "woodpecker",
		"full_name": "woodpecker",
		"full_path": "woodpecker",
		"parent_id": null,
		"web_url": "https://gitlab.com/groups/woodpecker",
		"avatar_url": "https://gitlab.com/uploads/group/avatar/1/woodpecker.png"
	},
	{
		"id": 2,
		"name": "woodpecker",
		"path": "woodpecker",
		"full_name": "eve / woodpecker",
		"full_path": "eve/woodpecker",
		"parent_id": 99,
		"web_url": "https://gitlab.com/groups/eve/woodpecker",
		"avatar_url": "https://gitlab.com/uploads/group/avatar/2/woodpecker.png"
	}
]
`)
