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
	e.GET("/2.0/workspaces/", getWorkspaces)
	e.GET("/2.0/repositories/:owner/:name", getRepo)
	e.GET("/2.0/repositories/:owner/:name/hooks", getRepoHooks)
	e.GET("/2.0/repositories/:owner/:name/src/:commit/:file", getRepoFile)
	e.DELETE("/2.0/repositories/:owner/:name/hooks/:hook", deleteRepoHook)
	e.POST("/2.0/repositories/:owner/:name/hooks", createRepoHook)
	e.POST("/2.0/repositories/:owner/:name/commit/:commit/statuses/build", createRepoStatus)
	e.GET("/2.0/repositories/:owner", getUserRepos)
	e.GET("/2.0/user/", getUser)
	e.GET("/2.0/user/permissions/repositories", getPermissions)
	e.GET("/2.0/repositories/:owner/:name/commits/:commit", getBranchHead)
	e.GET("/2.0/repositories/:owner/:name/pullrequests", getPullRequests)
	return e
}

func getOauth(c *gin.Context) {
	if c.PostForm("error") == "invalid_scope" {
		c.String(http.StatusInternalServerError, "")
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

func getWorkspaces(c *gin.Context) {
	// TODO: should the role be used ?
	// role, _ := c.Params.Get("role")

	switch c.Request.Header.Get("Authorization") {
	case "Bearer teams_not_found", "Bearer c81e728d":
		c.String(http.StatusNotFound, "")
	default:
		if c.Query("page") == "" || c.Query("page") == "1" {
			c.String(http.StatusOK, workspacesPayload)
		} else {
			c.String(http.StatusOK, "{\"values\":[]}")
		}
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
		if c.Query("page") == "" || c.Query("page") == "1" {
			c.String(http.StatusOK, repoHookPayload)
		} else {
			c.String(http.StatusOK, "{\"values\":[]}")
		}
	}
}

func getRepoFile(c *gin.Context) {
	switch c.Param("file") {
	case "dir":
		c.String(http.StatusOK, repoDirPayload)
	case "dir_not_found/":
		c.String(http.StatusNotFound, "")
	case "file_not_found":
		c.String(http.StatusNotFound, "")
	default:
		c.String(http.StatusOK, repoFilePayload)
	}
}

func getBranchHead(c *gin.Context) {
	switch c.Param("commit") {
	case "branch_name":
		c.String(http.StatusOK, branchCommitsPayload)
	default:
		c.String(http.StatusNotFound, "")
	}
}

func getPullRequests(c *gin.Context) {
	switch c.Param("name") {
	case "repo_name":
		c.String(http.StatusOK, pullRequestsPayload)
	default:
		c.String(http.StatusNotFound, "")
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

func getUserRepos(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "Bearer repos_not_found", "Bearer 70efdf2e":
		c.String(http.StatusNotFound, "")
	default:
		if c.Query("page") == "" || c.Query("page") == "1" {
			c.String(http.StatusOK, userRepoPayload)
		} else {
			c.String(http.StatusOK, "{\"values\":[]}")
		}
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

const repoDirPayload = `
{
    "pagelen": 10,
    "page": 1,
    "values": [
        {
            "path": "README.md",
            "type": "commit_file"
        },
        {
            "path": "test",
            "type": "commit_directory"
        },
        {
            "path": ".gitignore",
            "type": "commit_file"
        }
    ]
}
`

const branchCommitsPayload = `
{
    "values": [
        {
            "hash": "branch_head_name"
        },
        {
            "hash": "random1"
        },
        {
            "hash": "random2"
        }
    ]
}
`

const pullRequestsPayload = `
{
		 "values": [
        {
            "id": 123,
						"title": "PRs title"
        },
        {
            "id": 456,
						"title": "Another PRs title"
        }
    ],
		"pagelen": 10,
    "size": 2,
    "page": 1
}
`

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

const workspacesPayload = `
{
	"page": 1,
  "pagelen": 100,
	"size": 1,
  "values": [
    {
			"type": "workspace",
			"uuid": "{c7a04a76-fa20-43e4-dc42-a7506db4c95b}",
			"name": "Ueber Dev",
			"slug": "ueberdev42",
      "links": {
				"avatar": {
				  "href": "https://bitbucket.org/workspaces/ueberdev42/avatar/?ts=1658761964"
			  },
			  "html": {
				  "href": "https://bitbucket.org/ueberdev42/"
			  },
			  "self": {
				  "href": "https://api.bitbucket.org/2.0/workspaces/ueberdev42"
			  }
      }
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
