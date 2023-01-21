// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
)

type PRBranchInfo struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

type PullRequest struct {
	URL    string `json:"url"`
	Poster *User  `json:"user"`
	Title  string `json:"title"`
	State  string `json:"state"`

	Base *PRBranchInfo `json:"base"`
	Head *PRBranchInfo `json:"head"`
}

type ChangedFile struct {
	Filename string `json:"filename"`
}

func (f *Forgejo) ListPullRequestFiles(owner, repo string, index int64, opt ListOptions) ([]*ChangedFile, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	files := make([]*ChangedFile, 0, opt.PageSize)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/pulls/%d/files?%s", owner, repo, index, opt.getURLQuery().Encode()), nil, nil, &files)
	return files, resp, err
}
