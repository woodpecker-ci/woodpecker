// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

type User struct {
	UserName   string      `json:"login"`
	Email      string      `json:"email"`
	AvatarURL  string      `json:"avatar_url"`
	Visibility VisibleType `json:"visibility"`
}

func (f *Forgejo) GetMyUserInfo() (*User, *Response, error) {
	u := new(User)
	resp, err := f.getParsedResponse("GET", "/user", nil, nil, u)
	return u, resp, err
}
