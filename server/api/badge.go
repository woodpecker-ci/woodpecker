// Copyright 2022 Woodpecker Authors
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/badges"
	"go.woodpecker-ci.org/woodpecker/v3/server/ccmenu"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// badgeUnavailableLabel is shown whenever the state of a repo may not be
// served. It is used both for repos that do not exist and for repos the caller
// holds no badge token for, so the endpoints do not tell anonymous callers
// which repos exist.
const badgeUnavailableLabel = "repo does not exist or token missing"

// badgeDefaultName is the subject a badge carries if it shows the state of a
// whole pipeline rather than that of a single workflow or step.
const badgeDefaultName = "pipeline"

// resolveBadgeRepo returns the repo addressed by the request, or nil if it does
// not exist or its state may not be served. Public repos are served to
// everyone, every other repo requires the badge token of that repo as `token`
// query parameter.
func resolveBadgeRepo(c *gin.Context) *model.Repo {
	_store := store.FromContext(c)

	var repo *model.Repo
	var err error

	if c.Param("repo_name") != "" {
		repo, err = _store.GetRepoName(c.Param("repo_id_or_owner") + "/" + c.Param("repo_name"))
	} else {
		repoID, parseErr := strconv.ParseInt(c.Param("repo_id_or_owner"), 10, 64)
		if parseErr != nil {
			return nil
		}
		repo, err = _store.GetRepo(repoID)
	}

	if err != nil {
		if !errors.Is(err, types.ErrRecordNotExist) {
			log.Error().Err(err).Msg("could not get repo for badge")
		}
		return nil
	}

	if repo.Visibility == model.VisibilityPublic {
		return repo
	}

	token, err := _store.TokenFind(repo, model.TokenTypeBadge)
	if err != nil {
		if !errors.Is(err, types.ErrRecordNotExist) {
			log.Error().Err(err).Msg("could not get badge token")
		}
		return nil
	}

	// an empty stored value would match a request without a token at all
	if token.Value == "" || !token.Matches(c.Query("token")) {
		return nil
	}

	return repo
}

// GetBadge
//
//	@Summary	Get status of pipeline as SVG badge
//	@Router		/badges/{repo_id}/status.svg [get]
//	@Produce	image/svg+xml
//	@Success	200
//	@Tags		Badges
//	@Param		repo_id	path	int		true	"the repository id"
//	@Param		token	query	string	false	"the badge token of the repository, required for non-public repos"
func GetBadge(c *gin.Context) {
	_store := store.FromContext(c)

	// we serve an SVG, so set content type appropriately.
	c.Writer.Header().Set("Content-Type", "image/svg+xml")

	repo := resolveBadgeRepo(c)
	if repo == nil {
		writeUnavailableBadge(c)
		return
	}

	// if no commit was found then display
	// the 'none' badge, instead of throwing
	// an error response
	branch := c.Query("branch")
	if len(branch) == 0 {
		branch = repo.Branch
	}

	// Events to lookup, multiple separated by comma
	var events []model.WebhookEvent
	eventsQuery := c.Query("events")
	// If none given, fallback to default "push"
	if len(eventsQuery) == 0 {
		events = []model.WebhookEvent{model.EventPush}
	} else {
		strEvents := strings.Split(eventsQuery, ",")
		events = make([]model.WebhookEvent, len(strEvents))
		for i, strEvent := range strEvents {
			event := model.WebhookEvent(strEvent)
			if err := event.Validate(); err == nil {
				events[i] = event
			} else {
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}
	}

	name := badgeDefaultName
	var status *model.StatusValue = nil

	pl, err := _store.GetPipelineBadge(repo, branch, events)
	if err != nil {
		if !errors.Is(err, types.ErrRecordNotExist) {
			log.Warn().Err(err).Msg("could not get last pipeline for badge")
		}
	} else {
		status = &pl.Status
	}

	// Allow workflow (and step) specific badges
	workflowName := c.Query("workflow")
	if len(workflowName) != 0 {
		name = workflowName
		status = nil

		workflows, err := _store.WorkflowGetTree(pl)
		if err == nil {
			for _, wf := range workflows {
				if wf.Name == workflowName {
					stepName := c.Query("step")
					if len(stepName) == 0 {
						if status != nil {
							merged := pipeline.MergeStatusValues(*status, wf.State)
							status = &merged
						} else {
							status = &wf.State
						}
						continue
					}
					// If step is explicitly requested
					name = workflowName + ": " + stepName
					for _, s := range wf.Children {
						if s.Name == stepName {
							if status != nil {
								merged := pipeline.MergeStatusValues(*status, s.State)
								status = &merged
							} else {
								status = &s.State
							}
						}
					}
				}
			}
		}
	}

	writeBadge(c, name, status)
}

// writeBadge renders the badge for the given name and status and writes it to
// the response.
func writeBadge(c *gin.Context, name string, status *model.StatusValue) {
	badge, err := badges.Generate(name, status)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate badge.")
		return
	}
	c.String(http.StatusOK, badge)
}

// writeUnavailableBadge renders the badge shown for repos whose state may not
// be served. It states the reason instead of a status, as there is none to
// show.
func writeUnavailableBadge(c *gin.Context) {
	badge, err := badges.RenderBytes(badgeDefaultName, badgeUnavailableLabel, badges.ColorGray)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate badge.")
		return
	}
	c.String(http.StatusOK, string(badge))
}

// GetCC
//
//	@Summary		Provide pipeline status information to the CCMenu tool
//	@Description	CCMenu displays the pipeline status of projects on a CI server as an item in the Mac's menu bar.
//	@Description	More details on how to install, you can find at http://ccmenu.org/
//	@Description	The response format adheres to CCTray v1 Specification, https://cctray.org/v1/
//	@Router			/badges/{repo_id}/cc.xml [get]
//	@Produce		xml
//	@Success		200
//	@Tags			Badges
//	@Param			repo_id	path	int		true	"the repository id"
//	@Param			token	query	string	false	"the badge token of the repository, required for non-public repos"
func GetCC(c *gin.Context) {
	_store := store.FromContext(c)

	repo := resolveBadgeRepo(c)
	if repo == nil {
		c.XML(http.StatusOK, ccmenu.NewPlaceholder(badgeUnavailableLabel))
		return
	}

	pipelines, err := _store.GetPipelineList(repo, &model.ListOptionsWithAll{ListOptions: &model.ListOptions{Page: 1, PerPage: 1}}, nil)
	if err != nil && !errors.Is(err, types.ErrRecordNotExist) {
		log.Warn().Err(err).Msg("could not get pipeline list")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	// only reachable for repos the caller may see, so it leaks nothing
	if len(pipelines) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("%s/repos/%d/pipeline/%d", server.Config.Server.Host, repo.ID, pipelines[0].Number)
	cc := ccmenu.New(repo, pipelines[0], url)
	c.XML(http.StatusOK, cc)
}
