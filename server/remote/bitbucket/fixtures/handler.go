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

	"github.com/gin-gonic/gin"
)

// Handler returns an http.Handler that is capable of handling a variety of mock
// Bitbucket requests and returning mock responses.
func Handler() http.Handler {
	gin.SetMode(gin.TestMode)

	e := gin.New()
	e.POST("/site/oauth2/access_token", getOauth)
	e.GET("/2.0/repositories/:owner/:name", getRepo)
	e.GET("/2.0/repositories/:owner/:name/hooks", getRepoHooks)
	e.GET("/2.0/repositories/:owner/:name/src/:commit/:file", getRepoFile)
	e.DELETE("/2.0/repositories/:owner/:name/hooks/:hook", deleteRepoHook)
	e.POST("/2.0/repositories/:owner/:name/hooks", createRepoHook)
	e.POST("/2.0/repositories/:owner/:name/commit/:commit/statuses/build", createRepoStatus)
	e.GET("/2.0/repositories/:owner", getUserRepos)
	e.GET("/2.0/teams/", getUserTeams)
	e.GET("/2.0/user/", getUser)
	e.GET("/2.0/user/permissions/repositories", getPermissions)

	return e
}

func getOauth(c *gin.Context) {
	if c.PostForm("error") == "invalid_scope" {
		c.String(http.StatusInternalServerError, "")
		return
	}

	switch c.PostForm("code") {
	case "code_bad_request":
		c.String(http.StatusInternalServerError, "")
		return
	case "code_user_not_found":
		c.String(http.StatusOK, tokenNotFoundPayload)
		return
	}
	switch c.PostForm("refresh_token") {
	case "refresh_token_not_found":
		c.String(http.StatusNotFound, "")
	case "refresh_token_is_empty":
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, "{}")
	default:
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, tokenPayload)
	}
}

func getRepo(c *gin.Context) {
	switch c.Param("name") {
	case "not_found", "repo_unknown", "repo_not_found":
		c.String(http.StatusNotFound, "")
	case "permission_read", "permission_write", "permission_admin":
		c.String(http.StatusOK, fmt.Sprintf(permissionRepoPayload, c.Param("name")))
	default:
		c.String(http.StatusOK, repoPayload)
	}
}

func getRepoHooks(c *gin.Context) {
	switch c.Param("name") {
	case "hooks_not_found", "repo_no_hooks":
		c.String(http.StatusNotFound, "")
	case "hook_empty":
		c.String(http.StatusOK, "{}")
	default:
		c.String(http.StatusOK, repoHookPayload)
	}
}

func getRepoFile(c *gin.Context) {
	switch c.Param("file") {
	case "file_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, repoFilePayload)
	}
}

func createRepoStatus(c *gin.Context) {
	switch c.Param("name") {
	case "repo_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, "")
	}
}

func createRepoHook(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func deleteRepoHook(c *gin.Context) {
	switch c.Param("name") {
	case "hook_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, "")
	}
}

func getUser(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "Bearer user_not_found", "Bearer a87ff679":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, userPayload)
	}
}

func getUserTeams(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "Bearer teams_not_found", "Bearer c81e728d":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, userTeamPayload)
	}
}

func getUserRepos(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "Bearer repos_not_found", "Bearer 70efdf2e":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, userRepoPayload)
	}
}

func permission(p string) string {
	return fmt.Sprintf(permissionPayload, p)
}

func getPermissions(c *gin.Context) {
	query := c.Request.URL.Query()["q"][0]
	switch query {
	case `repository.full_name="test_name/permission_read"`:
		c.String(http.StatusOK, permission("read"))
	case `repository.full_name="test_name/permission_write"`:
		c.String(http.StatusOK, permission("write"))
	case `repository.full_name="test_name/permission_admin"`:
		c.String(http.StatusOK, permission("admin"))
	default:
		c.String(http.StatusOK, permission("read"))
	}
}

const tokenPayload = `
{
	"access_token":"2YotnFZFEjr1zCsicMWpAA",
	"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
	"token_type":"Bearer",
	"expires_in":3600
}
`

const tokenNotFoundPayload = `
{
	"access_token":"user_not_found",
	"refresh_token":"user_not_found",
	"token_type":"Bearer",
	"expires_in":3600
}
`

const repoPayload = `
{
  "full_name": "test_name/repo_name",
  "scm": "git",
  "is_private": true
}
`

const permissionRepoPayload = `
{
  "full_name": "test_name/%s",
  "scm": "git",
  "is_private": true
}
`

const repoHookPayload = `
{
  "pagelen": 10,
  "values": [
  	{
  	  "uuid": "{afe61e14-2c5f-49e8-8b68-ad1fb55fc052}",
  	  "url": "http://127.0.0.1"
  	}
  ],
  "page": 1,
  "size": 1
}
`

const repoFilePayload = "dummy payload"

const userPayload = `
{
  "username": "superman",
  "links": {
    "avatar": {
      "href": "http:\/\/i.imgur.com\/ZygP55A.jpg"
    }
  },
  "type": "user"
}
`

const userRepoPayload = `
{
  "page": 1,
  "pagelen": 10,
  "size": 1,
  "values": [
    {
      "links": {
        "avatar": {
            "href": "http:\/\/i.imgur.com\/ZygP55A.jpg"
        }
      },
      "full_name": "test_name/repo_name",
      "scm": "git",
      "is_private": true
    }
  ]
}
`

const userTeamPayload = `
{
  "pagelen": 100,
  "values": [
    {
      "username": "superfriends",
      "links": {
        "avatar": {
          "href": "http:\/\/i.imgur.com\/ZygP55A.jpg"
        }
      },
      "type": "team"
    }
  ]
}
`

const permissionPayload = `
{
  "pagelen": 1,
  "values": [
    {
      "permission": "%s"
    }
  ],
  "page": 1
}
`
