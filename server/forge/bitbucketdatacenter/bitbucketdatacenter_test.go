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
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucketdatacenter/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestNew(t *testing.T) {
	forge, err := New(Opts{
		URL:          "http://localhost:8080",
		Username:     "0ZXh0IjoiI",
		Password:     "I1NiIsInR5",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	assert.NoError(t, err)
	assert.NotNil(t, forge)
	cl, ok := forge.(*client)
	assert.True(t, ok)
	assert.Equal(t, &client{
		url:          "http://localhost:8080",
		urlAPI:       "http://localhost:8080/rest",
		username:     "0ZXh0IjoiI",
		password:     "I1NiIsInR5",
		clientID:     "client-id",
		clientSecret: "client-secret",
	}, cl)
}

func TestBitbucketDC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := fixtures.Server()
	defer s.Close()
	c := &client{
		urlAPI: s.URL,
	}

	ctx := context.Background()

	repo, err := c.Repo(ctx, fakeUser, model.ForgeRemoteID("1234"), "PRJ", "repo-slug")
	assert.NoError(t, err)
	assert.Equal(t, &model.Repo{
		Name:          "repo-slug-2",
		Owner:         "PRJ",
		Perm:          &model.Perm{Pull: true, Push: true},
		Branch:        "main",
		IsSCMPrivate:  true,
		PREnabled:     true,
		ForgeRemoteID: model.ForgeRemoteID("1234"),
		FullName:      "PRJ/repo-slug-2",
	}, repo)

	// org
	org, err := c.Org(ctx, fakeUser, "ORG")
	assert.NoError(t, err)
	assert.Equal(t, &model.Org{
		Name:   "ORG",
		IsUser: false,
	}, org)

	// user
	org, err = c.Org(ctx, fakeUser, "~ORG")
	assert.NoError(t, err)
	assert.Equal(t, &model.Org{
		Name:   "~ORG",
		IsUser: true,
	}, org)
}

var fakeUser = &model.User{
	AccessToken: "fake",
	Expiry:      time.Now().Add(1 * time.Hour).Unix(),
}
