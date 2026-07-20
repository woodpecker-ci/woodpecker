// Copyright 2026 Woodpecker Authors
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
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// GetParameterList
//
//	@Summary		List parameters of a repository
//	@Description	The list is used by the repo settings and to render typed inputs in the manual pipeline run form.
//	@Router			/repos/{repo_id}/parameters [get]
//	@Produce		json
//	@Success		200	{array}	Parameter
//	@Tags			Repository parameters
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			repo_id			path	int		true	"the repository id"
//	@Param			page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param			perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetParameterList(c *gin.Context) {
	repo := session.Repo(c)
	list, err := store.FromContext(c).ParameterList(repo, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting parameter list. %s", err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetParameter
//
//	@Summary	Get a parameter
//	@Router		/repos/{repo_id}/parameters/{parameter} [get]
//	@Produce	json
//	@Success	200	{object}	Parameter
//	@Tags		Repository parameters
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		parameter		path	int		true	"the parameter id"
func GetParameter(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing parameter id. %s", err)
		return
	}

	parameter, err := store.FromContext(c).ParameterFind(repo, id)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, parameter)
}

// PostParameter
//
//	@Summary	Create a parameter
//	@Router		/repos/{repo_id}/parameters [post]
//	@Produce	json
//	@Success	200	{object}	Parameter
//	@Tags		Repository parameters
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int			true	"the repository id"
//	@Param		parameter		body	Parameter	true	"the new parameter"
func PostParameter(c *gin.Context) {
	repo := session.Repo(c)
	_store := store.FromContext(c)

	in := new(model.Parameter)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	parameter := &model.Parameter{
		RepoID:      repo.ID,
		Name:        strings.TrimSpace(in.Name),
		Type:        in.Type,
		Description: in.Description,
		Default:     in.Default,
		Options:     in.Options,
		Required:    in.Required,
		Order:       in.Order,
		Source:      model.ParameterSourceRepoConfig,
	}
	if err := parameter.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting parameter. validate failed: %s", err)
		return
	}

	if err := _store.ParameterCreate(parameter); err != nil {
		if errors.Is(err, types.ErrInsertDuplicateDetected) {
			c.String(http.StatusConflict, "parameter with this name exists for this repo already")
		} else {
			c.String(http.StatusInternalServerError, "Error inserting parameter %q. %s", in.Name, err)
		}
		return
	}
	c.JSON(http.StatusOK, parameter)
}

// PatchParameter
//
//	@Summary	Update a parameter
//	@Router		/repos/{repo_id}/parameters/{parameter} [patch]
//	@Produce	json
//	@Success	200	{object}	Parameter
//	@Tags		Repository parameters
//	@Param		Authorization	header	string			true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int				true	"the repository id"
//	@Param		parameter		path	int				true	"the parameter id"
//	@Param		parameterData	body	ParameterPatch	true	"the parameter data"
func PatchParameter(c *gin.Context) {
	repo := session.Repo(c)
	_store := store.FromContext(c)

	id, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing parameter id. %s", err)
		return
	}

	in := new(model.ParameterPatch)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	parameter, err := _store.ParameterFind(repo, id)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if in.Name != nil {
		parameter.Name = strings.TrimSpace(*in.Name)
	}
	if in.Type != nil {
		parameter.Type = *in.Type
	}
	if in.Description != nil {
		parameter.Description = *in.Description
	}
	if in.Default != nil {
		parameter.Default = *in.Default
	}
	if in.Options != nil {
		parameter.Options = in.Options
	}
	if in.Required != nil {
		parameter.Required = *in.Required
	}
	if in.Order != nil {
		parameter.Order = *in.Order
	}

	if err := parameter.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error updating parameter. validate failed: %s", err)
		return
	}
	if err := _store.ParameterUpdate(parameter); err != nil {
		c.String(http.StatusInternalServerError, "Error updating parameter %q. %s", parameter.Name, err)
		return
	}
	c.JSON(http.StatusOK, parameter)
}

// DeleteParameter
//
//	@Summary	Delete a parameter
//	@Router		/repos/{repo_id}/parameters/{parameter} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Repository parameters
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		parameter		path	int		true	"the parameter id"
func DeleteParameter(c *gin.Context) {
	repo := session.Repo(c)
	id, err := strconv.ParseInt(c.Param("parameter"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing parameter id. %s", err)
		return
	}
	if err := store.FromContext(c).ParameterDelete(repo, id); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
