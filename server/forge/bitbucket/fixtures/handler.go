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
	case "{898477b2-a080-4089-b385-597a783db392}":
		c.String(http.StatusOK, repoPayloadFromHook)
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
	case "dir_not_found":
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

func getPermissions(c *gin.Context) {
	if c.Query("page") == "" || c.Query("page") == "1" {
		c.String(http.StatusOK, permissionsPayLoad)
	} else {
		c.String(http.StatusOK, "{\"values\":[]}")
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

const repoPayloadFromHook = `
{
  "type": "repository",
  "full_name": "martinherren1984/publictestrepo",
  "links": {
    "self": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo"
    },
    "html": {
      "href": "https://bitbucket.org/martinherren1984/publictestrepo"
    },
    "avatar": {
      "href": "https://bytebucket.org/ravatar/%7B898477b2-a080-4089-b385-597a783db392%7D?ts=default"
    },
    "pullrequests": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/pullrequests"
    },
    "commits": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/commits"
    },
    "forks": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/forks"
    },
    "watchers": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/watchers"
    },
    "branches": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/refs/branches"
    },
    "tags": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/refs/tags"
    },
    "downloads": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/downloads"
    },
    "source": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/src"
    },
    "clone": [
      {
        "name": "https",
        "href": "https://bitbucket.org/martinherren1984/publictestrepo.git"
      },
      {
        "name": "ssh",
        "href": "git@bitbucket.org:martinherren1984/publictestrepo.git"
      }
    ],
    "hooks": {
      "href": "https://api.bitbucket.org/2.0/repositories/martinherren1984/publictestrepo/hooks"
    }
  },
  "name": "PublicTestRepo",
  "slug": "publictestrepo",
  "description": "",
  "scm": "git",
  "website": null,
  "owner": {
    "display_name": "Martin Herren",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/users/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D"
      },
      "avatar": {
        "href": "https://secure.gravatar.com/avatar/37de364488b2ec474b5458ca86442bbb?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMH-2.png"
      },
      "html": {
        "href": "https://bitbucket.org/%7Bc5a0d676-fd27-4bd4-ac69-a7540d7b495b%7D/"
      }
    },
    "type": "user",
    "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
    "account_id": "5cf8e3a9678ca90f8e7cc8a8",
    "nickname": "Martin Herren"
  },
  "workspace": {
    "type": "workspace",
    "uuid": "{c5a0d676-fd27-4bd4-ac69-a7540d7b495b}",
    "name": "Martin Herren",
    "slug": "martinherren1984",
    "links": {
      "avatar": {
        "href": "https://bitbucket.org/workspaces/martinherren1984/avatar/?ts=1658761964"
      },
      "html": {
        "href": "https://bitbucket.org/martinherren1984/"
      },
      "self": {
        "href": "https://api.bitbucket.org/2.0/workspaces/martinherren1984"
      }
    }
  },
  "is_private": false,
  "project": {
    "type": "project",
    "key": "PUB",
    "uuid": "{2cede481-f59e-49ec-88d0-a85629b7925d}",
    "name": "PublicTestProject",
    "links": {
      "self": {
        "href": "https://api.bitbucket.org/2.0/workspaces/martinherren1984/projects/PUB"
      },
      "html": {
        "href": "https://bitbucket.org/martinherren1984/workspace/projects/PUB"
      },
      "avatar": {
        "href": "https://bitbucket.org/martinherren1984/workspace/projects/PUB/avatar/32?ts=1658768453"
      }
    }
  },
  "fork_policy": "allow_forks",
  "created_on": "2022-07-25T17:01:20.950706+00:00",
  "updated_on": "2022-09-07T20:19:30.622886+00:00",
  "size": 85955,
  "language": "",
  "uuid": "{898477b2-a080-4089-b385-597a783db392}",
  "mainbranch": {
    "name": "master",
    "type": "branch"
  },
  "override_settings": {
    "default_merge_strategy": true,
    "branching_model": true
  },
  "parent": null,
  "enforced_signed_commits": null,
  "has_issues": false,
  "has_wiki": false
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
            "hash": "branch_head_name",
						"links": {
							"html": {
								"href": "https://bitbucket.org/commitlink"
							}
						}
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
	"uuid": "{4d8c0f46-cd62-4b77-b0cf-faa3e4d932c6}",
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

const permissionsPayLoad = `
{
  "pagelen": 100,
	"page": 1,
  "values": [
    {
      "repository": {
        "full_name": "test_name/repo_name"
      },
      "permission": "read"
    },
		{
      "repository": {
        "full_name": "test_name/permission_read"
      },
      "permission": "read"
    },
		{
      "repository": {
        "full_name": "test_name/permission_write"
      },
      "permission": "write"
    },
		{
      "repository": {
        "full_name": "test_name/permission_admin"
      },
      "permission": "admin"
    }
  ]
}
`
