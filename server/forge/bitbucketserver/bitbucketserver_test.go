// Copyright 2023 Woodpecker Authors
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

package bitbucketserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/franela/goblin"
	"github.com/gin-gonic/gin"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucketserver/fixtures"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func Test_Bitbucket_DC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := fixtures.Server()
	k, _ := rsa.GenerateKey(rand.Reader, 2048)
	c := &client{
		URLApi:   s.URL,
		Consumer: CreateConsumer(s.URL, "somelongsecretkey", k),
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
					URL:               "http://localhost:8080",
					Username:          "0ZXh0IjoiI",
					Password:          "I1NiIsInR5",
					ConsumerKey:       "somelongsecretkey",
					ConsumerRSA:       "",
					ConsumerRSAString: generatePrivateKey(),
					SkipVerify:        true,
				})
				g.Assert(err).IsNil()
				g.Assert(forge).IsNotNil()
				g.Assert(forge.(*client).url).Equal("http://localhost:8080")
				g.Assert(forge.(*client).Username).Equal("0ZXh0IjoiI")
				g.Assert(forge.(*client).Password).Equal("I1NiIsInR5")
				g.Assert(forge.(*client).SkipVerify).Equal(true)
			})
		})

		g.Describe("Requesting a repository", func() {
			g.It("should return repository details", func() {
				repo, err := c.Repo(ctx, fakeUser, model.ForgeRemoteID("1234"), "PRJ", "repo-slug")
				g.Assert(err).IsNil()
				g.Assert(repo.Name).Equal("repo-slug-2")
				g.Assert(repo.Owner).Equal("PRJ")
				g.Assert(repo.Perm).Equal(&model.Perm{Pull: true})
				g.Assert(repo.Branch).Equal("main")
			})
		})
	})
}

func generatePrivateKey() string {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	der := x509.MarshalPKCS1PrivateKey(key)
	pemSpec := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   der,
	}
	privatePEM := pem.EncodeToMemory(&pemSpec)
	return string(privatePEM)
}

var fakeUser = &model.User{
	Token: "fake",
}
