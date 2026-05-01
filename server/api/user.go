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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/tink-crypto/tink-go/v2/subtle/random"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
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
//	@Success		200	{array}	RepoLastPipeline
//	@Tags			User
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			all				query	bool	false	"query all repos, including inactive ones"
//	@Param			name			query	string	false	"filter repos by name"
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
	filter := &model.RepoFilter{
		Name: c.Query("name"),
	}

	if all {
		dbRepos, err := _store.RepoList(user, true, false, filter)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
			return
		}

		dbReposMap := map[model.ForgeRemoteID]*model.Repo{}
		dbStaleReposMap := map[int64]*model.Repo{}
		dbReposFullNameMap := map[string]*model.Repo{}
		for _, r := range dbRepos {
			dbReposMap[r.ForgeRemoteID] = r
			dbReposFullNameMap[strings.ToLower(r.FullName)] = r
			dbStaleReposMap[r.ID] = r
		}

		_repos, err := utils.Paginate(func(page int) ([]*model.Repo, error) {
			return _forge.Repos(c, user, &model.ListOptions{
				Page:    page,
				PerPage: perPage,
			})
		}, maxPage)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
			return
		}

		var repos []*model.Repo
		for _, r := range _repos {
			// make sure forgeID is set
			r.ForgeID = user.ForgeID

			if r.Perm.Push && server.Config.Permissions.OwnersAllowlist.IsAllowed(r) {
				if existingRepo := dbReposMap[r.ForgeRemoteID]; existingRepo != nil {
					// update repo with forge response
					existingRepo.Update(r)
					// re-apply active info
					existingRepo.IsActive = dbReposMap[r.ForgeRemoteID].IsActive
					// add to final return list
					repos = append(repos, existingRepo)
					// not stale, so remove it
					delete(dbStaleReposMap, existingRepo.ID)
				} else if r.Perm.Admin {
					// you must be admin of the remote repo to enable the repo
					repos = append(repos, r)
				}
			}
		}

		// detect conflicts
		for _, r := range repos {
			// calc if we have a remote repo with different remote id but same name as a stored one
			if existingRepo := dbReposFullNameMap[strings.ToLower(r.FullName)]; existingRepo != nil && existingRepo.ForgeRemoteID != r.ForgeRemoteID {
				r.ID = existingRepo.ID
				r.HasForgeNameConflict = true

				// not stale, so remove it
				delete(dbStaleReposMap, existingRepo.ID)
			}
		}

		// return stale repos
		for _, staleRepo := range dbStaleReposMap {
			staleRepo.HasNoForgeRepo = true
			repos = append(repos, staleRepo)
		}

		c.JSON(http.StatusOK, repos)
		return
	}

	activeRepos, err := _store.RepoList(user, true, true, filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	repoIDs := make([]int64, len(activeRepos))
	for i, repo := range activeRepos {
		repoIDs[i] = repo.ID
	}

	pipelines, err := _store.GetRepoLatestPipelines(repoIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching repository list. %s", err)
		return
	}

	latestPipelines := make(map[int64]*model.Pipeline, len(activeRepos))
	for _, pipeline := range pipelines {
		latestPipelines[pipeline.RepoID] = pipeline
	}

	repos := make([]*model.RepoLastPipeline, len(activeRepos))
	for i, repo := range activeRepos {
		var lastAPIPipeline *model.APIPipeline
		lastPipeline, ok := latestPipelines[repo.ID]
		if ok {
			lastAPIPipeline = lastPipeline.ToAPIModel()
		}

		repos[i] = &model.RepoLastPipeline{
			Repo:         repo,
			LastPipeline: lastAPIPipeline,
		}
	}

	c.JSON(http.StatusOK, repos)
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
		random.GetRandomBytes(32),
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
