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

package server

import (
	"time"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/store"
)

// Syncer synces the user repository and permissions.
type Syncer interface {
	Sync(user *model.User) error
}

type syncer struct {
	remote remote.Remote
	store  store.Store
	perms  model.PermStore
	match  FilterFunc
}

// FilterFunc can be used to filter which repositories are
// synchronized with the local datastore.
type FilterFunc func(*model.Repo) bool

// NamespaceFilter
func NamespaceFilter(namespaces map[string]bool) FilterFunc {
	if namespaces == nil || len(namespaces) == 0 {
		return noopFilter
	}
	return func(repo *model.Repo) bool {
		if namespaces[repo.Owner] {
			return true
		} else {
			return false
		}
	}
}

// noopFilter is a filter function that always returns true.
func noopFilter(*model.Repo) bool {
	return true
}

// SetFilter sets the filter function.
func (s *syncer) SetFilter(fn FilterFunc) {
	s.match = fn
}

func (s *syncer) Sync(user *model.User, flatPermissions bool) error {
	unix := time.Now().Unix() - (3601) // force immediate expiration. note 1 hour expiration is hard coded at the moment
	repos, err := s.remote.Repos(user)
	if err != nil {
		return err
	}

	var remote []*model.Repo
	var perms []*model.Perm

	for _, repo := range repos {
		if s.match(repo) {
			remote = append(remote, repo)
			perm := model.Perm{
				UserID: user.ID,
				Repo:   repo.FullName,
				Pull:   true,
				Synced: unix,
			}
			// temporary workaround for v0.14.x to not hit api rate limits
			if flatPermissions {
				if repo.Perm != nil {
					perm.Push = repo.Perm.Push
					perm.Admin = repo.Perm.Admin
				} else {
					perm.Push = true
					perm.Admin = true
				}
			} else {
				remotePerm, err := s.remote.Perm(user, repo.Owner, repo.Name)
				if err == nil && remotePerm != nil {
					perm.Push = remotePerm.Push
					perm.Admin = remotePerm.Admin
				}
			}
			perms = append(perms, &perm)
		}
	}

	err = s.store.RepoBatch(remote)
	if err != nil {
		return err
	}

	err = s.store.PermBatch(perms)
	if err != nil {
		return err
	}

	// this is here as a precaution. I want to make sure that if an api
	// call to the version control system fails and (for some reason) returns
	// an empty list, we don't wipe out the user repository permissions.
	//
	// the side-effect of this code is that a user with 1 repository whose
	// access is removed will still display in the feed, but they will not
	// be able to access the actual repository data.
	if len(repos) == 0 {
		return nil
	}

	return s.perms.PermFlush(user, unix)
}
