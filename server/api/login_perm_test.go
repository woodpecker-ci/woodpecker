// Copyright 2026 Woodpecker Authors
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

//go:build test

package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func TestUpdateRepoPermissionsOrphanedRepos(t *testing.T) {
	s := newTestStore(t)
	const forgeID = 1

	user := &model.User{Login: "alice", ForgeID: forgeID}
	require.NoError(t, s.CreateUser(user))

	newRepo := func(name string, userID int64, orphaned bool) *model.Repo {
		repo := &model.Repo{
			UserID: userID, ForgeID: forgeID, ForgeRemoteID: model.ForgeRemoteID(name),
			Owner: "acme", Name: name, FullName: "acme/" + name, IsActive: true, HasNoForgeRepo: orphaned,
		}
		require.NoError(t, s.CreateRepo(repo))
		require.NoError(t, s.PermUpsert(&model.Perm{UserID: user.ID, RepoID: repo.ID, Pull: true, Push: true, Admin: true}))
		return repo
	}

	listed := newRepo("listed", user.ID, true)
	deleted := newRepo("deleted", user.ID, false)
	foreign := newRepo("foreign", 99, false)
	unreachable := newRepo("unreachable", user.ID, false)

	_forge := forge_mocks.NewMockForge(t)
	_forge.On("Repos", mock.Anything, user, &model.ListOptions{Page: 1, PerPage: perPage}).Return([]*model.Repo{
		{ForgeRemoteID: listed.ForgeRemoteID, Owner: "acme", Name: "listed", Perm: &model.Perm{Pull: true, Push: true, Admin: true}},
	}, nil).Once()
	_forge.On("Repos", mock.Anything, user, &model.ListOptions{Page: 2, PerPage: perPage}).Return([]*model.Repo{}, nil).Once()
	_forge.On("Repo", mock.Anything, user, deleted.ForgeRemoteID, deleted.Owner, deleted.Name).
		Return(nil, forge_types.ErrRepoNotFound).Once()
	_forge.On("Repo", mock.Anything, user, unreachable.ForgeRemoteID, unreachable.Owner, unreachable.Name).
		Return(nil, errors.New("forge not reachable")).Once()

	tc := newTestContext(t, s)
	require.NoError(t, updateRepoPermissions(tc.Ctx, user, s, _forge, forgeID))

	assertOrphaned := func(repo *model.Repo, orphaned bool) {
		stored, err := s.GetRepo(repo.ID)
		require.NoError(t, err)
		assert.Equal(t, orphaned, stored.HasNoForgeRepo, repo.Name)
	}
	assertPerm := func(repo *model.Repo, kept bool) {
		_, err := s.PermFind(user, repo)
		if kept {
			assert.NoError(t, err, repo.Name)
		} else {
			assert.ErrorIs(t, err, types.ErrRecordNotExist, repo.Name)
		}
	}

	// repo is back at the forge
	assertOrphaned(listed, false)
	assertPerm(listed, true)

	// activated repo got deleted at the forge
	assertOrphaned(deleted, true)
	assertPerm(deleted, true)

	// repo activated by another user is pruned as before
	assertOrphaned(foreign, false)
	assertPerm(foreign, false)

	// forge errors don't drop access
	assertOrphaned(unreachable, false)
	assertPerm(unreachable, true)
}
