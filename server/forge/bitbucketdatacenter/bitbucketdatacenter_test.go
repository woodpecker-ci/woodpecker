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

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucketdatacenter/fixtures"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestBitbucketDC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := fixtures.Server()
	c := &client{
		urlAPI: s.URL,
	}

	ctx := context.Background()
	g := goblin.Goblin(t)
	g.Describe("Bitbucket DataCenter/Server", func() {
		g.After(func() {
			s.Close()
		})

		g.Describe("Creating a forge", func() {
			g.It("Should return client with specified options", func() {
				forge, err := New(Opts{
					URL:          "http://localhost:8080",
					Username:     "0ZXh0IjoiI",
					Password:     "I1NiIsInR5",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				})
				g.Assert(err).IsNil()
				g.Assert(forge).IsNotNil()
				cl, ok := forge.(*client)
				g.Assert(ok).IsTrue()
				g.Assert(cl.url).Equal("http://localhost:8080")
				g.Assert(cl.username).Equal("0ZXh0IjoiI")
				g.Assert(cl.password).Equal("I1NiIsInR5")
				g.Assert(cl.clientID).Equal("client-id")
				g.Assert(cl.clientSecret).Equal("client-secret")
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("should return repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, model.ForgeRemoteID("1234"), "PRJ", "repo-slug")
				g.Assert(err).IsNil()
				g.Assert(repo.Name).Equal("repo-slug-2")
				g.Assert(repo.Owner).Equal("PRJ")
				g.Assert(repo.Perm).Equal(&model.Perm{Pull: true, Push: true})
				g.Assert(repo.Branch).Equal("main")
			})
		})

		g.Describe("Getting organization", func() {
			g.It("should map organization", func() {
				org, err := c.Org(ctx, fakeUser, "ORG")
				g.Assert(err).IsNil()
				g.Assert(org.Name).Equal("ORG")
				g.Assert(org.IsUser).IsFalse()
			})
			g.It("should map user organization", func() {
				org, err := c.Org(ctx, fakeUser, "~ORG")
				g.Assert(err).IsNil()
				g.Assert(org.Name).Equal("~ORG")
				g.Assert(org.IsUser).IsTrue()
			})
		})
	})
}

var fakeUser = &model.User{
	AccessToken: "fake",
	Expiry:      time.Now().Add(1 * time.Hour).Unix(),
}
