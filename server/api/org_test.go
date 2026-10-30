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
		user := &model.User{ID: 1, Login: "alice", ForgeID: 1, OrgID: 4}
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

func TestGetOrgPermissionsPersonalOrg(t *testing.T) {
	s := newTestStore(t)
	tests := []struct {
		name  string
		user  model.User
		org   model.Org
		allow bool
	}{
		{
			name:  "different casing on linked org",
			user:  model.User{Login: "Alice", ForgeID: 1, OrgID: 10},
			org:   model.Org{ID: 10, Name: "alice", ForgeID: 1, IsUser: true},
			allow: true,
		},
		{
			name:  "renamed user keeps linked org",
			user:  model.User{Login: "new-name", ForgeID: 1, OrgID: 10},
			org:   model.Org{ID: 10, Name: "old-name", ForgeID: 1, IsUser: true},
			allow: true,
		},
		{
			name: "same name does not grant another personal org",
			user: model.User{Login: "alice", ForgeID: 1, OrgID: 20},
			org:  model.Org{ID: 10, Name: "alice", ForgeID: 1, IsUser: true},
		},
		{
			name: "missing link does not fall back to name",
			user: model.User{Login: "alice", ForgeID: 1},
			org:  model.Org{ID: 10, Name: "alice", ForgeID: 1, IsUser: true},
		},
		{
			name: "invalid cross-forge link is denied",
			user: model.User{Login: "alice", ForgeID: 1, OrgID: 10},
			org:  model.Org{ID: 10, Name: "alice", ForgeID: 2, IsUser: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installOrgForgeManager(t)
			membership := &fakeMembership{perm: &model.OrgPerm{Member: true, Admin: true}}
			server.Config.Services.Membership = membership
			tc := newTestContext(t, s)
			withUser(&tt.user)(tc)
			tc.Ctx.Set("org", &tt.org)

			GetOrgPermissions(tc.Ctx)

			require.Equal(t, http.StatusOK, tc.Recorder.Code)
			perm := new(model.OrgPerm)
			tc.decodeJSON(t, perm)
			assert.Equal(t, tt.allow, perm.Member)
			assert.Equal(t, tt.allow, perm.Admin)
			assert.False(t, membership.called, "personal ownership must not use forge membership")
		})
	}
}

