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

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/store"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func Repos(c *gin.Context) []*model.RepoLite {
	v, ok := c.Get("repos")
	if !ok {
		return nil
	}
	r, ok := v.([]*model.RepoLite)
	if !ok {
		return nil
	}
	return r
}

func SetRepo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			owner = c.Param("owner")
			name  = c.Param("name")
			user  = User(c)
		)

		repo, err := store.GetRepoOwnerName(c, owner, name)
		if err == nil {
			c.Set("repo", repo)
			c.Next()
			return
		}

		// debugging
		log.Debugf("Cannot find repository %s/%s. %s",
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
		user := User(c)
		repo := Repo(c)
		perm := new(model.Perm)

		switch {
		case user != nil:
			var err error
			perm, err = store.FromContext(c).PermFind(user, repo)
			if err != nil {
				log.Errorf("Error fetching permission for %s %s. %s",
					user.Login, repo.FullName, err)
			}
			if time.Unix(perm.Synced, 0).Add(time.Hour).Before(time.Now()) {
				perm, err = remote.FromContext(c).Perm(user, repo.Owner, repo.Name)
				if err == nil {
					log.Debugf("Synced user permission for %s %s", user.Login, repo.FullName)
					perm.Repo = repo.FullName
					perm.UserID = user.ID
					perm.Synced = time.Now().Unix()
					store.FromContext(c).PermUpsert(perm)
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
			log.Debugf("%s granted %+v permission to %s",
				user.Login, perm, repo.FullName)

		} else {
			log.Debugf("Guest granted %+v to %s", perm, repo.FullName)
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
		log.Debugf("User %s denied read access to %s",
			user.Login, c.Request.URL.Path)
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debugf("Guest denied read access to %s %s",
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
		log.Debugf("User %s denied write access to %s",
			user.Login, c.Request.URL.Path)

	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Debugf("Guest denied write access to %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}
