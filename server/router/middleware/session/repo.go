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

package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func Repo(c *gin.Context) *model.Repo {
	v, ok := c.Get("repo")
	if !ok {
		return nil
	}
	r, ok := v.(*model.Repo)
	if !ok {
		return nil
	}
	return r
}

func SetRepo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			_store = store.FromContext(c)
			owner  = c.Param("owner")
			name   = c.Param("name")
			user   = User(c)
		)

		repo, err := _store.GetRepoName(owner + "/" + name)
		if err == nil {
			c.Set("repo", repo)
			c.Next()
			return
		}

		// debugging
		log.Debug().Msgf("Cannot find repository %s/%s. %s",
			owner,
			name,
			err.Error(),
		)

		if user != nil {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func Perm(c *gin.Context) *model.Perm {
	v, ok := c.Get("perm")
	if !ok {
		return nil
	}
	u, ok := v.(*model.Perm)
	if !ok {
		return nil
	}
	return u
}

func SetPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		_store := store.FromContext(c)
		user := User(c)
		repo := Repo(c)
		perm := new(model.Perm)

		switch {
		case user != nil:
			var err error
			perm, err = _store.PermFind(user, repo)
			if err != nil {
				log.Error().Msgf("Error fetching permission for %s %s. %s",
					user.Login, repo.FullName, err)
			}
			if time.Unix(perm.Synced, 0).Add(time.Hour).Before(time.Now()) {
				perm, err = server.Config.Services.Remote.Perm(c, user, repo)
				if err == nil {
					log.Debug().Msgf("Synced user permission for %s %s", user.Login, repo.FullName)
					perm.Repo = repo
					perm.UserID = user.ID
					perm.Synced = time.Now().Unix()
					if err := _store.PermUpsert(perm); err != nil {
						_ = c.AbortWithError(http.StatusInternalServerError, err)
						return
					}
				}
			}
		}

		if perm == nil {
			perm = new(model.Perm)
		}

		if user != nil && user.Admin {
			perm.Pull = true
			perm.Push = true
			perm.Admin = true
		}

		switch {
		case repo.Visibility == model.VisibilityPublic:
			perm.Pull = true
		case repo.Visibility == model.VisibilityInternal && user != nil:
			perm.Pull = true
		}

		if user != nil {
			log.Debug().Msgf("%s granted %+v permission to %s",
				user.Login, perm, repo.FullName)
		} else {
			log.Debug().Msgf("Guest granted %+v to %s", perm, repo.FullName)
		}

		c.Set("perm", perm)
		c.Next()
	}
}

func MustPull(c *gin.Context) {
	user := User(c)
	perm := Perm(c)

	if perm.Pull {
		c.Next()
		return
	}

	// debugging
	if user != nil {
		c.AbortWithStatus(http.StatusNotFound)
		log.Debug().Msgf("User %s denied read access to %s",
			user.Login, c.Request.URL.Path)
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debug().Msgf("Guest denied read access to %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}

func MustPush(c *gin.Context) {
	user := User(c)
	perm := Perm(c)

	// if the user has push access, immediately proceed
	// the middleware execution chain.
	if perm.Push {
		c.Next()
		return
	}

	// debugging
	if user != nil {
		c.AbortWithStatus(http.StatusNotFound)
		log.Debug().Msgf("User %s denied write access to %s",
			user.Login, c.Request.URL.Path)
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debug().Msgf("Guest denied write access to %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}
