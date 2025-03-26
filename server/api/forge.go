// Copyright 2024 Woodpecker Authors
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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// GetForges
//
//	@Summary	List forges
//	@Router		/forges [get]
//	@Produce	json
//	@Success	200	{array}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string	false	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetForges(c *gin.Context) {
	forges, err := store.FromContext(c).ForgeList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting forge list. %s", err)
		return
	}

	user := session.User(c)
	if user != nil && user.Admin {
		c.JSON(http.StatusOK, forges)
		return
	}

	// copy forges data without sensitive information
	for i, forge := range forges {
		forges[i] = forge.PublicCopy()
	}

	c.JSON(http.StatusOK, forges)
}

// GetForge
//
//	@Summary	Get a forge
//	@Router		/forges/{forgeId} [get]
//	@Produce	json
//	@Success	200	{object}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string	false	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forgeId			path	int		true	"the forge's id"
func GetForge(c *gin.Context) {
	forgeID, err := strconv.ParseInt(c.Param("forgeId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := store.FromContext(c).ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	user := session.User(c)
	if user != nil && user.Admin {
		c.JSON(http.StatusOK, forge)
	} else {
		c.JSON(http.StatusOK, forge.PublicCopy())
	}
}

// PatchForge
//
//	@Summary	Update a forge
//	@Router		/forges/{forgeId} [patch]
//	@Produce	json
//	@Success	200	{object}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forgeId			path	int		true	"the forge's id"
//	@Param		forgeData		body	Forge	true	"the forge's data"
func PatchForge(c *gin.Context) {
	_store := store.FromContext(c)

	// use this struct to allow updating the client secret
	type ForgeWithClientSecret struct {
		model.Forge
		ClientSecret string `json:"client_secret"`
	}

	in := &ForgeWithClientSecret{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	forgeID, err := strconv.ParseInt(c.Param("forgeId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := _store.ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	forge.URL = in.URL
	forge.Type = in.Type
	forge.Client = in.Client
	forge.OAuthHost = in.OAuthHost
	forge.SkipVerify = in.SkipVerify
	forge.AdditionalOptions = in.AdditionalOptions
	if in.ClientSecret != "" {
		forge.ClientSecret = in.ClientSecret
	}

	err = _store.ForgeUpdate(forge)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, forge)
}

// PostForge
//
//	@Summary		Create a new forge
//	@Description	Creates a new forge with a random token
//	@Router			/forges [post]
//	@Produce		json
//	@Success		200	{object}	Forge
//	@Tags			Forges
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			forge			body	Forge	true	"the forge's data (only 'name' and 'no_schedule' are read)"
func PostForge(c *gin.Context) {
	in := &model.Forge{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	forge := &model.Forge{
		URL:               in.URL,
		Type:              in.Type,
		Client:            in.Client,
		ClientSecret:      in.ClientSecret,
		OAuthHost:         in.OAuthHost,
		SkipVerify:        in.SkipVerify,
		AdditionalOptions: in.AdditionalOptions,
	}
	if err = store.FromContext(c).ForgeCreate(forge); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, forge)
}

// DeleteForge
//
//	@Summary	Delete a forge
//	@Router		/forges/{forgeId} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Forges
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forgeId			path	int		true	"the forge's id"
func DeleteForge(c *gin.Context) {
	_store := store.FromContext(c)

	forgeID, err := strconv.ParseInt(c.Param("forgeId"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := _store.ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if err = _store.ForgeDelete(forge); err != nil {
		c.String(http.StatusInternalServerError, "Error deleting user. %s", err)
		return
	}
	c.Status(http.StatusNoContent)
}
