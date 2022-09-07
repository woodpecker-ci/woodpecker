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

package gogs

import "github.com/gogits/go-gogs-client"

type pushHook struct {
	Ref     string `json:"ref"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Compare string `json:"compare_url"`
	RefType string `json:"ref_type"`

	Pusher *gogs.User `json:"pusher"`

	Repo *gogs.Repository `json:"repository"`

	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	} `json:"commits"`

	Sender *gogs.User `json:"sender"`
}

type pullRequestHook struct {
	Action      string `json:"action"`
	Number      int64  `json:"number"`
	PullRequest struct {
		ID         int64      `json:"id"`
		User       *gogs.User `json:"user"`
		Title      string     `json:"title"`
		Body       string     `json:"body"`
		State      string     `json:"state"`
		URL        string     `json:"html_url"`
		Mergeable  bool       `json:"mergeable"`
		Merged     bool       `json:"merged"`
		MergeBase  string     `json:"merge_base"`
		BaseBranch string     `json:"base_branch"`
		Base       struct {
			Label string           `json:"label"`
			Ref   string           `json:"ref"`
			Sha   string           `json:"sha"`
			Repo  *gogs.Repository `json:"repo"`
		} `json:"base"`
		HeadBranch string `json:"head_branch"`
		Head       struct {
			Label string           `json:"label"`
			Ref   string           `json:"ref"`
			Sha   string           `json:"sha"`
			Repo  *gogs.Repository `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
	Repo   *gogs.Repository `json:"repository"`
	Sender *gogs.User       `json:"sender"`
}
