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

package client

type QMap map[string]string

type Project struct {
	Id        int        `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Namespace *Namespace `json:"namespace,omitempty"`
}

type Namespace struct {
	Name string `json:"name,omitempty"`
}

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type hProject struct {
	HttpUrl           string `json:"http_url"`
	GitHttpUrl        string `json:"git_http_url"`
	AvatarUrl         string `json:"avatar_url"`
	VisibilityLevel   int    `json:"visibility_level"`
	WebUrl            string `json:"web_url"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
}

type hRepository struct {
	URL             string `json:"url,omitempty"`
	GitHttpUrl      string `json:"git_http_url,omitempty"`
	VisibilityLevel int    `json:"visibility_level,omitempty"`
}

type hCommit struct {
	Id      string  `json:"id,omitempty"`
	Message string  `json:"message,omitempty"`
	Author  *Person `json:"author,omitempty"`
}

type HookObjAttr struct {
	Title        string    `json:"title,omitempty"`
	IId          int       `json:"iid,omitempty"`
	SourceBranch string    `json:"source_branch,omitempty"`
	Url          string    `json:"url,omiyempty"`
	Source       *hProject `json:"source,omitempty"`
	Target       *hProject `json:"target,omitempty"`
	LastCommit   *hCommit  `json:"last_commit,omitempty"`
}

type HookPayload struct {
	After            string       `json:"after,omitempty"`
	Ref              string       `json:"ref,omitempty"`
	UserName         string       `json:"user_name,omitempty"`
	Project          *hProject    `json:"project,omitempty"`
	Repository       *hRepository `json:"repository,omitempty"`
	Commits          []hCommit    `json:"commits,omitempty"`
	ObjectKind       string       `json:"object_kind,omitempty"`
	ObjectAttributes *HookObjAttr `json:"object_attributes,omitempty"`
}
