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

package hashicorp

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
	R *model.Repo     `json:"r"`
	B *model.Pipeline `json:"b"`
	F string          `json:"f"`
}

type argumentsStatus struct {
	U *modelUser      `json:"u"`
	R *model.Repo     `json:"r"`
	B *model.Pipeline `json:"b"`
	P *model.Workflow `json:"p"`
}

type argumentsNetrc struct {
	U *modelUser  `json:"u"`
	R *model.Repo `json:"r"`
}

type argumentsActivateDeactivate struct {
	U    *modelUser  `json:"u"`
	R    *model.Repo `json:"r"`
	Link string      `json:"link"`
}

type argumentsBranchesPullRequests struct {
	U *modelUser         `json:"u"`
	R *model.Repo        `json:"r"`
	P *model.ListOptions `json:"p"`
}

type argumentsBranchHead struct {
	U      *modelUser  `json:"u"`
	R      *model.Repo `json:"r"`
	Branch string      `json:"branch"`
}

type argumentsOrgMembershipOrg struct {
	U   *modelUser `json:"u"`
	Org string     `json:"org"`
}

type responseHook struct {
	Repo     *model.Repo     `json:"repo"`
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

// modelUser is a model.User, but all fields are marshaled to JSON
type modelUser struct {
	// the id for this user.
	ID int64 `json:"id"`

	ForgeRemoteID model.ForgeRemoteID `json:"forge_remote_id"`

	// Login is the username for this user.
	Login string `json:"login"`

	// Token is the oauth2 token.
	Token string `json:"token"`

	// Secret is the oauth2 token secret.
	Secret string `json:"secret"`

	// Expiry is the token and secret expiration timestamp.
	Expiry int64 `json:"expiry"`

	// Email is the email address for this user.
	Email string `json:"email"`

	// the avatar url for this user.
	Avatar string `json:"avatar_url"`

	// Admin indicates the user is a system administrator.
	Admin bool `json:"admin"`

	// Hash is a unique token used to sign tokens.
	Hash string `json:"hash"`

	// OrgID is the of the user as model.Org.
	OrgID int64 `json:"org_id"`
}

func (m *modelUser) asModel() *model.User {
	return &model.User{
		ID:            m.ID,
		ForgeRemoteID: m.ForgeRemoteID,
		Login:         m.Login,
		Token:         m.Token,
		Secret:        m.Secret,
		Expiry:        m.Expiry,
		Email:         m.Email,
		Avatar:        m.Avatar,
		Admin:         m.Admin,
		Hash:          m.Hash,
		OrgID:         m.OrgID,
	}
}

func modelUserFromModel(u *model.User) *modelUser {
	return &modelUser{
		ID:            u.ID,
		ForgeRemoteID: u.ForgeRemoteID,
		Login:         u.Login,
		Token:         u.Token,
		Secret:        u.Secret,
		Expiry:        u.Expiry,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Admin:         u.Admin,
		Hash:          u.Hash,
		OrgID:         u.OrgID,
	}
}
