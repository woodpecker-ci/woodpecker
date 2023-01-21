// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

type StatusState string

const (
	StatusPending StatusState = "pending"
	StatusSuccess StatusState = "success"
	StatusError   StatusState = "error"
	StatusFailure StatusState = "failure"
	StatusWarning StatusState = "warning"
)

type Status struct{}

type CreateStatusOption struct {
	State       StatusState `json:"state"`
	TargetURL   string      `json:"target_url"`
	Description string      `json:"description"`
	Context     string      `json:"context"`
}

func (f *Forgejo) CreateStatus(owner, repo, sha string, opts CreateStatusOption) (*Status, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opts)
	if err != nil {
		return nil, nil, err
	}
	status := new(Status)
	resp, err := f.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/statuses/%s", owner, repo, url.QueryEscape(sha)), jsonHeader, bytes.NewReader(body), status)
	return status, resp, err
}
