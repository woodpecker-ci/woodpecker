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

package addon

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type argumentsAuth struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type argumentsRepo struct {
	U        *modelUser          `json:"u"`
	RemoteID model.ForgeRemoteID `json:"remote_id"`
	Owner    string              `json:"owner"`
	Name     string              `json:"name"`
}

type argumentsFileDir struct {
	U *modelUser      `json:"u"`
	R *modelRepo      `json:"r"`
	B *model.Pipeline `json:"b"`
	F string          `json:"f"`
}

type argumentsStatus struct {
	U *modelUser      `json:"u"`
	R *modelRepo      `json:"r"`
	B *model.Pipeline `json:"b"`
	P *model.Workflow `json:"p"`
}

type argumentsNetrc struct {
	U *modelUser `json:"u"`
	R *modelRepo `json:"r"`
}

type argumentsActivateDeactivate struct {
	U    *modelUser `json:"u"`
	R    *modelRepo `json:"r"`
	Link string     `json:"link"`
}

type argumentsBranchesPullRequests struct {
	U *modelUser         `json:"u"`
	R *modelRepo         `json:"r"`
	P *model.ListOptions `json:"p"`
}

type argumentsBranchHead struct {
	U      *modelUser `json:"u"`
	R      *modelRepo `json:"r"`
	Branch string     `json:"branch"`
}

type argumentsOrgMembershipOrg struct {
	U   *modelUser `json:"u"`
	Org string     `json:"org"`
}

type responseHook struct {
	Repo     *modelRepo      `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
}

type responseLogin struct {
	User        *modelUser `json:"user"`
	RedirectURL string     `json:"redirect_url"`
}

type httpRequest struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header"`
	Form   map[string][]string `json:"form"`
	Body   []byte              `json:"body"`
}

// modelUser is an extension of model.User to marshal all fields to JSON.
type modelUser struct {
	User *model.User `json:"user"`

	ForgeRemoteID model.ForgeRemoteID `json:"forge_remote_id"`

	// Token is the oauth2 token.
	Token string `json:"token"`

	// Secret is the oauth2 token secret.
	Secret string `json:"secret"`

	// Expiry is the token and secret expiration timestamp.
	Expiry int64 `json:"expiry"`

	// Hash is a unique token used to sign tokens.
	Hash string `json:"hash"`
}

func (m *modelUser) asModel() *model.User {
	m.User.ForgeRemoteID = m.ForgeRemoteID
	m.User.AccessToken = m.Token
	m.User.RefreshToken = m.Secret
	m.User.Expiry = m.Expiry
	m.User.Hash = m.Hash
	return m.User
}

func modelUserFromModel(u *model.User) *modelUser {
	return &modelUser{
		User:          u,
		ForgeRemoteID: u.ForgeRemoteID,
		Token:         u.AccessToken,
		Secret:        u.RefreshToken,
		Expiry:        u.Expiry,
		Hash:          u.Hash,
	}
}

// modelRepo is an extension of model.Repo to marshal all fields to JSON.
type modelRepo struct {
	Repo   *model.Repo `json:"repo"`
	UserID int64       `json:"user_id"`
	Hash   string      `json:"hash"`
	Perm   *model.Perm `json:"perm"`
}

func (m *modelRepo) asModel() *model.Repo {
	m.Repo.UserID = m.UserID
	m.Repo.Hash = m.Hash
	m.Repo.Perm = m.Perm
	return m.Repo
}

func modelRepoFromModel(r *model.Repo) *modelRepo {
	return &modelRepo{
		Repo:   r,
		UserID: r.UserID,
		Hash:   r.Hash,
		Perm:   r.Perm,
	}
}
