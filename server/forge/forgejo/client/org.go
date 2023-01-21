// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
)

type Organization struct {
	UserName  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type OrgPermissions struct {
	// CanCreateRepository bool `json:"can_create_repository"`
	// CanRead             bool `json:"can_read"`
	// CanWrite            bool `json:"can_write"`
	IsAdmin bool `json:"is_admin"`
	IsOwner bool `json:"is_owner"`
}

func (f *Forgejo) ListMyOrgs(opt ListOptions) ([]*Organization, *Response, error) {
	opt.setDefaults()
	orgs := make([]*Organization, 0, opt.PageSize)
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/user/orgs?%s", opt.getURLQuery().Encode()), nil, nil, &orgs)
	return orgs, resp, err
}

func (f *Forgejo) CheckOrgMembership(org, user string) (bool, *Response, error) {
	if err := escapeValidatePathSegments(&org, &user); err != nil {
		return false, nil, err
	}
	status, resp, err := f.getStatusCode("GET", fmt.Sprintf("/orgs/%s/members/%s", org, user), nil, nil)
	if err != nil {
		return false, resp, err
	}
	switch status {
	case http.StatusNoContent:
		return true, resp, nil
	case http.StatusNotFound:
		return false, resp, nil
	default:
		return false, resp, fmt.Errorf("unexpected Status: %d", status)
	}
}

func (f *Forgejo) GetOrgPermissions(org, user string) (*OrgPermissions, *Response, error) {
	if err := escapeValidatePathSegments(&org, &user); err != nil {
		return nil, nil, err
	}

	perm := &OrgPermissions{}
	resp, err := f.getParsedResponse("GET", fmt.Sprintf("/users/%s/orgs/%s/permissions", user, org), jsonHeader, nil, &perm)
	if err != nil {
		return nil, resp, err
	}
	return perm, resp, nil
}
