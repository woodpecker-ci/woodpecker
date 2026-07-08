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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
)

// fakeMembership is a canned cache.MembershipService. It records whether it
// was consulted so tests can assert the forge was (not) asked.
type fakeMembership struct {
	perm   *model.OrgPerm
	called bool
}

func (f *fakeMembership) Get(_ context.Context, _ forge.Forge, _ *model.User, _ string) (*model.OrgPerm, error) {
	f.called = true
	if f.perm == nil {
		return &model.OrgPerm{}, nil
	}
	return f.perm, nil
}

// installOrgForgeManager wires a mock manager whose ForgeFromUser returns a
// bare mock forge (no expectations; membership is faked separately).
func installOrgForgeManager(t *testing.T) {
	t.Helper()
	mgr := manager_mocks.NewMockManager(t)
	_forge := forge_mocks.NewMockForge(t)
	mgr.On("ForgeFromUser", mock.Anything).Return(_forge, nil).Maybe()
	server.Config.Services.Manager = mgr
}

func TestGetOrgPermissions(t *testing.T) {
	s := newTestStore(t)

	t.Run("anonymous user gets empty permissions without forge lookup", func(t *testing.T) {
		// no ForgeFromUser expectation: mock manager fails the test if the
		// handler dereferences the nil user to resolve a forge
		mgr := manager_mocks.NewMockManager(t)
		server.Config.Services.Manager = mgr
		membership := &fakeMembership{}
		server.Config.Services.Membership = membership

		tc := newTestContext(t, s)
		org := &model.Org{ID: 1, Name: "some-org", ForgeID: 1}
		tc.Ctx.Set("org", org)

		GetOrgPermissions(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		perm := new(model.OrgPerm)
		tc.decodeJSON(t, perm)
		assert.False(t, perm.Member)
		assert.False(t, perm.Admin)
		assert.False(t, membership.called)
	})

	t.Run("same login on another forge gets no admin on foreign user-org", func(t *testing.T) {
		installOrgForgeManager(t)
		membership := &fakeMembership{}
		server.Config.Services.Membership = membership

		tc := newTestContext(t, s)
		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 2, Name: "alice", ForgeID: 2, IsUser: true}
		withUser(user)(tc)
		tc.Ctx.Set("org", org)

		GetOrgPermissions(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		perm := new(model.OrgPerm)
		tc.decodeJSON(t, perm)
		assert.False(t, perm.Member)
		assert.False(t, perm.Admin)
	})

	t.Run("membership of same-named org on user's forge does not leak to foreign org", func(t *testing.T) {
		installOrgForgeManager(t)
		// the user IS an admin of org "acme" on their own forge (1) ...
		membership := &fakeMembership{perm: &model.OrgPerm{Member: true, Admin: true}}
		server.Config.Services.Membership = membership

		tc := newTestContext(t, s)
		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		// ... but this org "acme" lives on forge 2
		org := &model.Org{ID: 3, Name: "acme", ForgeID: 2}
		withUser(user)(tc)
		tc.Ctx.Set("org", org)

		GetOrgPermissions(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		perm := new(model.OrgPerm)
		tc.decodeJSON(t, perm)
		assert.False(t, perm.Member, "membership on forge 1 must not apply to org of forge 2")
		assert.False(t, perm.Admin)
		assert.False(t, membership.called, "forge of the user must not be asked about a foreign forge's org")
	})

	t.Run("own user-org on same forge grants admin", func(t *testing.T) {
		installOrgForgeManager(t)
		membership := &fakeMembership{}
		server.Config.Services.Membership = membership

		tc := newTestContext(t, s)
		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 4, Name: "alice", ForgeID: 1, IsUser: true}
		withUser(user)(tc)
		tc.Ctx.Set("org", org)

		GetOrgPermissions(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		perm := new(model.OrgPerm)
		tc.decodeJSON(t, perm)
		assert.True(t, perm.Member)
		assert.True(t, perm.Admin)
	})

	t.Run("membership check runs for org on user's own forge", func(t *testing.T) {
		installOrgForgeManager(t)
		membership := &fakeMembership{perm: &model.OrgPerm{Member: true}}
		server.Config.Services.Membership = membership

		tc := newTestContext(t, s)
		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 5, Name: "acme", ForgeID: 1}
		withUser(user)(tc)
		tc.Ctx.Set("org", org)

		GetOrgPermissions(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		perm := new(model.OrgPerm)
		tc.decodeJSON(t, perm)
		assert.True(t, perm.Member)
		assert.True(t, membership.called)
	})
}
