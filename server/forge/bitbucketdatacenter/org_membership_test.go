// Copyright 2024 Woodpecker Authors
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

package bitbucketdatacenter

import (
	"testing"
	"time"

	"github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucketdatacenter/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var fakeUserOrgTest = &model.User{
	AccessToken: "fake",
	Expiry:      time.Now().Add(1 * time.Hour).Unix(),
}

func TestOrgMembership(t *testing.T) {
	tests := []struct {
		name    string
		project string
		want    model.OrgPerm
	}{
		{"user has admin permissions", "PRJ-ADMIN", model.OrgPerm{Member: true, Admin: true}},
		{"user has write permissions only", "PRJ-WRITE", model.OrgPerm{Member: true, Admin: false}},
		{"user has no permissions", "PRJ-NONE", model.OrgPerm{Member: false, Admin: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := fixtures.ServerWithOrgPermissions()
			defer s.Close()

			c := &client{urlAPI: s.URL}

			orgPerm, err := c.OrgMembership(t.Context(), fakeUserOrgTest, tt.project)
			require.NoError(t, err)
			require.NotNil(t, orgPerm)
			assert.Equal(t, tt.want, *orgPerm)
		})
	}
}

func TestCheckUserOrgPermissions(t *testing.T) {
	tests := []struct {
		name    string
		project string
		want    model.OrgPerm
	}{
		{"admin permissions", "PRJ-ADMIN", model.OrgPerm{Member: true, Admin: true}},
		{"write permissions", "PRJ-WRITE", model.OrgPerm{Member: true, Admin: false}},
		{"no permissions", "PRJ-NONE", model.OrgPerm{Member: false, Admin: false}},
	}

	s := fixtures.ServerWithOrgPermissions()
	defer s.Close()

	c := &client{urlAPI: s.URL}
	bc, err := c.newClient(t.Context(), fakeUserOrgTest)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgPerm, err := checkUserOrgPermissions(t.Context(), tt.project, bc)
			require.NoError(t, err)
			require.NotNil(t, orgPerm)
			assert.Equal(t, tt.want, *orgPerm)
		})
	}
}

func TestHasRepositoriesWithPermissionLevel(t *testing.T) {
	tests := []struct {
		name       string
		project    string
		permission bitbucket.Permission
		want       bool
	}{
		{"has admin repos", "PRJ-ADMIN", bitbucket.PermissionRepoAdmin, true},
		{"no admin repos for write project", "PRJ-WRITE", bitbucket.PermissionRepoAdmin, false},
		{"has write repos", "PRJ-WRITE", bitbucket.PermissionRepoWrite, true},
		{"no repos with permission", "PRJ-NONE", bitbucket.PermissionRepoAdmin, false},
	}

	s := fixtures.ServerWithOrgPermissions()
	defer s.Close()

	c := &client{urlAPI: s.URL}
	bc, err := c.newClient(t.Context(), fakeUserOrgTest)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRepos, err := hasRepositoriesWithPermissionLevel(t.Context(), tt.project, tt.permission, bc)
			require.NoError(t, err)
			assert.Equal(t, tt.want, hasRepos)
		})
	}
}
