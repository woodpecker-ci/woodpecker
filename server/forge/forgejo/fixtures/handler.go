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

	"github.com/gin-gonic/gin"
)

// Handler returns an http.Handler that is capable of handling a variety of mock
// Forgejo requests and returning mock responses.
func Handler() http.Handler {
	gin.SetMode(gin.TestMode)

	e := gin.New()
	e.GET("/api/v1/repos/:owner/:name", getRepo)
	e.GET("/api/v1/repositories/:id", getRepoByID)
	e.GET("/api/v1/repos/:owner/:name/raw/:file", getRepoFile)
	e.POST("/api/v1/repos/:owner/:name/hooks", createRepoHook)
	e.GET("/api/v1/repos/:owner/:name/hooks", listRepoHooks)
	e.DELETE("/api/v1/repos/:owner/:name/hooks/:id", deleteRepoHook)
	e.POST("/api/v1/repos/:owner/:name/statuses/:commit", createRepoCommitStatus)
	e.GET("/api/v1/repos/:owner/:name/pulls/:index/files", getPRFiles)
	e.GET("/api/v1/user/repos", getUserRepos)
	e.GET("/api/v1/version", getVersion)

	return e
}

func listRepoHooks(c *gin.Context) {
	c.String(200, listRepoHookPayloads)
}

func getRepo(c *gin.Context) {
	switch c.Param("name") {
	case "repo_not_found":
		c.String(404, "")
	default:
		c.String(200, repoPayload)
	}
}

func getRepoByID(c *gin.Context) {
	switch c.Param("id") {
	case "repo_not_found":
		c.String(404, "")
	default:
		c.String(200, repoPayload)
	}
}

func createRepoCommitStatus(c *gin.Context) {
	if c.Param("commit") == "v1.0.0" || c.Param("commit") == "9ecad50" {
		c.String(200, repoPayload)
	}
	c.String(404, "")
}

func getRepoFile(c *gin.Context) {
	file := c.Param("file")
	ref := c.Query("ref")

	if file == "file_not_found" {
		c.String(404, "")
	}
	if ref == "v1.0.0" || ref == "9ecad50" {
		c.String(200, repoFilePayload)
	}
	c.String(404, "")
}

func createRepoHook(c *gin.Context) {
	in := struct {
		Type string `json:"type"`
		Conf struct {
			Type string `json:"content_type"`
			URL  string `json:"url"`
		} `json:"config"`
	}{}
	_ = c.BindJSON(&in)
	if (in.Type != "gitea" && in.Type != "forgejo") ||
		in.Conf.Type != "json" ||
		in.Conf.URL != "http://localhost" {
		c.String(500, in.Type)
		return
	}

	c.String(200, "{}")
}

func deleteRepoHook(c *gin.Context) {
	c.String(200, "{}")
}

func getUserRepos(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "token repos_not_found":
		c.String(404, "")
	default:
		page := c.Query("page")
		if page != "" && page != "1" {
			c.String(200, "[]")
		} else {
			c.String(200, userRepoPayload)
		}
	}
}

func getVersion(c *gin.Context) {
	c.JSON(200, map[string]interface{}{"version": "1.18.0"})
}

func getPRFiles(c *gin.Context) {
	page := c.Query("page")
	if page == "1" {
		c.String(200, prFilesPayload)
	} else {
		c.String(200, "[]")
	}
}

const listRepoHookPayloads = `
[
  {
    "id": 1,
    "type": "forgejo",
    "config": {
      "content_type": "json",
      "url": "http:\/\/localhost\/hook?access_token=1234567890"
    }
  }
]
`

const repoPayload = `
{
	"id": 5,
  "owner": {
    "login": "test_name",
    "email": "octocat@github.com",
    "avatar_url": "https:\/\/secure.gravatar.com\/avatar\/8c58a0be77ee441bb8f8595b7f1b4e87"
  },
  "full_name": "test_name\/repo_name",
  "private": true,
  "html_url": "http:\/\/localhost\/test_name\/repo_name",
  "clone_url": "http:\/\/localhost\/test_name\/repo_name.git",
  "permissions": {
    "admin": true,
    "push": true,
    "pull": true
  }
}
`

const repoFilePayload = `{ platform: linux/amd64 }`

const userRepoPayload = `
[
  {
		"id": 5,
    "owner": {
      "login": "test_name",
      "email": "octocat@github.com",
      "avatar_url": "https:\/\/secure.gravatar.com\/avatar\/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "full_name": "test_name\/repo_name",
    "private": true,
    "html_url": "http:\/\/localhost\/test_name\/repo_name",
    "clone_url": "http:\/\/localhost\/test_name\/repo_name.git",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    }
  }
]
`

const prFilesPayload = `
[
  {
    "filename": "README.md",
    "status": "changed",
    "additions": 2,
    "deletions": 0,
    "changes": 2,
    "html_url": "http://localhost/username/repo/src/commit/e79e4b0e8d9dd6f72b70e776c3317db7c19ca0fd/README.md",
    "contents_url": "http://localhost:3000/api/v1/repos/username/repo/contents/README.md?ref=e79e4b0e8d9dd6f72b70e776c3317db7c19ca0fd",
    "raw_url": "http://localhost/username/repo/raw/commit/e79e4b0e8d9dd6f72b70e776c3317db7c19ca0fd/README.md"
  }
]
`
