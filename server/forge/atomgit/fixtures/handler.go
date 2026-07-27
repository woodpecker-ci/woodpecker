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

package fixtures

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler returns an http.Handler that is capable of handling a variety of mock
// AtomGit requests and returning mock responses.
func Handler() http.Handler {
	gin.SetMode(gin.TestMode)

	e := gin.New()
	e.GET("/api/v5/user", getUser)
	e.GET("/api/v5/user/repos", getUserRepos)
	e.GET("/api/v5/repos/:owner/:name", getRepo)
	e.GET("/api/v5/repositories/:id", getRepoByID)
	e.GET("/api/v5/repos/:owner/:name/raw/:file", getRepoFile)
	e.POST("/api/v5/repos/:owner/:name/hooks", createRepoHook)
	e.GET("/api/v5/repos/:owner/:name/hooks", listRepoHooks)
	e.DELETE("/api/v5/repos/:owner/:name/hooks/:id", deleteRepoHook)
	e.GET("/api/v5/repos/:owner/:name/pulls/:index/files", getPRFiles)
	e.GET("/api/v5/user/memberships/orgs/:org", getOrgMembership)
	e.GET("/api/v5/users/:org", getOrg)

	return e
}

func getUser(c *gin.Context) {
	c.JSON(http.StatusOK, userPayload)
}

func getUserRepos(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []any{repoPayload}})
}

func getRepo(c *gin.Context) {
	if c.Param("name") == "repo_not_found" {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found", "code": 404})
		return
	}
	c.JSON(http.StatusOK, repoPayload)
}

func getRepoByID(c *gin.Context) {
	if c.Param("id") == "0" {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found", "code": 404})
		return
	}
	c.JSON(http.StatusOK, repoPayload)
}

func getRepoFile(c *gin.Context) {
	c.String(http.StatusOK, "{ platform: linux/amd64 }")
}

func createRepoHook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": 9529, "url": c.PostForm("url")})
}

func listRepoHooks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []any{
		gin.H{
			"id":                    9529,
			"url":                   "http://localhost",
			"password":              "secret",
			"push_events":           true,
			"tag_push_events":       true,
			"merge_requests_events": true,
		},
	}})
}

func deleteRepoHook(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func getPRFiles(c *gin.Context) {
	c.JSON(http.StatusOK, []any{
		gin.H{"new_path": "README.md", "old_path": "README.md", "new_file": false, "deleted_file": false},
	})
}

func getOrgMembership(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"state": "active"})
}

func getOrg(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": 1, "username": c.Param("org"), "name": c.Param("org")})
}

var userPayload = gin.H{
	"id":         1,
	"username":   "someuser",
	"name":       "Some User",
	"email":      "someuser@atomgit.com",
	"state":      "active",
	"avatar_url": "https://atomgit.com/avatar.png",
	"html_url":   "https://atomgit.com/someuser",
}

var repoPayload = gin.H{
	"id":                  5,
	"human_name":          "repo_name",
	"name":                "repo_name",
	"path":                "repo_name",
	"path_with_namespace": "test_name/repo_name",
	"full_name":           "test_name/repo_name",
	"description":         "test repo",
	"html_url":            "http://localhost/test_name/repo_name",
	"web_url":             "http://localhost/test_name/repo_name",
	"http_url_to_repo":    "http://localhost/test_name/repo_name.git",
	"ssh_url_to_repo":     "git@localhost:test_name/repo_name.git",
	"default_branch":      "master",
	"visibility_level":    0,
	"public":              true,
	"archived":            false,
	"has_pull_requests":   true,
	"namespace": gin.H{
		"id":        2,
		"name":      "test_name",
		"path":      "test_name",
		"kind":      "user",
		"full_path": "test_name",
	},
}
