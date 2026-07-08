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

package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
)

// fakeMembership is a canned cache.MembershipService that records whether it
// was consulted.
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

func newOrgMemberTestContext(user *model.User, org *model.Org) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if user != nil {
		c.Set("user", user)
	}
	if org != nil {
		c.Set("org", org)
	}
	return c, rec
}

func installMemberForgeManager(t *testing.T) {
	t.Helper()
	mgr := manager_mocks.NewMockManager(t)
	_forge := forge_mocks.NewMockForge(t)
	mgr.On("ForgeFromUser", mock.Anything).Return(_forge, nil).Maybe()
	server.Config.Services.Manager = mgr
}

func TestMustOrgMember(t *testing.T) {
	t.Run("same login on another forge is denied access to foreign user-org", func(t *testing.T) {
		installMemberForgeManager(t)
		membership := &fakeMembership{}
		server.Config.Services.Membership = membership

		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 2, Name: "alice", ForgeID: 2, IsUser: true}
		c, rec := newOrgMemberTestContext(user, org)

		MustOrgMember(true)(c)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("membership on user's forge does not grant access to same-named org on another forge", func(t *testing.T) {
		installMemberForgeManager(t)
		// admin of "acme" on forge 1 ...
		membership := &fakeMembership{perm: &model.OrgPerm{Member: true, Admin: true}}
		server.Config.Services.Membership = membership

		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		// ... must not grant anything on "acme" of forge 2
		org := &model.Org{ID: 3, Name: "acme", ForgeID: 2}
		c, rec := newOrgMemberTestContext(user, org)

		MustOrgMember(true)(c)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, membership.called, "forge of the user must not be asked about a foreign forge's org")
	})

	t.Run("own user-org on same forge is allowed", func(t *testing.T) {
		installMemberForgeManager(t)
		membership := &fakeMembership{}
		server.Config.Services.Membership = membership

		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 4, Name: "alice", ForgeID: 1, IsUser: true}
		c, rec := newOrgMemberTestContext(user, org)

		MustOrgMember(true)(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("instance admin is allowed across forges", func(t *testing.T) {
		installMemberForgeManager(t)
		server.Config.Services.Membership = &fakeMembership{}

		user := &model.User{ID: 1, Login: "root", ForgeID: 1, Admin: true}
		org := &model.Org{ID: 5, Name: "acme", ForgeID: 2}
		c, rec := newOrgMemberTestContext(user, org)

		MustOrgMember(true)(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("org member on same forge is allowed", func(t *testing.T) {
		installMemberForgeManager(t)
		membership := &fakeMembership{perm: &model.OrgPerm{Member: true, Admin: true}}
		server.Config.Services.Membership = membership

		user := &model.User{ID: 1, Login: "alice", ForgeID: 1}
		org := &model.Org{ID: 6, Name: "acme", ForgeID: 1}
		c, rec := newOrgMemberTestContext(user, org)

		MustOrgMember(true)(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.False(t, c.IsAborted())
		assert.True(t, membership.called)
	})
}
