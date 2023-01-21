// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
)

type Repository struct {
	ID            int64       `json:"id"`
	Owner         *User       `json:"owner"`
	Name          string      `json:"name"`
	FullName      string      `json:"full_name"`
	Private       bool        `json:"private"`
	HTMLURL       string      `json:"html_url"`
	CloneURL      string      `json:"clone_url"`
	DefaultBranch string      `json:"default_branch"`
	Permissions   *Permission `json:"permissions,omitempty"`
}

type Permission struct {
	Admin bool `json:"admin"`
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
}

func (f *Forgejo) GetRepoByID(id int64) (*Repository, *Response, error) {
	repo := new(Repository)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repositories/%d", id), nil, nil, repo)
	return repo, resp, err
}

func (f *Forgejo) GetRepo(owner, reponame string) (*Repository, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &reponame); err != nil {
		return nil, nil, err
	}
	repo := new(Repository)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s", owner, reponame), nil, nil, repo)
	return repo, resp, err
}

func (f *Forgejo) ListMyRepos(opt ListOptions) ([]*Repository, *Response, error) {
	opt.setDefaults()
	repos := make([]*Repository, 0, opt.PageSize)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/user/repos?%s", opt.getURLQuery().Encode()), nil, nil, &repos)
	return repos, resp, err
}

type GitEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
}

type GitTreeResponse struct {
	SHA        string     `json:"sha"`
	URL        string     `json:"url"`
	Entries    []GitEntry `json:"tree"`
	Truncated  bool       `json:"truncated"`
	Page       int        `json:"page"`
	TotalCount int        `json:"total_count"`
}

type GetTreeOptions struct {
	ListOptions

	Recursive bool
}

func (opt *GetTreeOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Recursive {
		query.Add("recursive", "1")
	}
	query.Add("per_page", fmt.Sprintf("%d", opt.PageSize))
	return query.Encode()
}

func (f *Forgejo) GetTrees(user, repo, ref string, opt GetTreeOptions) (*GitTreeResponse, error) {
	opt.setDefaults()
	if err := escapeValidatePathSegments(&user, &repo, &ref); err != nil {
		return nil, err
	}
	trees := new(GitTreeResponse)
	path := fmt.Sprintf("/repos/%s/%s/git/trees/%s?%s", user, repo, ref, opt.QueryEncode())
	_, err := f.getParsedResponse("GET", path, nil, nil, trees)
	return trees, err
}

func (f *Forgejo) ShaExists(user, repo, sha string) (bool, error) {
	status, _, err := f.getStatusCode("GET", fmt.Sprintf("/repos/%s/%s/git/commits/%s", user, repo, sha), nil, nil)
	if err != nil {
		return false, err
	}
	return status == 200, nil
}
