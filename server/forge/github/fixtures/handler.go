// Copyright 2018 Drone.IO Inc.
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

package fixtures

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// InstallationToken is the installation access token returned by the mock
// installation token endpoint.
const InstallationToken = "ghs_mock_installation_token"

// Handler returns an http.Handler that is capable of handling a variety of mock
// Bitbucket requests and returning mock responses.
func Handler() http.Handler {
	gin.SetMode(gin.TestMode)

	e := gin.New()
	e.GET("/api/v3/repos/:owner/:name", getRepo)
	e.GET("/api/v3/repositories/:id", getRepoByID)
	e.GET("/api/v3/orgs/:org/memberships/:user", getMembership)
	e.GET("/api/v3/user/memberships/orgs/:org", getMembership)
	e.GET("/api/v3/repos/:owner/:name/installation", getRepoInstallation)
	e.POST("/api/v3/app/installations/:id/access_tokens", createInstallationToken)
	e.POST("/api/v3/repos/:owner/:name/statuses/:commit", createStatus)
	e.GET("/api/v3/app", getApp)
	e.GET("/api/v3/app/installations", listInstallations)

	return e
}

func getApp(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	c.String(http.StatusOK, `{"id": 12345, "name": "Woodpecker Test App", "slug": "woodpecker-test-app"}`)
}

func listInstallations(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	c.String(http.StatusOK, `[{"id": 42}]`)
}

func getRepoInstallation(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	switch c.Param("owner") {
	case "not-installed":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, `{"id": 42}`)
	}
}

func createInstallationToken(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	if c.Param("id") != "42" {
		c.String(http.StatusNotFound, "")
		return
	}
	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	c.String(http.StatusCreated, fmt.Sprintf(`{"token": %q, "expires_at": %q}`, InstallationToken, expiresAt))
}

// createStatus only accepts the mock installation token, so tests can assert
// that a status was sent with GitHub App credentials.
func createStatus(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer "+InstallationToken {
		c.String(http.StatusUnauthorized, "")
		return
	}
	c.String(http.StatusCreated, `{"id": 1}`)
}

func getRepo(c *gin.Context) {
	switch c.Param("name") {
	case "repo_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, repoPayload)
	}
}

func getRepoByID(c *gin.Context) {
	switch c.Param("id") {
	case "repo_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, repoPayload)
	}
}

func getMembership(c *gin.Context) {
	switch c.Param("org") {
	case "org_not_found":
		c.String(http.StatusNotFound, "")
	case "github":
		c.String(http.StatusOK, membershipIsMemberPayload)
	default:
		c.String(http.StatusOK, membershipIsOwnerPayload)
	}
}

var repoPayload = `
{
	"id": 5,
	"owner": {
		"login": "octocat",
		"avatar_url": "https://github.com/images/error/octocat_happy.gif"
	},
	"name": "Hello-World",
	"full_name": "octocat/Hello-World",
	"private": true,
	"html_url": "https://github.com/octocat/Hello-World",
	"clone_url": "https://github.com/octocat/Hello-World.git",
	"language": null,
	"permissions": {
		"admin": true,
		"push": true,
		"pull": true
	}
}
`

var membershipIsOwnerPayload = `
{
	"url": "https://api.github.com/orgs/octocat/memberships/octocat",
	"state": "active",
	"role": "admin",
	"organization_url": "https://api.github.com/orgs/octocat",
	"user": {
		"login": "octocat",
		"id": 5555555,
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"gravatar_id": "",
		"url": "https://api.github.com/users/octocat",
		"html_url": "https://github.com/octocat",
		"followers_url": "https://api.github.com/users/octocat/followers",
		"following_url": "https://api.github.com/users/octocat/following{/other_user}",
		"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
		"organizations_url": "https://api.github.com/users/octocat/orgs",
		"repos_url": "https://api.github.com/users/octocat/repos",
		"events_url": "https://api.github.com/users/octocat/events{/privacy}",
		"received_events_url": "https://api.github.com/users/octocat/received_events",
		"type": "User",
		"site_admin": false
	},
	"organization": {
		"login": "octocat",
		"id": 5555556,
		"url": "https://api.github.com/orgs/octocat",
		"repos_url": "https://api.github.com/orgs/octocat/repos",
		"events_url": "https://api.github.com/orgs/octocat/events",
		"hooks_url": "https://api.github.com/orgs/octocat/hooks",
		"issues_url": "https://api.github.com/orgs/octocat/issues",
		"members_url": "https://api.github.com/orgs/octocat/members{/member}",
		"public_members_url": "https://api.github.com/orgs/octocat/public_members{/member}",
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"description": ""
	}
}
`

var membershipIsMemberPayload = `
{
	"url": "https://api.github.com/orgs/github/memberships/octocat",
	"state": "active",
	"role": "member",
	"organization_url": "https://api.github.com/orgs/github",
	"user": {
		"login": "octocat",
		"id": 5555555,
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"gravatar_id": "",
		"url": "https://api.github.com/users/octocat",
		"html_url": "https://github.com/octocat",
		"followers_url": "https://api.github.com/users/octocat/followers",
		"following_url": "https://api.github.com/users/octocat/following{/other_user}",
		"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
		"organizations_url": "https://api.github.com/users/octocat/orgs",
		"repos_url": "https://api.github.com/users/octocat/repos",
		"events_url": "https://api.github.com/users/octocat/events{/privacy}",
		"received_events_url": "https://api.github.com/users/octocat/received_events",
		"type": "User",
		"site_admin": false
	},
	"organization": {
		"login": "octocat",
		"id": 5555557,
		"url": "https://api.github.com/orgs/github",
		"repos_url": "https://api.github.com/orgs/github/repos",
		"events_url": "https://api.github.com/orgs/github/events",
		"hooks_url": "https://api.github.com/orgs/github/hooks",
		"issues_url": "https://api.github.com/orgs/github/issues",
		"members_url": "https://api.github.com/orgs/github/members{/member}",
		"public_members_url": "https://api.github.com/orgs/github/public_members{/member}",
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"description": ""
	}
}
`
