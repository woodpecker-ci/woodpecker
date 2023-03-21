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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

func GetSelf(c *gin.Context) {
	c.JSON(http.StatusOK, session.User(c))
}

func GetFeed(c *gin.Context) {
	_store := store.FromContext(c)
	_forge := server.Config.Services.Forge

	user := session.User(c)
	latest, _ := strconv.ParseBool(c.Query("latest"))

	if time.Unix(user.Synced, 0).Add(time.Hour * 72).Before(time.Now()) {
		log.Debug().Msgf("sync begin: %s", user.Login)

		user.Synced = time.Now().Unix()
		if err := _store.UpdateUser(user); err != nil {
			log.Error().Err(err).Msg("UpdateUser")
			return
		}

		config := ToConfig(c)

		sync := forge.Syncer{
			Forge: _forge,
			Store: _store,
			Perms: _store,
			Match: forge.NamespaceFilter(config.OwnersWhitelist),
		}
		if err := sync.Sync(c, user, server.Config.FlatPermissions); err != nil {
			log.Debug().Msgf("sync error: %s: %s", user.Login, err)
		} else {
			log.Debug().Msgf("sync complete: %s", user.Login)
		}
	}

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

func GetRepos(c *gin.Context) {
	_store := store.FromContext(c)
	_forge := server.Config.Services.Forge

	user := session.User(c)
	all, _ := strconv.ParseBool(c.Query("all"))
	flush, _ := strconv.ParseBool(c.Query("flush"))

	if flush || time.Unix(user.Synced, 0).Add(time.Hour*72).Before(time.Now()) {
		log.Debug().Msgf("sync begin: %s", user.Login)
		user.Synced = time.Now().Unix()
		if err := _store.UpdateUser(user); err != nil {
			log.Err(err).Msgf("update user '%s'", user.Login)
			return
		}

		config := ToConfig(c)

		sync := forge.Syncer{
			Forge: _forge,
			Store: _store,
			Perms: _store,
			Match: forge.NamespaceFilter(config.OwnersWhitelist),
		}

		if err := sync.Sync(c, user, server.Config.FlatPermissions); err != nil {
			log.Debug().Msgf("sync error: %s: %s", user.Login, err)
		} else {
			log.Debug().Msgf("sync complete: %s", user.Login)
		}
	}

	repos, err := _store.RepoList(user, true)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	if all {
		c.JSON(http.StatusOK, repos)
		return
	}

	active := make([]*model.Repo, 0)
	for _, repo := range repos {
		if repo.IsActive {
			active = append(active, repo)
		}
	}
	c.JSON(http.StatusOK, active)
}

func PostToken(c *gin.Context) {
	user := session.User(c)
	tokenString, err := token.New(token.UserToken, user.Login).Sign(user.Hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, tokenString)
}

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

	tokenString, err := token.New(token.UserToken, user.Login).Sign(user.Hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, tokenString)
}
