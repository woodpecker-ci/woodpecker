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

package atomgit

import (
	"encoding/json"
	"strconv"
	"time"
)

// id is a AtomGit identifier. AtomGit's real API serializes identifiers (user
// id, repository id, hook id, merge request iid, ...) as opaque hex strings
// (e.g. "6638af02bbeee41d0fe74c35"), but the local fixtures return them as
// JSON numbers. This type accepts both encodings and always stores the value
// as a string so it can be passed through to model.ForgeRemoteID unchanged.
type id string

// UnmarshalJSON decodes a value that is either a JSON number or a JSON string.
func (i *id) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*i = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*i = id(s)
		return nil
	}
	// JSON number (used by fixtures): render it without loss of precision.
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*i = id(n.String())
	return nil
}

// String returns the identifier as a string.
func (i id) String() string { return string(i) }

// Int64 parses the identifier as a base-10 integer when it is numeric.
// It is provided only for callers that need an integer (e.g. building URLs);
// AtomGit hex ids are not numeric and callers should prefer String().
func (i id) Int64() int64 {
	n, _ := strconv.ParseInt(string(i), 10, 64)
	return n
}

// boolInt decodes a AtomGit boolean flag that may arrive either as a JSON
// boolean (e.g. true) or as an integer (e.g. 0 / 1). It is used for fields
// like repository.public / repository.archived whose encoding differs between
// the real AtomGit API and the Gitee-style fixtures.
type boolInt bool

// UnmarshalJSON accepts both a JSON boolean and a 0/1 integer.
func (b *boolInt) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 't' || len(data) > 0 && data[0] == 'f' {
		var v bool
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*b = boolInt(v)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		// Fall back to a plain bool if the value is neither int nor bool-shaped.
		var v bool
		if e2 := json.Unmarshal(data, &v); e2 != nil {
			return err
		}
		*b = boolInt(v)
		return nil
	}
	*b = boolInt(n != 0)
	return nil
}

// Bool returns the underlying boolean value.
func (b boolInt) Bool() bool { return bool(b) }

// namespace represents the namespace (owner) of a repository on AtomGit.
type namespace struct {
	ID              id     `json:"id"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Kind            string `json:"kind"`
	FullPath        string `json:"full_path"`
	HTMLURL         string `json:"html_url"`
	AvatarURL       string `json:"avatar_url"`
	VisibilityLevel int    `json:"visibility_level"`
}

// permissions represents the access permissions a user has on a repository.
type permissions struct {
	ProjectAccess *projectAccess `json:"project_access"`
	GroupAccess   *projectAccess `json:"group_access"`
}

type projectAccess struct {
	AccessLevel int `json:"access_level"`
}

// user is the AtomGit user payload returned by /api/v5/user.
type user struct {
	ID        id     `json:"id"`
	Username  string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Bio       string `json:"bio"`
	Blog      string `json:"blog"`
	Company   string `json:"company"`
}

// repository is the AtomGit repository payload returned by the API.
type repository struct {
	ID                id           `json:"id"`
	HumanName         string       `json:"human_name"`
	Name              string       `json:"name"`
	Path              string       `json:"path"`
	PathWithNamespace string       `json:"path_with_namespace"`
	FullName          string       `json:"full_name"`
	Description       string       `json:"description"`
	HTMLURL           string       `json:"html_url"`
	WebURL            string       `json:"web_url"`
	GitURL            string       `json:"http_url_to_repo"`
	SSHURL            string       `json:"ssh_url_to_repo"`
	Namespace         *namespace   `json:"namespace"`
	DefaultBranch     string       `json:"default_branch"`
	VisibilityLevel   int          `json:"visibility_level"`
	Public            boolInt      `json:"public"`
	Archived          boolInt      `json:"archived"`
	AvatarURL         string       `json:"avatar_url"`
	Owner             *user        `json:"owner"`
	Permissions       *permissions `json:"permissions"`
	CreatedAt         time.Time    `json:"created_at"`
	LastActivityAt    time.Time    `json:"last_activity_at"`
	StarCount         int          `json:"star_count"`
	ForksCount        int          `json:"forks_count"`
	WatchCount        int          `json:"watch_count"`
	HasPullRequests   bool         `json:"has_pull_requests"`
}

// branch is the AtomGit branch payload returned by /api/v5/repos/{owner}/{repo}/branches.
type branch struct {
	Name      string  `json:"name"`
	Protected bool    `json:"protected"`
	Default   bool    `json:"default"`
	Commit    *commit `json:"commit"`
	HTMLURL   string  `json:"html_url"`
}

// commit is the AtomGit commit payload.
type commit struct {
	ID             string    `json:"id"`
	SHA            string    `json:"sha"`
	ShortID        string    `json:"short_id"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	AuthoredDate   time.Time `json:"authored_date"`
	CommittedDate  time.Time `json:"committed_date"`
	CommitterName  string    `json:"committer_name"`
	CommitterEmail string    `json:"committer_email"`
	CreatedAt      time.Time `json:"created_at"`
	URL            string    `json:"url"`
}