func TestLookupOrgAccess(t *testing.T) {
	tests := []struct {
		name            string
		user            model.User
		org             model.Org
		linked          bool
		anonymous       bool
		perm            *model.OrgPerm
		status          int
		checkMembership bool
	}{
		{
			name:   "private personal org with different casing",
			user:   model.User{Login: "Alice", ForgeID: 1},
			org:    model.Org{Name: "alice", ForgeID: 1, IsUser: true, Private: true},
			linked: true,
			status: http.StatusOK,
		},
		{
			name:   "private personal org after rename",
			user:   model.User{Login: "new-name", ForgeID: 1},
			org:    model.Org{Name: "old-name", ForgeID: 1, IsUser: true, Private: true},
			linked: true,
			status: http.StatusOK,
		},
		{
			name:   "same name does not expose another private personal org",
			user:   model.User{Login: "alice", ForgeID: 1, OrgID: 999},
			org:    model.Org{Name: "alice", ForgeID: 1, IsUser: true, Private: true},
			perm:   &model.OrgPerm{Member: true, Admin: true},
			status: http.StatusNotFound,
		},
		{
			name:   "missing link does not expose private personal org by name",
			user:   model.User{Login: "alice", ForgeID: 1},
			org:    model.Org{Name: "alice", ForgeID: 1, IsUser: true, Private: true},
			perm:   &model.OrgPerm{Member: true, Admin: true},
			status: http.StatusNotFound,
		},
		{
			name:   "instance admin can look up private org",
			user:   model.User{Login: "root", ForgeID: 1, Admin: true},
			org:    model.Org{Name: "alice", ForgeID: 1, IsUser: true, Private: true},
			status: http.StatusOK,
		},
		{
			name:            "private ordinary org requires forge membership",
			user:            model.User{Login: "alice", ForgeID: 1},
			org:             model.Org{Name: "acme", ForgeID: 1, Private: true},
			perm:            &model.OrgPerm{Member: true},
			status:          http.StatusOK,
			checkMembership: true,
		},
		{
			name:            "nonmember cannot look up private ordinary org",
			user:            model.User{Login: "alice", ForgeID: 1},
			org:             model.Org{Name: "acme", ForgeID: 1, Private: true},
			status:          http.StatusNotFound,
			checkMembership: true,
		},
		{
			name:            "stale link to ordinary org does not bypass membership",
			user:            model.User{Login: "alice", ForgeID: 1},
			org:             model.Org{Name: "alice", ForgeID: 1, Private: true},
			linked:          true,
			status:          http.StatusNotFound,
			checkMembership: true,
		},
		{
			name:      "anonymous user cannot look up private org",
			org:       model.Org{Name: "alice", ForgeID: 1, IsUser: true, Private: true},
			anonymous: true,
			status:    http.StatusNotFound,
		},
		{
			name:   "public personal org remains visible to other users",
			user:   model.User{Login: "bob", ForgeID: 1},
			org:    model.Org{Name: "alice", ForgeID: 1, IsUser: true},
			status: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			require.NoError(t, s.OrgCreate(&tt.org))
			if tt.linked {
				tt.user.OrgID = tt.org.ID
			}
			installOrgForgeManager(t)
			membership := &fakeMembership{perm: tt.perm}
			server.Config.Services.Membership = membership
			tc := newTestContext(t, s)
			if !tt.anonymous {
				withUser(&tt.user)(tc)
			}
			withParam("org_full_name", "/"+tt.org.Name)(tc)

			LookupOrg(tc.Ctx)

			require.Equal(t, tt.status, tc.Recorder.Code)
			assert.Equal(t, tt.checkMembership, membership.called)
			if tt.status == http.StatusOK {
				got := new(model.Org)
				tc.decodeJSON(t, got)
				assert.Equal(t, tt.org.ID, got.ID)
			} else {
				assert.Empty(t, tc.Recorder.Body.String(), "denied lookup must not expose org information")
			}
		})
	}
}

func TestPatchOrg(t *testing.T) {
	s := newTestStore(t)

	org := &model.Org{Name: "acme", ForgeID: 1}
	require.NoError(t, s.OrgCreate(org))

	patch := func(t *testing.T, body any) *testContext {
		t.Helper()
		stored, err := s.OrgGet(org.ID)
		require.NoError(t, err)

		tc := newTestContext(t, s)
		withRequest(http.MethodPatch, body)(tc)
		tc.Ctx.Set("org", stored)

		PatchOrg(tc.Ctx)
		return tc
	}

	t.Run("enables require org agents", func(t *testing.T) {
		requireOrgAgents := true
		tc := patch(t, &model.OrgPatch{RequireOrgAgents: &requireOrgAgents})

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		got := new(model.Org)
		tc.decodeJSON(t, got)
		assert.True(t, got.RequireOrgAgents)

		stored, err := s.OrgGet(org.ID)
		require.NoError(t, err)
		assert.True(t, stored.RequireOrgAgents)
		assert.Equal(t, "acme", stored.Name)
	})

	t.Run("empty patch keeps the setting", func(t *testing.T) {
		tc := patch(t, make(map[string]any))

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		stored, err := s.OrgGet(org.ID)
		require.NoError(t, err)
		assert.True(t, stored.RequireOrgAgents)
	})

	t.Run("disables require org agents", func(t *testing.T) {
		requireOrgAgents := false
		tc := patch(t, &model.OrgPatch{RequireOrgAgents: &requireOrgAgents})

		require.Equal(t, http.StatusOK, tc.Recorder.Code)
		stored, err := s.OrgGet(org.ID)
		require.NoError(t, err)
		assert.False(t, stored.RequireOrgAgents)
	})

	t.Run("rejects malformed body", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRawBody(http.MethodPatch, "application/json", []byte("{"))(tc)
		tc.Ctx.Set("org", org)

		PatchOrg(tc.Ctx)

		assert.Equal(t, http.StatusBadRequest, tc.Recorder.Code)
	})
}
