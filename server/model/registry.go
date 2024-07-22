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
	OrgID    int64  `json:"org_id"   xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID   int64  `json:"repo_id"  xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Address  string `json:"address"  xorm:"NOT NULL UNIQUE(s) INDEX 'address'"`
	Username string `json:"username" xorm:"varchar(2000) 'username'"`
	Password string `json:"password" xorm:"TEXT 'password'"`
	ReadOnly bool   `json:"readonly" xorm:"-"`
} //	@name Registry

func (r Registry) TableName() string {
	return "registries"
}

// Global registry.
func (r Registry) IsGlobal() bool {
	return r.RepoID == 0 && r.OrgID == 0
}

// Organization registry.
func (r Registry) IsOrganization() bool {
	return r.RepoID == 0 && r.OrgID != 0
}

// Repository registry.
func (r Registry) IsRepository() bool {
	return r.RepoID != 0 && r.OrgID == 0
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
		OrgID:    r.OrgID,
		RepoID:   r.RepoID,
		Address:  r.Address,
		Username: r.Username,
		ReadOnly: r.ReadOnly,
	}
}
