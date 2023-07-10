// Copyright 2022 Woodpecker Authors
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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

// PostRepo
//
//	@Summary	Activate a repository
//	@Router		/repos/{repo_id} [post]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func PostRepo(c *gin.Context) {
	forge := server.Config.Services.Forge
	_store := store.FromContext(c)
	user := session.User(c)

	forgeRemoteID := model.ForgeRemoteID(c.Query("forge_remote_id"))
	repo, err := _store.GetRepoForgeID(forgeRemoteID)
	enabledOnce := err == nil // if there's no error, the repo was found and enabled once already
	if enabledOnce && repo.IsActive {
		c.String(http.StatusConflict, "Repository is already active.")
		return
	} else if err != nil && !errors.Is(err, types.RecordNotExist) {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	from, err := forge.Repo(c, user, forgeRemoteID, "", "")
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not fetch repository from forge.")
		return
	}
	if !from.Perm.Admin {
		c.String(http.StatusForbidden, "User has to be a admin of this repository")
	}

	if enabledOnce {
		repo.Update(from)
	} else {
		repo = from
		repo.AllowPull = true
		repo.NetrcOnlyTrusted = true
		repo.CancelPreviousPipelineEvents = server.Config.Pipeline.DefaultCancelPreviousPipelineEvents
	}
	repo.IsActive = true
	repo.UserID = user.ID

	if repo.Visibility == "" {
		repo.Visibility = model.VisibilityPublic
		if repo.IsSCMPrivate {
			repo.Visibility = model.VisibilityPrivate
		}
	}

	if repo.Timeout == 0 {
		repo.Timeout = server.Config.Pipeline.DefaultTimeout
	} else if repo.Timeout > server.Config.Pipeline.MaxTimeout {
		repo.Timeout = server.Config.Pipeline.MaxTimeout
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
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		server.Config.Server.WebhookHost,
		sig,
	)

	err = forge.Activate(c, user, repo, link)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if enabledOnce {
		err = _store.UpdateRepo(repo)
	} else {
		err = _store.CreateRepo(repo)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	repo.Perm = from.Perm
	repo.Perm.Synced = time.Now().Unix()
	repo.Perm.UserID = user.ID
	repo.Perm.RepoID = repo.ID
	repo.Perm.Repo = repo
	err = _store.PermUpsert(repo.Perm)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, repo)
}

// PatchRepo
//
//	@Summary	Change a repository
//	@Router		/repos/{repo_id} [patch]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string			true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		repo			body	RepoPatch	true	"the repository's information"
func PatchRepo(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	in := new(model.RepoPatch)
	if err := c.Bind(in); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if in.Timeout != nil && *in.Timeout > server.Config.Pipeline.MaxTimeout && !user.Admin {
		c.String(http.StatusForbidden, fmt.Sprintf("Timeout is not allowed to be higher than max timeout (%dmin)", server.Config.Pipeline.MaxTimeout))
		return
	}
	if in.IsTrusted != nil && *in.IsTrusted != repo.IsTrusted && !user.Admin {
		log.Trace().Msgf("user '%s' wants to make repo trusted without being an instance admin ", user.Login)
		c.String(http.StatusForbidden, "Insufficient privileges")
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
	if in.CancelPreviousPipelineEvents != nil {
		repo.CancelPreviousPipelineEvents = *in.CancelPreviousPipelineEvents
	}
	if in.NetrcOnlyTrusted != nil {
		repo.NetrcOnlyTrusted = *in.NetrcOnlyTrusted
	}
	if in.Visibility != nil {
		switch *in.Visibility {
		case string(model.VisibilityInternal), string(model.VisibilityPrivate), string(model.VisibilityPublic):
			repo.Visibility = model.RepoVisibility(*in.Visibility)
		default:
			c.String(http.StatusBadRequest, "Invalid visibility type")
			return
		}
	}

	err := _store.UpdateRepo(repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, repo)
}

// ChownRepo
//
//	@Summary	Change a repository's owner, to the one holding the access token
//	@Router		/repos/{repo_id}/chown [post]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func ChownRepo(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)
	repo.UserID = user.ID

	err := _store.UpdateRepo(repo)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, repo)
}

