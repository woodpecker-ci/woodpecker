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
	"net/url"
)

var (
	errRegistryAddressInvalid  = errors.New("invalid registry address")
	errRegistryUsernameInvalid = errors.New("invalid registry username")
	errRegistryPasswordInvalid = errors.New("invalid registry password")
)

// Registry represents a docker registry with credentials.
type Registry struct {
	ID       int64  `json:"id"       xorm:"pk autoincr 'id'"`
	RepoID   int64  `json:"-"        xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Address  string `json:"address"  xorm:"UNIQUE(s) INDEX 'address'"`
	Username string `json:"username" xorm:"varchar(2000) 'username'"`
	Password string `json:"password" xorm:"TEXT 'password'"`
} //	@name Registry

func (r Registry) TableName() string {
	return "registries"
}

// Validate validates the registry information.
func (r *Registry) Validate() error {
	switch {
	case len(r.Address) == 0:
		return errRegistryAddressInvalid
	case len(r.Username) == 0:
		return errRegistryUsernameInvalid
	case len(r.Password) == 0:
		return errRegistryPasswordInvalid
	}

	_, err := url.Parse(r.Address)
	return err
}

// Copy makes a copy of the registry without the password.
func (r *Registry) Copy() *Registry {
	return &Registry{
		ID:       r.ID,
		RepoID:   r.RepoID,
		Address:  r.Address,
		Username: r.Username,
	}
}