// hook is the AtomGit webhook payload returned by the hooks API.
type hook struct {
	ID                  id     `json:"id"`
	URL                 string `json:"url"`
	Password            string `json:"password"`
	ProjectID           id     `json:"project_id"`
	PushEvents          bool   `json:"push_events"`
	TagPushEvents       bool   `json:"tag_push_events"`
	IssuesEvents        bool   `json:"issues_events"`
	NoteEvents          bool   `json:"note_events"`
	MergeRequestsEvents bool   `json:"merge_requests_events"`
	CreatedAt           string `json:"created_at"`
}

// pullRequest is the AtomGit merge request payload.
type pullRequest struct {
	ID           id          `json:"id"`
	IID          id          `json:"iid"`
	ProjectID    id          `json:"project_id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	State        string      `json:"state"`
	MergeStatus  string      `json:"merge_status"`
	TargetBranch string      `json:"target_branch"`
	SourceBranch string      `json:"source_branch"`
	Upvotes      int         `json:"upvotes"`
	Downvotes    int         `json:"downvotes"`
	Author       *user       `json:"author"`
	Assignee     *user       `json:"assignee"`
	SourceRepo   *repository `json:"source_repo"`
	TargetRepo   *repository `json:"target_repo"`
	HTTPURL      string      `json:"html_url"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	ClosedAt     time.Time   `json:"closed_at"`
	MergedAt     time.Time   `json:"merged_at"`
	Draft        bool        `json:"draft"`
	Milestone    *milestone  `json:"milestone"`
	Labels       []label     `json:"labels"`
	LastCommit   *commit     `json:"last_commit"`
}

type milestone struct {
	ID          id     `json:"id"`
	IID         id     `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
}

type label struct {
	ID          id     `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// commitFile is a changed file entry in a merge request.
type commitFile struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	AMode       string `json:"a_mode"`
	BMode       string `json:"b_mode"`
	Diff        string `json:"diff"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}

// webhook push payload sent by AtomGit.
type pushHook struct {
	ObjectKind        string          `json:"object_kind"`
	EventType         string          `json:"event_name"`
	Before            string          `json:"before"`
	After             string          `json:"after"`
	Ref               string          `json:"ref"`
	CheckoutSHA       string          `json:"checkout_sha"`
	UserID            id              `json:"user_id"`
	UserName          string          `json:"user_name"`
	UserEmail         string          `json:"user_email"`
	UserAvatar        string          `json:"user_avatar"`
	ProjectID         id              `json:"project_id"`
	Project           *repository     `json:"project"`
	Repository        *repository     `json:"repository"`
	Commits           []payloadCommit `json:"commits"`
	TotalCommitsCount int             `json:"total_commits_count"`
}

// webhook tag push payload sent by AtomGit.
type tagPushHook struct {
	ObjectKind        string          `json:"object_kind"`
	EventType         string          `json:"event_name"`
	Before            string          `json:"before"`
	After             string          `json:"after"`
	Ref               string          `json:"ref"`
	CheckoutSHA       string          `json:"checkout_sha"`
	UserID            id              `json:"user_id"`
	UserName          string          `json:"user_name"`
	UserEmail         string          `json:"user_email"`
	UserAvatar        string          `json:"user_avatar"`
	ProjectID         id              `json:"project_id"`
	Project           *repository     `json:"project"`
	Repository        *repository     `json:"repository"`
	Commits           []payloadCommit `json:"commits"`
	TotalCommitsCount int             `json:"total_commits_count"`
}

type payloadCommit struct {
	ID        id           `json:"id"`
	Message   string       `json:"message"`
	Title     string       `json:"title"`
	Timestamp string       `json:"timestamp"`
	URL       string       `json:"url"`
	Author    commitAuthor `json:"author"`
	Added     []string     `json:"added"`
	Modified  []string     `json:"modified"`
	Removed   []string     `json:"removed"`
}

type commitAuthor struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// mergeRequestHook is the webhook payload for merge request events.
type mergeRequestHook struct {
	ObjectKind       string         `json:"object_kind"`
	EventType        string         `json:"event_name"`
	User             *user          `json:"user"`
	Project          *repository    `json:"project"`
	Repository       *repository    `json:"repository"`
	ObjectAttributes *pullRequest   `json:"object_attributes"`
	Labels           []label        `json:"labels"`
	Changes          map[string]any `json:"changes"`
}
