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

package shared

import (
	"time"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/store"
)

// UserSyncer syncs the user repository and permissions.
type UserSyncer interface {
	Sync(user *model.User) error
}

type Syncer struct {
	Remote remote.Remote
	Store  store.Store
	Perms  model.PermStore
	Match  FilterFunc
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
func (s *Syncer) SetFilter(fn FilterFunc) {
	s.Match = fn
}

func (s *Syncer) Sync(user *model.User) error {
	unix := time.Now().Unix() - (3601) // force immediate expiration. note 1 hour expiration is hard coded at the moment
	repos, err := s.Remote.Repos(user)
	if err != nil {
		return err
	}

	var remote []*model.Repo
	var perms []*model.Perm

	for _, repo := range repos {
		if s.Match(repo) {
			remote = append(remote, repo)
			perm := model.Perm{
				UserID: user.ID,
				Repo:   repo.FullName,
				Pull:   true,
				Synced: unix,
			}
			if repo.Perm != nil {
				perm.Push = repo.Perm.Push
				perm.Admin = repo.Perm.Admin
			}
			perms = append(perms, &perm)
		}
	}

	err = s.Store.RepoBatch(remote)
	if err != nil {
		return err
	}

	err = s.Store.PermBatch(perms)
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

	return s.Perms.PermFlush(user, unix)
}
