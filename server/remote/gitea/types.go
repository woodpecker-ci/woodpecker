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

package gitea

type giteaUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Login    string `json:"login"`
	Avatar   string `json:"avatar_url"`
}

type giteaRepo struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	FullName string    `json:"full_name"`
	URL      string    `json:"html_url"`
	Private  bool      `json:"private"`
	Owner    giteaUser `json:"owner,omitempty"`
}

type giteaSender struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Username string `json:"username"`
	Name     string `json:"full_name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
}

type pushHook struct {
	Sha     string `json:"sha"`
	Ref     string `json:"ref"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Compare string `json:"compare_url"`
	RefType string `json:"ref_type"`

	Sender giteaSender `json:"sender"`
	Repo   giteaRepo   `json:"repository"`
	Pusher giteaUser   `json:"pusher"`

	Commits []struct {
		ID       string   `json:"id"`
		Message  string   `json:"message"`
		URL      string   `json:"url"`
		Added    []string `json:"added"`
		Removed  []string `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
}

type pullRequestHook struct {
	Action string      `json:"action"`
	Number int64       `json:"number"`
	Repo   giteaRepo   `json:"repository"`
	Sender giteaSender `json:"sender"`

	PullRequest struct {
		ID   int64 `json:"id"`
		User struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Name     string `json:"full_name"`
			Email    string `json:"email"`
			Avatar   string `json:"avatar_url"`
		} `json:"user"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		URL       string `json:"html_url"`
		Mergeable bool   `json:"mergeable"`
		Merged    bool   `json:"merged"`
		MergeBase string `json:"merge_base"`
		Base      struct {
			Label string `json:"label"`
			Ref   string `json:"ref"`
			Sha   string `json:"sha"`
			Repo  struct {
				ID       int64     `json:"id"`
				Name     string    `json:"name"`
				FullName string    `json:"full_name"`
				URL      string    `json:"html_url"`
				Private  bool      `json:"private"`
				Owner    giteaUser `json:"owner"`
			} `json:"repo"`
		} `json:"base"`
		Head struct {
			Label string `json:"label"`
			Ref   string `json:"ref"`
			Sha   string `json:"sha"`
			Repo  struct {
				ID       int64     `json:"id"`
				Name     string    `json:"name"`
				FullName string    `json:"full_name"`
				URL      string    `json:"html_url"`
				Private  bool      `json:"private"`
				Owner    giteaUser `json:"owner"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
}

type releaseHook struct {
	Action  string      `json:"action"`
	Repo    giteaRepo   `json:"repository"`
	Sender  giteaSender `json:"sender"`
	Release struct {
		ID              int64     `json:"id"`
		TagName         string    `json:"tag_name"`
		TargetCommitish string    `json:"target_commitish"`
		Name            string    `json:"name"`
		Body            string    `json:"body"`
		URL             string    `json:"url"`
		HTMLURL         string    `json:"html_url"`
		TarballURL      string    `json:"tarball_url"`
		ZipballURL      string    `json:"zipball_url"`
		Draft           bool      `json:"draft"`
		Prerelease      bool      `json:"prerelease"`
		CreatedAt       string    `json:"created_at"`
		PublishedAt     string    `json:"published_at"`
		Assets          []string  `json:"assets"`
		Author          giteaUser `json:"author"`
	}
}
