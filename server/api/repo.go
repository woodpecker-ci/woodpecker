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
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

func PostRepo(c *gin.Context) {
	remote_ := remote.FromContext(c)
	store_ := store.FromContext(c)
	user := session.User(c)
	repo := session.Repo(c)

	if repo.IsActive {
		c.String(409, "Repository is already active.")
		return
	}

	repo.IsActive = true
	repo.UserID = user.ID
	repo.AllowPull = true

	if repo.Visibility == "" {
		repo.Visibility = model.VisibilityPublic
		if repo.IsPrivate {
			repo.Visibility = model.VisibilityPrivate
		}
	}

	if repo.Timeout == 0 {
		repo.Timeout = 60 // 1 hour default build time
	}

	if repo.Hash == "" {
		repo.Hash = base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		server.Config.Server.Host,
		sig,
	)

	err = remote_.Activate(c, user, repo, link)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	from, err := remote_.Repo(c, user, repo.Owner, repo.Name)
	if err == nil {
		repo.Update(from)
	}

	err = store_.UpdateRepo(repo)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.JSON(200, repo)
}

func PatchRepo(c *gin.Context) {
	store_ := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	in := new(model.RepoPatch)
	if err := c.Bind(in); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if (in.IsTrusted != nil || in.Timeout != nil) && !user.Admin {
		c.String(403, "Insufficient privileges")
		return
	}

	if in.AllowPull != nil {
		repo.AllowPull = *in.AllowPull
	}
	if in.IsGated != nil {
		repo.IsGated = *in.IsGated
	}
	if in.IsTrusted != nil {
		repo.IsTrusted = *in.IsTrusted
	}
	if in.Timeout != nil {
		repo.Timeout = *in.Timeout
	}
	if in.Config != nil {
		repo.Config = *in.Config
	}
	if in.Visibility != nil {
		switch *in.Visibility {
		case model.VisibilityInternal, model.VisibilityPrivate, model.VisibilityPublic:
			repo.Visibility = *in.Visibility
		default:
			c.String(400, "Invalid visibility type")
			return
		}
	}
	if in.BuildCounter != nil {
		repo.Counter = *in.BuildCounter
	}

	err := store_.UpdateRepo(repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, repo)
}

func ChownRepo(c *gin.Context) {
	store_ := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)
	repo.UserID = user.ID

	err := store_.UpdateRepo(repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, repo)
}

func GetRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}

func GetRepoPermissions(c *gin.Context) {
	perm := session.Perm(c)
	c.JSON(http.StatusOK, perm)
}

func GetRepoBranches(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	r := remote.FromContext(c)

	branches, err := r.Branches(c, user, repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, branches)
}

func DeleteRepo(c *gin.Context) {
	remove, _ := strconv.ParseBool(c.Query("remove"))
	remote_ := remote.FromContext(c)
	store_ := store.FromContext(c)

	repo := session.Repo(c)
	user := session.User(c)

	repo.IsActive = false
	repo.UserID = 0

	err := store_.UpdateRepo(repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if remove {
		err := store_.DeleteRepo(repo)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if err := remote_.Deactivate(c, user, repo, server.Config.Server.Host); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, repo)
}

func RepairRepo(c *gin.Context) {
	remote_ := remote.FromContext(c)
	store_ := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// reconstruct the link
	host := server.Config.Server.Host
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		host,
		sig,
	)

	if err := remote_.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err)
	}
	if err := remote_.Activate(c, user, repo, link); err != nil {
		c.String(500, err.Error())
		return
	}

	from, err := remote_.Repo(c, user, repo.Owner, repo.Name)
	if err != nil {
		log.Error().Err(err).Msgf("get repo '%s/%s' from remote", repo.Owner, repo.Name)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	repo.Name = from.Name
	repo.Owner = from.Owner
	repo.FullName = from.FullName
	repo.Avatar = from.Avatar
	repo.Link = from.Link
	repo.Clone = from.Clone
	repo.IsPrivate = from.IsPrivate
	if repo.IsPrivate != from.IsPrivate {
		repo.ResetVisibility()
	}
	if err := store_.UpdateRepo(repo); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func MoveRepo(c *gin.Context) {
	remote_ := remote.FromContext(c)
	store_ := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	to, exists := c.GetQuery("to")
	if !exists {
		err := fmt.Errorf("Missing required to query value")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	owner, name, errParse := model.ParseRepo(to)
	if errParse != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errParse)
		return
	}

	from, err := remote_.Repo(c, user, owner, name)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !from.Perm.Admin {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	repo.Name = from.Name
	repo.Owner = from.Owner
	repo.FullName = from.FullName
	repo.Avatar = from.Avatar
	repo.Link = from.Link
	repo.Clone = from.Clone
	repo.IsPrivate = from.IsPrivate
	if repo.IsPrivate != from.IsPrivate {
		repo.ResetVisibility()
	}

	errStore := store_.UpdateRepo(repo)
	if errStore != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errStore)
		return
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// reconstruct the link
	host := server.Config.Server.Host
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		host,
		sig,
	)

	if err := remote_.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err).Msgf("deactivate repo '%s' for move to activate later, got an error", repo.FullName)
	}
	if err := remote_.Activate(c, user, repo, link); err != nil {
		c.String(500, err.Error())
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