// LookupRepo
//
//	@Summary	Get repository by full-name
//	@Router		/repos/lookup/{repo_full_name} [get]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_full_name	path	string	true	"the repository full-name / slug"
func LookupRepo(c *gin.Context) {
	_store := store.FromContext(c)
	repoFullName := strings.TrimLeft(c.Param("repo_full_name"), "/")

	repo, err := _store.GetRepoName(repoFullName)
	if err != nil {
		if errors.Is(err, types.RecordNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, repo)
}

// GetRepo
//
//	@Summary	Get repository information
//	@Router		/repos/{repo_id} [get]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func GetRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}

// GetRepoPermissions
//
//	@Summary		Repository permission information
//	@Description	The repository permission, according to the used access token.
//	@Router			/repos/{repo_id}/permissions [get]
//	@Produce		json
//	@Success		200	{object}	Perm
//	@Tags			Repositories
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			owner			path	string	true	"the repository owner's name"
//	@Param			name			path	string	true	"the repository name"
func GetRepoPermissions(c *gin.Context) {
	perm := session.Perm(c)
	c.JSON(http.StatusOK, perm)
}

// GetRepoBranches
//
//	@Summary	Get repository branches
//	@Router		/repos/{repo_id}/branches [get]
//	@Produce	json
//	@Success	200	{array}	string
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetRepoBranches(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	f := server.Config.Services.Forge

	branches, err := f.Branches(c, user, repo, session.Pagination(c))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, branches)
}

// GetRepoPullRequests
//
//	@Summary	List active pull requests
//	@Router		/repos/{repo_id}/pull_requests [get]
//	@Produce	json
//	@Success	200	{array}	PullRequest
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetRepoPullRequests(c *gin.Context) {
	repo := session.Repo(c)
	user := session.User(c)
	f := server.Config.Services.Forge

	prs, err := f.PullRequests(c, user, repo, session.Pagination(c))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, prs)
}

// DeleteRepo
//
//	@Summary	Delete a repository
//	@Router		/repos/{repo_id} [delete]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func DeleteRepo(c *gin.Context) {
	remove, _ := strconv.ParseBool(c.Query("remove"))
	_store := store.FromContext(c)

	repo := session.Repo(c)
	user := session.User(c)

	repo.IsActive = false
	repo.UserID = 0

	if err := _store.UpdateRepo(repo); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if remove {
		if err := _store.DeleteRepo(repo); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if err := server.Config.Services.Forge.Deactivate(c, user, repo, server.Config.Server.Host); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, repo)
}

// RepairRepo
//
//	@Summary	Repair a repository
//	@Router		/repos/{repo_id}/repair [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func RepairRepo(c *gin.Context) {
	forge := server.Config.Services.Forge
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// reconstruct the link
	host := server.Config.Server.Host
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		host,
		sig,
	)

	from, err := forge.Repo(c, user, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		log.Error().Err(err).Msgf("get repo '%s/%s' from forge", repo.Owner, repo.Name)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if repo.FullName != from.FullName {
		// create a redirection
		err = _store.CreateRedirection(&model.Redirection{RepoID: repo.ID, FullName: repo.FullName})
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	repo.Update(from)
	if err := _store.UpdateRepo(repo); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	repo.Perm.Pull = from.Perm.Pull
	repo.Perm.Push = from.Perm.Push
	repo.Perm.Admin = from.Perm.Admin
	if err := _store.PermUpsert(repo.Perm); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := forge.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err).Msgf("deactivate repo '%s' to repair failed", repo.FullName)
	}
	if err := forge.Activate(c, user, repo, link); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

// MoveRepo
//
//	@Summary	Move a repository to a new owner
//	@Router		/repos/{repo_id}/move [post]
//	@Produce	plain
//	@Success	200
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		to				query	string	true	"the username to move the repository to"
func MoveRepo(c *gin.Context) {
	forge := server.Config.Services.Forge
	_store := store.FromContext(c)
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

	from, err := forge.Repo(c, user, "", owner, name)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !from.Perm.Admin {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = _store.CreateRedirection(&model.Redirection{RepoID: repo.ID, FullName: repo.FullName})
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	repo.Update(from)
	errStore := _store.UpdateRepo(repo)
	if errStore != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errStore)
		return
	}
	repo.Perm = from.Perm
	errStore = _store.PermUpsert(repo.Perm)
	if errStore != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errStore)
		return
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// reconstruct the link
	host := server.Config.Server.Host
	link := fmt.Sprintf(
		"%s/hook?access_token=%s",
		host,
		sig,
	)

	if err := forge.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err).Msgf("deactivate repo '%s' for move to activate later, got an error", repo.FullName)
	}
	if err := forge.Activate(c, user, repo, link); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
