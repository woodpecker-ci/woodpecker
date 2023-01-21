// Copyright 2022 Woodpecker Authors
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

package forgejo

import "github.com/woodpecker-ci/woodpecker/server/forge/forgejo/client"

type pushHook struct {
	Sha     string `json:"sha"`
	Ref     string `json:"ref"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Compare string `json:"compare_url"`
	RefType string `json:"ref_type"`

	Pusher *client.User `json:"pusher"`

	Repo *client.Repository `json:"repository"`

	Commits []client.PayloadCommit `json:"commits"`

	HeadCommit client.PayloadCommit `json:"head_commit"`

	Sender *client.User `json:"sender"`
}

type pullRequestHook struct {
	Action      string              `json:"action"`
	Number      int64               `json:"number"`
	PullRequest *client.PullRequest `json:"pull_request"`
	Repo        *client.Repository  `json:"repository"`
	Sender      *client.User        `json:"sender"`
}
