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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestMembershipCacheForgeScoping(t *testing.T) {
	cache := NewMembershipService(nil)

	// two distinct users on two forges that happen to share the same
	// forge-local remote id
	userForge1 := &model.User{ID: 1, Login: "alice", ForgeID: 1, ForgeRemoteID: "42"}
	userForge2 := &model.User{ID: 2, Login: "bob", ForgeID: 2, ForgeRemoteID: "42"}

	forge1 := forge_mocks.NewMockForge(t)
	forge1.On("OrgMembership", mock.Anything, userForge1, "acme").
		Return(&model.OrgPerm{Member: true, Admin: true}, nil).Once()

	forge2 := forge_mocks.NewMockForge(t)
	forge2.On("OrgMembership", mock.Anything, userForge2, "acme").
		Return(&model.OrgPerm{}, nil).Once()

	perm1, err := cache.Get(t.Context(), forge1, userForge1, "acme")
	require.NoError(t, err)
	assert.True(t, perm1.Admin)

	// must NOT be served from userForge1's cache entry
	perm2, err := cache.Get(t.Context(), forge2, userForge2, "acme")
	require.NoError(t, err)
	assert.False(t, perm2.Member, "membership of a same-remote-id user on another forge leaked from the cache")
	assert.False(t, perm2.Admin)

	// repeated lookups are served from the per-user cache entries (the
	// .Once() expectations above fail the test on a second forge call)
	perm1again, err := cache.Get(t.Context(), forge1, userForge1, "acme")
	require.NoError(t, err)
	assert.True(t, perm1again.Admin)

	perm2again, err := cache.Get(t.Context(), forge2, userForge2, "acme")
	require.NoError(t, err)
	assert.False(t, perm2again.Member)
}
