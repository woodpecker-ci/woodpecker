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
	errRegistryAddressInvalid  = errors.New("Invalid Registry Address")
	errRegistryUsernameInvalid = errors.New("Invalid Registry Username")
	errRegistryPasswordInvalid = errors.New("Invalid Registry Password")
)

// RegistryService defines a service for managing registries.
type RegistryService interface {
	RegistryListPipeline(*Repo, *Pipeline) ([]*Registry, error)
	// Repository registries
	RegistryFind(*Repo, string) (*Registry, error)
	RegistryList(*Repo) ([]*Registry, error)
	RegistryCreate(*Repo, *Registry) error
	RegistryUpdate(*Repo, *Registry) error
	RegistryDelete(*Repo, string) error
	// Organization registries
	OrgRegistryFind(string, string) (*Registry, error)
	OrgRegistryList(string) ([]*Registry, error)
	OrgRegistryCreate(string, *Registry) error
	OrgRegistryUpdate(string, *Registry) error
	OrgRegistryDelete(string, string) error
	// Global registries
	GlobalRegistryFind(string) (*Registry, error)
	GlobalRegistryList() ([]*Registry, error)
	GlobalRegistryCreate(*Registry) error
	GlobalRegistryUpdate(*Registry) error
	GlobalRegistryDelete(string) error
}

// ReadOnlyRegistryService defines a service for managing registries.
type ReadOnlyRegistryService interface {
	RegistryFind(string) (*Registry, error)
	RegistryList() ([]*Registry, error)
}

// RegistryStore persists registry information to storage.
type RegistryStore interface {
	RegistryFind(*Repo, string) (*Registry, error)
	RegistryList(*Repo, bool) ([]*Registry, error)
	RegistryCreate(*Registry) error
	RegistryUpdate(*Registry) error
	RegistryDelete(*Registry) error
	OrgRegistryFind(string, string) (*Registry, error)
	OrgRegistryList(string) ([]*Registry, error)
	GlobalRegistryFind(string) (*Registry, error)
	GlobalRegistryList() ([]*Registry, error)
}

// Registry represents a docker registry with credentials.
// swagger:model registry
type Registry struct {
	ID       int64  `json:"id"       xorm:"pk autoincr 'registry_id'"`
	Owner    string `json:"-"        xorm:"NOT NULL DEFAULT '' UNIQUE(s) INDEX 'registry_owner'"`
	RepoID   int64  `json:"-"        xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'registry_repo_id'"`
	Address  string `json:"address"  xorm:"NOT NULL UNIQUE(s) INDEX 'registry_addr'"`
	Username string `json:"username" xorm:"varchar(2000) 'registry_username'"`
	Password string `json:"password" xorm:"TEXT 'registry_password'"`
	Token    string `json:"token"    xorm:"TEXT 'registry_token'"`
	Email    string `json:"email"    xorm:"varchar(500) 'registry_email'"`
}

// Global registry.
func (r Registry) Global() bool {
	return r.RepoID == 0 && r.Owner == ""
}

// Organization registry.
func (r Registry) Organization() bool {
	return r.RepoID == 0 && r.Owner != ""
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
		Owner:    r.Owner,
		RepoID:   r.RepoID,
		Address:  r.Address,
		Username: r.Username,
		Email:    r.Email,
		Token:    r.Token,
	}
}
