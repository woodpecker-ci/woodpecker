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

package internal

import (
	"net/url"
	"strconv"
	"time"
)

type Account struct {
	ID    int64  `json:"id"`
	Login string `json:"username"`
	Name  string `json:"display_name"`
	Type  string `json:"type"`
	Links Links  `json:"links"`
}

type Workspace struct {
	UUID  string `json:"uuid"`
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Links Links  `json:"links"`
}

type WorkspacesResp struct {
	Page   int          `json:"page"`
	Pages  int          `json:"pagelen"`
	Size   int          `json:"size"`
	Next   string       `json:"next"`
	Values []*Workspace `json:"values"`
}

type PipelineStatus struct {
	State string `json:"state"`
	Key   string `json:"key"`
	Name  string `json:"name,omitempty"`
	URL   string `json:"url"`
	Desc  string `json:"description,omitempty"`
}

type Email struct {
	Email       string `json:"email"`
	IsConfirmed bool   `json:"is_confirmed"`
	IsPrimary   bool   `json:"is_primary"`
}

type EmailResp struct {
	Page   int      `json:"page"`
	Pages  int      `json:"pagelen"`
	Size   int      `json:"size"`
	Next   string   `json:"next"`
	Values []*Email `json:"values"`
}

type Hook struct {
	UUID   string   `json:"uuid,omitempty"`
	Desc   string   `json:"description"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

type HookResp struct {
	Page   int     `json:"page"`
	Pages  int     `json:"pagelen"`
	Size   int     `json:"size"`
	Next   string  `json:"next"`
	Values []*Hook `json:"values"`
}

type Links struct {
	Self   Link   `json:"self"`
	Avatar Link   `json:"avatar"`
	HTML   Link   `json:"html"`
	Clone  []Link `json:"clone"`
}

type Link struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type LinkClone struct {
	Link
}

type Repo struct {
	UUID      string  `json:"uuid"`
	Owner     Account `json:"owner"`
	Name      string  `json:"name"`
	FullName  string  `json:"full_name"`
	Language  string  `json:"language"`
	IsPrivate bool    `json:"is_private"`
	Scm       string  `json:"scm"`
	Desc      string  `json:"desc"`
	Links     Links   `json:"links"`
}

type RepoResp struct {
	Page   int     `json:"page"`
	Pages  int     `json:"pagelen"`
	Size   int     `json:"size"`
	Next   string  `json:"next"`
	Values []*Repo `json:"values"`
}

type Change struct {
	New struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Target struct {
			Type    string    `json:"type"`
			Hash    string    `json:"hash"`
			Message string    `json:"message"`
			Date    time.Time `json:"date"`
			Links   Links     `json:"links"`
			Author  struct {
				Raw  string  `json:"raw"`
				User Account `json:"user"`
			} `json:"author"`
		} `json:"target"`
	} `json:"new"`
}

type PushHook struct {
	Actor Account `json:"actor"`
	Repo  Repo    `json:"repository"`
	Push  struct {
		Changes []Change `json:"changes"`
	} `json:"push"`
}

type PullRequestHook struct {
	Actor       Account `json:"actor"`
	Repo        Repo    `json:"repository"`
	PullRequest struct {
		ID      int       `json:"id"`
		Type    string    `json:"type"`
		Reason  string    `json:"reason"`
		Desc    string    `json:"description"`
		Title   string    `json:"title"`
		State   string    `json:"state"`
		Links   Links     `json:"links"`
		Created time.Time `json:"created_on"`
		Updated time.Time `json:"updated_on"`

		Source struct {
			Repo   Repo `json:"repository"`
			Commit struct {
				Hash  string `json:"hash"`
				Links Links  `json:"links"`
			} `json:"commit"`
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`

		Dest struct {
			Repo   Repo `json:"repository"`
			Commit struct {
				Hash  string `json:"hash"`
				Links Links  `json:"links"`
			} `json:"commit"`
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"destination"`
	} `json:"pullrequest"`
}

type WorkspaceMembershipResp struct {
	Page   int    `json:"page"`
	Pages  int    `json:"pagelen"`
	Size   int    `json:"size"`
	Next   string `json:"next"`
	Values []struct {
		Permission string `json:"permission"`
		User       struct {
			Nickname string `json:"nickname"`
		}
	} `json:"values"`
}

type ListOpts struct {
	Page    int
	PageLen int
}

func (o *ListOpts) Encode() string {
	params := url.Values{}
	if o.Page != 0 {
		params.Set("page", strconv.Itoa(o.Page))
	}
	if o.PageLen != 0 {
		params.Set("pagelen", strconv.Itoa(o.PageLen))
	}
	return params.Encode()
}

type ListWorkspacesOpts struct {
	Page    int
	PageLen int
	Role    string
}

func (o *ListWorkspacesOpts) Encode() string {
	params := url.Values{}
	if o.Page != 0 {
		params.Set("page", strconv.Itoa(o.Page))
	}
	if o.PageLen != 0 {
		params.Set("pagelen", strconv.Itoa(o.PageLen))
	}
	if len(o.Role) != 0 {
		params.Set("role", o.Role)
	}
	return params.Encode()
}

type Error struct {
	Status int
	Body   struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e Error) Error() string {
	return e.Body.Message
}

type RepoPermResp struct {
	Page   int         `json:"page"`
	Pages  int         `json:"pagelen"`
	Values []*RepoPerm `json:"values"`
}

type RepoPerm struct {
	Permission string `json:"permission"`
}

type BranchResp struct {
	Values []*Branch `json:"values"`
}

type Branch struct {
	Name string `json:"name"`
}
