// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
)

// PayloadUser represents the author or committer of a commit
type PayloadUser struct {
	// Full name of the commit author
	Name     string `json:"name"`
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// PayloadCommit represents a commit
type PayloadCommit struct {
	// sha1 hash of the commit
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
	// Author       *PayloadUser               `json:"author"`
	// Committer    *PayloadUser               `json:"committer"`
	// Verification *PayloadCommitVerification `json:"verification"`
	// Timestamp    time.Time                  `json:"timestamp"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

type Branch struct {
	Name   string         `json:"name"`
	Commit *PayloadCommit `json:"commit"`
}

func (f *Forgejo) ListRepoBranches(user, repo string, opt ListOptions) ([]*Branch, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	branches := make([]*Branch, 0, opt.PageSize)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/branches?%s", user, repo, opt.getURLQuery().Encode()), nil, nil, &branches)
	return branches, resp, err
}

func (f *Forgejo) GetRepoBranch(user, repo, branch string) (*Branch, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo, &branch); err != nil {
		return nil, nil, err
	}
	b := new(Branch)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/branches/%s", user, repo, branch), nil, nil, &b)
	return b, resp, err
}
