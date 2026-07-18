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
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler returns an http.Handler that is capable of handling a variety of mock
// Bitbucket requests and returning mock responses.
func Handler() http.Handler {
	gin.SetMode(gin.TestMode)

	e := gin.New()
	e.GET("/api/v3/repos/:owner/:name", getRepo)
	e.GET("/api/v3/repositories/:id", getRepoByID)
	e.GET("/api/v3/orgs/:org/memberships/:user", getMembership)
	e.GET("/api/v3/user/memberships/orgs/:org", getMembership)
	e.POST("/api/graphql", graphqlDir)
	e.GET("/api/v3/repos/:owner/:name/contents/*path", getContents)

	return e
}

// graphqlDir mocks the single-request directory fetch used by Dir.
func graphqlDir(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	var req struct {
		Variables struct {
			Expression string `json:"expression"`
		} `json:"variables"`
	}
	_ = c.ShouldBindJSON(&req)
	switch {
	case strings.HasSuffix(req.Variables.Expression, ":somedir"):
		c.String(http.StatusOK, `{"data": {"repository": {"object": {"__typename": "Tree", "entries": [
			{"name": "a.yaml", "type": "blob", "mode": 33188, "object": {"text": "pipeline:", "isTruncated": false, "isBinary": false}},
			{"name": "sub", "type": "tree", "mode": 16384, "object": {}}
		]}}}}`)
	case strings.HasSuffix(req.Variables.Expression, ":symlink-dir"):
		// symlink blobs carry the link target as text, the forge must fall
		// back to the REST path which resolves them
		c.String(http.StatusOK, `{"data": {"repository": {"object": {"__typename": "Tree", "entries": [
			{"name": "link.yaml", "type": "blob", "mode": 40960, "object": {"text": "../target.yaml", "isTruncated": false, "isBinary": false}}
		]}}}}`)
	case strings.HasSuffix(req.Variables.Expression, ":rest-fallback-dir"):
		// force the caller onto the per-file REST fallback
		c.String(http.StatusBadGateway, "")
	default:
		c.String(http.StatusOK, `{"data": {"repository": {"object": null}}}`)
	}
}

// getContents serves the per-file REST fallback.
func getContents(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.String(http.StatusUnauthorized, "")
		return
	}
	if strings.HasSuffix(c.Param("path"), "rest-fallback-dir") {
		c.String(http.StatusOK, `[{"type": "file", "name": "b.yaml", "path": "rest-fallback-dir/b.yaml"}, {"type": "dir", "name": "nested", "path": "rest-fallback-dir/nested"}]`)
		return
	}
	if strings.HasSuffix(c.Param("path"), "symlink-dir") {
		c.String(http.StatusOK, `[{"type": "symlink", "name": "link.yaml", "path": "symlink-dir/link.yaml"}]`)
		return
	}
	// base64 of "pipeline:"
	c.String(http.StatusOK, `{"type": "file", "encoding": "base64", "name": "config", "content": "cGlwZWxpbmU6"}`)
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
