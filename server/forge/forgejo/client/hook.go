// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type Hook struct {
	ID      int64             `json:"id"`
	Type    string            `json:"type"`
	URL     string            `json:"-"`
	Config  map[string]string `json:"config"`
	Events  []string          `json:"events"`
	Active  bool              `json:"active"`
	Updated time.Time         `json:"updated_at"`
	Created time.Time         `json:"created_at"`
}

type CreateHookOption struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
	Events []string          `json:"events"`
	Active bool              `json:"active"`
}

func (f *Forgejo) CreateRepoHook(user, repo string, opt CreateHookOption) (*Hook, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	h := new(Hook)
	resp, err := f.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/hooks", user, repo), jsonHeader, bytes.NewReader(body), h)
	return h, resp, err
}

func (f *Forgejo) ListRepoHooks(user, repo string) ([]*Hook, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	hooks := make([]*Hook, 0, 10)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/hooks", user, repo), nil, nil, &hooks)
	return hooks, resp, err
}

func (f *Forgejo) DeleteRepoHook(user, repo string, id int64) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, err
	}
	_, resp, err := f.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/hooks/%d", user, repo, id), nil, nil)
	return resp, err
}
