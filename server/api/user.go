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

package api

import (
	"encoding/base32"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

// GetSelf
//
//	@Summary	Get the currently authenticated user
//	@Router		/user [get]
//	@Produce	json
//	@Success	200	{object}	User
//	@Tags		User
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func GetSelf(c *gin.Context) {
	c.JSON(http.StatusOK, session.User(c))
}

// GetFeed
//
//	@Summary		Get the currently authenticated users pipeline feed
//	@Description	The feed lists the most recent pipeline for the currently authenticated user.
//	@Router			/user/feed [get]
//	@Produce		json
//	@Success		200	{array}	Feed
//	@Tags			User
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func GetFeed(c *gin.Context) {
	_store := store.FromContext(c)

	user := session.User(c)
	latest, _ := strconv.ParseBool(c.Query("latest"))

	if latest {
		feed, err := _store.RepoListLatest(user)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching feed. %s", err)
		} else {
			c.JSON(http.StatusOK, feed)
		}
		return
	}

	feed, err := _store.UserFeed(user)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching user feed. %s", err)
		return
	}
	c.JSON(http.StatusOK, feed)
}

// GetRepos
//
//	@Summary		Get user's repositories
//	@Description	Retrieve the currently authenticated User's Repository list
//	@Router			/user/repos [get]
//	@Produce		json
//	@Success		200	{array}	Repo
//	@Tags			User
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			all				query	bool	false	"query all repos, including inactive ones"
func GetRepos(c *gin.Context) {
	_store := store.FromContext(c)
	user := session.User(c)
	_forge, err := server.Config.Services.Manager.ForgeFromUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from user")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	all, _ := strconv.ParseBool(c.Query("all"))

	if all {
		dbRepos, err := _store.RepoList(user, true, false)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
			return
		}

		active := map[model.ForgeRemoteID]*model.Repo{}
		for _, r := range dbRepos {
			active[r.ForgeRemoteID] = r
		}

		_repos, err := _forge.Repos(c, user)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
			return
		}

		var repos []*model.Repo
		for _, r := range _repos {
			if r.Perm.Push && server.Config.Permissions.OwnersAllowlist.IsAllowed(r) {
				if active[r.ForgeRemoteID] != nil {
					existingRepo := active[r.ForgeRemoteID]
					existingRepo.Update(r)
					existingRepo.IsActive = active[r.ForgeRemoteID].IsActive
					repos = append(repos, existingRepo)
				} else if r.Perm.Admin {
					// you must be admin to enable the repo
					repos = append(repos, r)
				}
			}
		}

		c.JSON(http.StatusOK, repos)
		return
	}

	activeRepos, err := _store.RepoList(user, true, true)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	c.JSON(http.StatusOK, activeRepos)
}

// PostToken
//
//	@Summary	Return the token of the current user as string
//	@Router		/user/token [post]
//	@Produce	plain
//	@Success	200
//	@Tags		User
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func PostToken(c *gin.Context) {
	user := session.User(c)
	t := token.New(token.UserToken)
	t.Set("user-id", strconv.FormatInt(user.ID, 10))
	tokenString, err := t.Sign(user.Hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, tokenString)
}

// DeleteToken
//
//	@Summary		Reset a token
//	@Description	Reset's the current personal access token of the user and returns a new one.
//	@Router			/user/token [delete]
//	@Produce		plain
//	@Success		200
//	@Tags			User
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func DeleteToken(c *gin.Context) {
	_store := store.FromContext(c)

	user := session.User(c)
	user.Hash = base32.StdEncoding.EncodeToString(
		securecookie.GenerateRandomKey(32),
	)
	if err := _store.UpdateUser(user); err != nil {
		c.String(http.StatusInternalServerError, "Error revoking tokens. %s", err)
		return
	}

	t := token.New(token.UserToken)
	t.Set("user-id", strconv.FormatInt(user.ID, 10))
	tokenString, err := t.Sign(user.Hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, tokenString)
}
