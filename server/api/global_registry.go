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

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
)

// GetGlobalRegistryList
//
//	@Summary	List global registries
//	@Router		/registries [get]
//	@Produce	json
//	@Success	200	{array}	Registry
//	@Tags		Registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page				query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetGlobalRegistryList(c *gin.Context) {
	registryService := server.Config.Services.Manager.RegistryService()
	list, err := registryService.GlobalRegistryList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting global registry list. %s", err)
		return
	}
	// copy the registry detail to remove the sensitive
	// password and token fields.
	for i, registry := range list {
		list[i] = registry.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// GetGlobalRegistry
//
//	@Summary	Get a global registry by name
//	@Router		/registries/{registry} [get]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		registry			path	string	true	"the registry's name"
func GetGlobalRegistry(c *gin.Context) {
	addr := c.Param("registry")
	registryService := server.Config.Services.Manager.RegistryService()
	registry, err := registryService.GlobalRegistryFind(addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// PostGlobalRegistry
//
//	@Summary	Create a global registry
//	@Router		/registries [post]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		registry			body	Registry	true	"the registry object data"
func PostGlobalRegistry(c *gin.Context) {
	in := new(model.Registry)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing global registry. %s", err)
		return
	}
	registry := &model.Registry{
		Address:  in.Address,
		Username: in.Username,
		Password: in.Password,
	}
	if err := registry.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error inserting global registry. %s", err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryService()
	if err := registryService.GlobalRegistryCreate(registry); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting global registry %q. %s", in.Address, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// PatchGlobalRegistry
//
//	@Summary	Update a global registry by name
//	@Router		/registries/{registry} [patch]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		registry			path	string		true	"the registry's name"
//	@Param		registryData	body	Registry	true	"the registry's data"
func PatchGlobalRegistry(c *gin.Context) {
	addr := c.Param("registry")

	in := new(model.Registry)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing registry. %s", err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryService()
	registry, err := registryService.GlobalRegistryFind(addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if in.Address != "" {
		registry.Address = in.Address
	}
	if in.Username != "" {
		registry.Username = in.Username
	}
	if in.Password != "" {
		registry.Password = in.Password
	}

	if err := registry.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error updating global registry. %s", err)
		return
	}

	if err := registryService.GlobalRegistryUpdate(registry); err != nil {
		c.String(http.StatusInternalServerError, "Error updating global registry %q. %s", in.Address, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// DeleteGlobalRegistry
//
//	@Summary	Delete a global registry by name
//	@Router		/registries/{registry} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		registry			path		string	true	"the registry's name"
func DeleteGlobalRegistry(c *gin.Context) {
	addr := c.Param("registry")
	registryService := server.Config.Services.Manager.RegistryService()
	if err := registryService.GlobalRegistryDelete(addr); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
