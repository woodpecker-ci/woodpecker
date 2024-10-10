// Copyright 2021 Woodpecker Authors
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

package model

import (
	"errors"
	"regexp"
)

// Validate a username (e.g. from github).
var reUsername = regexp.MustCompile("^[a-zA-Z0-9-_.]+$")

var errUserLoginInvalid = errors.New("invalid user login")

const maxLoginLen = 250

// User represents a registered user.
type User struct {
	// the id for this user.
	//
	// required: true
	ID int64 `json:"id" xorm:"pk autoincr 'id'"`

	ForgeID int64 `json:"forge_id,omitempty" xorm:"forge_id"`

	ForgeRemoteID ForgeRemoteID `json:"-" xorm:"forge_remote_id"`

	// Login is the username for this user.
	//
	// required: true
	Login string `json:"login"  xorm:"UNIQUE 'login'"`

	// AccessToken is the oauth2 access token.
	AccessToken string `json:"-"  xorm:"TEXT 'token'"`

	// RefreshToken is the oauth2 refresh token.
	RefreshToken string `json:"-" xorm:"TEXT 'secret'"`

	// Expiry is the AccessToken expiration timestamp (unix seconds).
	Expiry int64 `json:"-" xorm:"expiry"`

	// Email is the email address for this user.
	//
	// required: true
	Email string `json:"email" xorm:" varchar(500) 'email'"`

	// the avatar url for this user.
	Avatar string `json:"avatar_url" xorm:" varchar(500) 'avatar'"`

	// Admin indicates the user is a system administrator.
	//
	// NOTE: If the username is part of the WOODPECKER_ADMIN
	// environment variable, this value will be set to true on login.
	Admin bool `json:"admin,omitempty" xorm:"admin"`

	// Hash is a unique token used to sign tokens.
	Hash string `json:"-" xorm:"UNIQUE varchar(500) 'hash'"`

	// OrgID is the of the user as model.Org.
	OrgID int64 `json:"org_id" xorm:"org_id"`
} //	@name User

// TableName return database table name for xorm.
func (User) TableName() string {
	return "users"
}

// Validate validates the required fields and formats.
func (u *User) Validate() error {
	switch {
	case len(u.Login) == 0:
		return errUserLoginInvalid
	case len(u.Login) > maxLoginLen:
		return errUserLoginInvalid
	case !reUsername.MatchString(u.Login):
		return errUserLoginInvalid
	default:
		return nil
	}
}
