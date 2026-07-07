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
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
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
}
