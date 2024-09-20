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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

// PostRepo
//
//	@Summary	Activate a repository
//	@Router		/repos [post]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forge_remote_id	query	string	true	"the id of a repository at the forge"
func PostRepo(c *gin.Context) {
	_store := store.FromContext(c)
	user := session.User(c)
	_forge, err := server.Config.Services.Manager.ForgeFromUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from user")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	forgeRemoteID := model.ForgeRemoteID(c.Query("forge_remote_id"))
	if !forgeRemoteID.IsValid() {
		c.String(http.StatusBadRequest, "No forge_remote_id provided")
		return
	}

	repo, err := _store.GetRepoForgeID(forgeRemoteID)
	enabledOnce := err == nil // if there's no error, the repo was found and enabled once already
	if enabledOnce && repo.IsActive {
		c.String(http.StatusConflict, "Repository is already active.")
		return
	} else if err != nil && !errors.Is(err, types.RecordNotExist) {
		msg := "could not get repo by remote id from store."
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	from, err := _forge.Repo(c, user, forgeRemoteID, "", "")
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not fetch repository from forge.")
		return
	}
	if !from.Perm.Admin {
		c.String(http.StatusForbidden, "User has to be a admin of this repository")
		return
	}
	if !server.Config.Permissions.OwnersAllowlist.IsAllowed(from) {
		c.String(http.StatusForbidden, "Repo owner is not allowed")
		return
	}

	if enabledOnce {
		repo.Update(from)
	} else {
		repo = from
		repo.AllowPull = true
		repo.AllowDeploy = false
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

	// find org of repo
	var org *model.Org
	org, err = _store.OrgFindByName(repo.Owner)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// create an org if it doesn't exist yet
	if errors.Is(err, types.RecordNotExist) {
		org, err = _forge.Org(c, user, repo.Owner)
		if err != nil {
			msg := "could not fetch organization from forge."
			log.Error().Err(err).Msg(msg)
			c.String(http.StatusInternalServerError, msg)
			return
		}

		org.ForgeID = user.ForgeID
		err = _store.OrgCreate(org)
		if err != nil {
			msg := "could not create organization in store."
			log.Error().Err(err).Msg(msg)
			c.String(http.StatusInternalServerError, msg)
			return
		}
	}

	repo.OrgID = org.ID

	if enabledOnce {
		err = _store.UpdateRepo(repo)
	} else {
		repo.ForgeID = user.ForgeID // TODO: allow to use other connected forges of the user
		err = _store.CreateRepo(repo)
	}
	if err != nil {
		msg := "could not create/update repo in store."
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken)
	t.Set("repo-id", strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		msg := "could not generate new jwt token."
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
		return
	}

	hookURL := fmt.Sprintf(
		"%s/api/hook?access_token=%s",
		server.Config.Server.WebhookHost,
		sig,
	)

	err = _forge.Activate(c, user, repo, hookURL)
	if err != nil {
		msg := "could not create webhook in forge."
		log.Error().Err(err).Msg(msg)
		c.String(http.StatusInternalServerError, msg)
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
//	@Summary	Update a repository
//	@Router		/repos/{repo_id} [patch]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int			true	"the repository id"
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
		c.String(http.StatusForbidden, fmt.Sprintf("Timeout is not allowed to be higher than max timeout (%d min)", server.Config.Pipeline.MaxTimeout))
		return
	}
	if in.IsTrusted != nil && *in.IsTrusted != repo.IsTrusted && !user.Admin {
		log.Trace().Msgf("user '%s' wants to make repo trusted without being an instance admin", user.Login)
		c.String(http.StatusForbidden, "Insufficient privileges")
		return
	}

	if in.AllowPull != nil {
		repo.AllowPull = *in.AllowPull
	}
	if in.AllowDeploy != nil {
		repo.AllowDeploy = *in.AllowDeploy
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
//	@Summary	Change a repository's owner to the currently authenticated user
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
//	@Summary	Lookup a repository by full name
//	@Router		/repos/lookup/{repo_full_name} [get]
//	@Produce	json
//	@Success	200	{object}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_full_name	path	string	true	"the repository full name / slug"
func LookupRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}

// GetRepo
//
//	@Summary	Get a repository
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
//	@Summary		Check current authenticated users access to the repository
//	@Description	The repository permission, according to the used access token.
//	@Router			/repos/{repo_id}/permissions [get]
//	@Produce		json
//	@Success		200	{object}	Perm
//	@Tags			Repositories
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			repo_id			path	int		true	"the repository id"
func GetRepoPermissions(c *gin.Context) {
	perm := session.Perm(c)
	c.JSON(http.StatusOK, perm)
}

// GetRepoBranches
//
//	@Summary	Get branches of a repository
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
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	branches, err := _forge.Branches(c, user, repo, session.Pagination(c))
	if err != nil {
		log.Error().Err(err).Msg("failed to load branches")
		c.String(http.StatusInternalServerError, "failed to load branches: %s", err)
		return
	}

	c.JSON(http.StatusOK, branches)
}

// GetRepoPullRequests
//
//	@Summary	List active pull requests of a repository
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
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	prs, err := _forge.PullRequests(c, user, repo, session.Pagination(c))
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
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	repo.IsActive = false
	repo.UserID = 0

	if err := _store.UpdateRepo(repo); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if remove {
		if err := _store.DeleteRepo(repo); err != nil {
			handleDBError(c, err)
			return
		}
	}

	if err := _forge.Deactivate(c, user, repo, server.Config.Server.WebhookHost); err != nil {
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
//	@Success	204
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
func RepairRepo(c *gin.Context) {
	repo := session.Repo(c)
	repairRepo(c, repo, true, false)
	if c.Writer.Written() {
		return
	}
	c.Status(http.StatusNoContent)
}

// MoveRepo
//
//	@Summary	Move a repository to a new owner
//	@Router		/repos/{repo_id}/move [post]
//	@Produce	plain
//	@Success	204
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		to				query	string	true	"the username to move the repository to"
func MoveRepo(c *gin.Context) {
	_store := store.FromContext(c)
	repo := session.Repo(c)
	user := session.User(c)
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	to, exists := c.GetQuery("to")
	if !exists {
		err := fmt.Errorf("missing required to query value")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	owner, name, errParse := model.ParseRepo(to)
	if errParse != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errParse)
		return
	}

	from, err := _forge.Repo(c, user, "", owner, name)
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
	t := token.New(token.HookToken)
	t.Set("repo-id", strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// reconstruct the hook url
	host := server.Config.Server.WebhookHost
	hookURL := fmt.Sprintf(
		"%s/api/hook?access_token=%s",
		host,
		sig,
	)

	if err := _forge.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err).Msgf("deactivate repo '%s' for move to activate later, got an error", strconv.FormatInt(repo.ID, 10))
	}
	if err := _forge.Activate(c, user, repo, hookURL); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// GetAllRepos
//
//	@Summary	List all repositories on the server
//	@Description	Returns a list of all repositories. Requires admin rights.
//	@Router		/repos [get]
//	@Produce	json
//	@Success	200	{array}	Repo
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		active			query	bool	false	"only list active repos"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetAllRepos(c *gin.Context) {
	_store := store.FromContext(c)

	active, _ := strconv.ParseBool(c.Query("active"))

	repos, err := _store.RepoListAll(active, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	c.JSON(http.StatusOK, repos)
}

// RepairAllRepos
//
//	@Summary	Repair all repositories on the server
//	@Description Executes a repair process on all repositories. Requires admin rights.
//	@Router		/repos/repair [post]
//	@Produce	plain
//	@Success	204
//	@Tags		Repositories
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func RepairAllRepos(c *gin.Context) {
	_store := store.FromContext(c)

	repos, err := _store.RepoListAll(true, &model.ListOptions{All: true})
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	for _, r := range repos {
		repairRepo(c, r, false, true)
		if c.Writer.Written() {
			return
		}
	}

	c.Status(http.StatusNoContent)
}

func repairRepo(c *gin.Context, repo *model.Repo, withPerms, skipOnErr bool) {
	_store := store.FromContext(c)
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from repo")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		if errors.Is(err, types.RecordNotExist) {
			oldUserID := repo.UserID
			user = session.User(c)
			repo.UserID = user.ID
			err = _store.UpdateRepo(repo)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			log.Debug().Msgf("Could not find repo user with ID %d during repo repair, set to repair request user with ID %d", oldUserID, user.ID)
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	// creates the jwt token used to verify the repository
	t := token.New(token.HookToken)
	t.Set("repo-id", strconv.FormatInt(repo.ID, 10))
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// reconstruct the hook url
	host := server.Config.Server.WebhookHost
	hookURL := fmt.Sprintf(
		"%s/api/hook?access_token=%s",
		host,
		sig,
	)

	from, err := _forge.Repo(c, user, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		log.Error().Err(err).Msgf("get repo '%s/%s' from forge", repo.Owner, repo.Name)
		if !skipOnErr {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
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
	if withPerms {
		repo.Perm.Pull = from.Perm.Pull
		repo.Perm.Push = from.Perm.Push
		repo.Perm.Admin = from.Perm.Admin
		if err := _store.PermUpsert(repo.Perm); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if err := _forge.Deactivate(c, user, repo, host); err != nil {
		log.Trace().Err(err).Msgf("deactivate repo '%s' to repair failed", repo.FullName)
	}
	if err := _forge.Activate(c, user, repo, hookURL); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
}
