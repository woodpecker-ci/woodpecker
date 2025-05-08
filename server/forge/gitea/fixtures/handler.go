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
// Gitea requests and returning mock responses.
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
	e.GET("/api/v1/repos/:owner/:name/git/commits/:sha", getCommit)

	return e
}

func listRepoHooks(c *gin.Context) {
	page := c.Query("page")
	if page != "" && page != "1" {
		c.String(http.StatusOK, "[]")
	} else {
		c.String(http.StatusOK, listRepoHookPayloads)
	}
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

func createRepoCommitStatus(c *gin.Context) {
	if c.Param("commit") == "v1.0.0" || c.Param("commit") == "9ecad50" {
		c.String(http.StatusOK, repoPayload)
	}
	c.String(http.StatusNotFound, "")
}

func getRepoFile(c *gin.Context) {
	file := c.Param("file")
	ref := c.Query("ref")

	if file == "file_not_found" {
		c.String(http.StatusNotFound, "")
	}
	if ref == "v1.0.0" || ref == "9ecad50" {
		c.String(http.StatusOK, repoFilePayload)
	}
	c.String(http.StatusNotFound, "")
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
	if in.Type != "gitea" ||
		in.Conf.Type != "json" ||
		in.Conf.URL != "http://localhost" {
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.String(http.StatusOK, "{}")
}

func deleteRepoHook(c *gin.Context) {
	c.String(http.StatusOK, "{}")
}

func getUserRepos(c *gin.Context) {
	switch c.Request.Header.Get("Authorization") {
	case "token repos_not_found":
		c.String(http.StatusNotFound, "")
	default:
		page := c.Query("page")
		if page != "" && page != "1" {
			c.String(http.StatusOK, "[]")
		} else {
			c.String(http.StatusOK, userRepoPayload)
		}
	}
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{"version": "1.18.0"})
}

func getPRFiles(c *gin.Context) {
	page := c.Query("page")
	if page == "1" {
		c.String(http.StatusOK, prFilesPayload)
	} else {
		c.String(http.StatusOK, "[]")
	}
}

func getCommit(c *gin.Context) {
	switch c.Param("sha") {
	case "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c":
		c.String(http.StatusOK, commitPayload)
	default:
		c.String(http.StatusNotFound, "")
	}
}

const listRepoHookPayloads = `
[
  {
    "id": 1,
    "type": "gitea",
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

const commitPayload = `
{
  "url": "http://localhost:3000/api/v1/repos/qwerty287/woodpecker/git/commits/0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
  "sha": "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
  "created": "2025-01-05T12:31:42+02:00",
  "html_url": "http://localhost:3000/qwerty287/woodpecker/commit/0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
  "commit": {
    "url": "http://localhost:3000/api/v1/repos/qwerty287/woodpecker/git/commits/0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
    "author": {
      "name": "qwerty287",
      "email": "qwerty287@noreply.localhost",
      "date": "2025-01-05T12:31:42+02:00"
    },
    "committer": {
      "name": "qwerty287",
      "email": "qwerty287@noreply.localhost",
      "date": "2025-01-05T12:31:42+02:00"
    },
    "message": "README.md aktualisiert\n",
    "tree": {
      "url": "http://localhost:3000/api/v1/repos/qwerty287/woodpecker/git/trees/0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
      "sha": "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c",
      "created": "2025-01-05T12:31:42+02:00"
    },
    "verification": {
      "verified": false,
      "reason": "gpg.error.not_signed_commit",
      "signature": "",
      "signer": null,
      "payload": ""
    }
  },
  "author": {
    "id": 1,
    "login": "qwerty287",
    "login_name": "",
    "source_id": 0,
    "full_name": "",
    "email": "qwerty287@noreply.localhost",
    "avatar_url": "http://localhost:3000/avatars/25a4ce9e2945c8583f82ce7b2ee8bc3c",
    "html_url": "http://localhost:3000/qwerty287",
    "language": "",
    "is_admin": false,
    "last_login": "0001-01-01T00:00:00Z",
    "created": "2023-04-12T18:52:45+03:00",
    "restricted": false,
    "active": false,
    "prohibit_login": false,
    "location": "",
    "website": "",
    "description": "",
    "visibility": "public",
    "followers_count": 0,
    "following_count": 0,
    "starred_repos_count": 0,
    "username": "qwerty287"
  },
  "files": [
    {
      "filename": "README.md",
      "status": "modified"
    }
  ],
  "stats": {
    "total": 2,
    "additions": 1,
    "deletions": 1
  }
}
`
