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
	"net/http"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
)

// GetRegistry
//
//	@Summary	Get a registry by name
//	@Router		/repos/{repo_id}/registries/{registry} [get]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Repository registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		registry		path	string	true	"the registry name"
func GetRegistry(c *gin.Context) {
	repo := session.Repo(c)
	addr := c.Param("registry")

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	registry, err := registryService.RegistryFind(repo, addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// PostRegistry
//
//	@Summary	Create a registry
//	@Router		/repos/{repo_id}/registries [post]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Repository registries
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int			true	"the repository id"
//	@Param		registry		body	Registry	true	"the new registry data"
func PostRegistry(c *gin.Context) {
	repo := session.Repo(c)

	in := new(model.Registry)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}
	registry := &model.Registry{
		RepoID:   repo.ID,
		Address:  in.Address,
		Username: in.Username,
		Password: in.Password,
	}
	if err := registry.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error inserting registry. %s", err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	if err := registryService.RegistryCreate(repo, registry); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting registry %q. %s", in.Address, err)
		return
	}
	c.JSON(http.StatusOK, in.Copy())
}

// PatchRegistry
//
//	@Summary	Update a registry by name
//	@Router		/repos/{repo_id}/registries/{registry} [patch]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Repository registries
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int			true	"the repository id"
//	@Param		registry		path	string		true	"the registry name"
//	@Param		registryData	body	Registry	true	"the attributes for the registry"
func PatchRegistry(c *gin.Context) {
	repo := session.Repo(c)
	addr := c.Param("registry")

	in := new(model.Registry)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing request. %s", err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	registry, err := registryService.RegistryFind(repo, addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if in.Username != "" {
		registry.Username = in.Username
	}
	if in.Password != "" {
		registry.Password = in.Password
	}

	if err := registry.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error updating registry. %s", err)
		return
	}
	if err := registryService.RegistryUpdate(repo, registry); err != nil {
		c.String(http.StatusInternalServerError, "Error updating registry %q. %s", in.Address, err)
		return
	}
	c.JSON(http.StatusOK, in.Copy())
}

// GetRegistryList
//
//	@Summary	List registries
//	@Router		/repos/{repo_id}/registries [get]
//	@Produce	json
//	@Success	200	{array}	Registry
//	@Tags		Repository registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetRegistryList(c *gin.Context) {
	repo := session.Repo(c)
	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	list, err := registryService.RegistryList(repo, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting registry list. %s", err)
		return
	}
	// copy the registry detail to remove the sensitive
	// password and token fields.
	for i, registry := range list {
		list[i] = registry.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// DeleteRegistry
//
//	@Summary	Delete a registry by name
//	@Router		/repos/{repo_id}/registries/{registry} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Repository registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		repo_id			path	int		true	"the repository id"
//	@Param		registry		path	string	true	"the registry name"
func DeleteRegistry(c *gin.Context) {
	repo := session.Repo(c)
	addr := c.Param("registry")

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	err := registryService.RegistryDelete(repo, addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
