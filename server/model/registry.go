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
	RegistryFind(*Repo, string) (*Registry, error)
	RegistryList(*Repo) ([]*Registry, error)
	RegistryCreate(*Repo, *Registry) error
	RegistryUpdate(*Repo, *Registry) error
	RegistryDelete(*Repo, string) error
}

// ReadOnlyRegistryService defines a service for managing registries.
type ReadOnlyRegistryService interface {
	RegistryFind(*Repo, string) (*Registry, error)
	RegistryList(*Repo) ([]*Registry, error)
}

// RegistryStore persists registry information to storage.
type RegistryStore interface {
	RegistryFind(*Repo, string) (*Registry, error)
	RegistryList(*Repo) ([]*Registry, error)
	RegistryCreate(*Registry) error
	RegistryUpdate(*Registry) error
	RegistryDelete(repo *Repo, addr string) error
}

// Registry represents a docker registry with credentials.
// swagger:model registry
type Registry struct {
	ID       int64  `json:"id"       xorm:"pk autoincr 'registry_id'"`
	RepoID   int64  `json:"-"        xorm:"UNIQUE(s) INDEX 'registry_repo_id'"`
	Address  string `json:"address"  xorm:"UNIQUE(s) INDEX 'registry_addr'"`
	Username string `json:"username" xorm:"varchar(2000) 'registry_username'"`
	Password string `json:"password" xorm:"TEXT 'registry_password'"`
	Token    string `json:"token"    xorm:"TEXT 'registry_token'"`         // TODO: deprecate
	Email    string `json:"email"    xorm:"varchar(500) 'registry_email'"` // TODO: deprecate
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
		Email:    r.Email,
		Token:    r.Token,
	}
}
