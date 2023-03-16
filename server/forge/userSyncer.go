// Copyright 2022 Woodpecker Authors
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

package forge

import (
	"context"
	"fmt"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// UserSyncer syncs the user repository and permissions.
type UserSyncer interface {
	Sync(ctx context.Context, user *model.User) error
}

type Syncer struct {
	Forge Forge
	Store store.Store
	Perms model.PermStore
	Match FilterFunc
}

// FilterFunc can be used to filter which repositories are
// synchronized with the local datastore.
type FilterFunc func(*model.Repo) bool

func NamespaceFilter(namespaces map[string]bool) FilterFunc {
	if len(namespaces) == 0 {
		return noopFilter
	}
	return func(repo *model.Repo) bool {
		return namespaces[repo.Owner]
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

func (s *Syncer) Sync(ctx context.Context, user *model.User, flatPermissions bool) error {
	unix := time.Now().Unix() - (3601) // force immediate expiration. note 1 hour expiration is hard coded at the moment
	repos, err := s.Forge.Repos(ctx, user)
	if err != nil {
		return err
	}

	forgeRepos := make([]*model.Repo, 0, len(repos))
	for _, repo := range repos {
		if s.Match(repo) {
			repo.Perm = &model.Perm{
				UserID: user.ID,
				RepoID: repo.ID,
				Repo:   repo,
				Synced: unix,
			}

			// TODO(485) temporary workaround to not hit api rate limits
			if flatPermissions {
				repo.Perm.Pull = true
				repo.Perm.Push = true
				repo.Perm.Admin = true
			} else {
				forgePerm, err := s.Forge.Perm(ctx, user, repo)
				if err != nil {
					return fmt.Errorf("could not fetch permission of repo '%s': %w", repo.FullName, err)
				}
				repo.Perm.Pull = forgePerm.Pull
				repo.Perm.Push = forgePerm.Push
				repo.Perm.Admin = forgePerm.Admin
			}

			forgeRepos = append(forgeRepos, repo)
		}
	}

	err = s.Store.RepoBatch(forgeRepos)
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
