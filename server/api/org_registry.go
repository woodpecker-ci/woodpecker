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

// GetOrgRegistry
//
//	@Summary	Get a organization registry by address
//	@Router		/orgs/{org_id}/registries/{registry} [get]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Organization registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		registry		path	string	true	"the registry's address"
func GetOrgRegistry(c *gin.Context) {
	org := session.Org(c)
	addr := c.Param("registry")

	registryService := server.Config.Services.Manager.RegistryService()
	registry, err := registryService.OrgRegistryFind(org.ID, addr)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// GetOrgRegistryList
//
//	@Summary	List organization registries
//	@Router		/orgs/{org_id}/registries [get]
//	@Produce	json
//	@Success	200	{array}	Registry
//	@Tags		Organization registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		page				query	int			false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int			false	"for response pagination, max items per page"	default(50)
func GetOrgRegistryList(c *gin.Context) {
	org := session.Org(c)

	registryService := server.Config.Services.Manager.RegistryService()
	list, err := registryService.OrgRegistryList(org.ID, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting registry list for %q. %s", org.ID, err)
		return
	}
	// copy the registry detail to remove the sensitive
	// password and token fields.
	for i, registry := range list {
		list[i] = registry.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// PostOrgRegistry
//
//	@Summary	Create an organization registry
//	@Router		/orgs/{org_id}/registries [post]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Organization registries
//	@Param		Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id					path	string		true	"the org's id"
//	@Param		registryData		body	Registry	true	"the new registry"
func PostOrgRegistry(c *gin.Context) {
	org := session.Org(c)

	in := new(model.Registry)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing org %q registry. %s", org.ID, err)
		return
	}
	registry := &model.Registry{
		OrgID:    org.ID,
		Address:  in.Address,
		Username: in.Username,
		Password: in.Password,
	}
	if err := registry.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting org %q registry. %s", org.ID, err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryService()
	if err := registryService.OrgRegistryCreate(org.ID, registry); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting org %q registry %q. %s", org.ID, in.Address, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// PatchOrgRegistry
//
//	@Summary	Update an organization registry by name
//	@Router		/orgs/{org_id}/registries/{registry} [patch]
//	@Produce	json
//	@Success	200	{object}	Registry
//	@Tags		Organization registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id				path	string		true	"the org's id"
//	@Param		registry			path	string		true	"the registry's name"
//	@Param		registryData	body	Registry	true	"the update registry data"
func PatchOrgRegistry(c *gin.Context) {
	org := session.Org(c)
	addr := c.Param("registry")

	in := new(model.Registry)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing registry. %s", err)
		return
	}

	registryService := server.Config.Services.Manager.RegistryService()
	registry, err := registryService.OrgRegistryFind(org.ID, addr)
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
		c.String(http.StatusUnprocessableEntity, "Error updating org %q registry. %s", org.ID, err)
		return
	}

	if err := registryService.OrgRegistryUpdate(org.ID, registry); err != nil {
		c.String(http.StatusInternalServerError, "Error updating org %q registry %q. %s", org.ID, in.Address, err)
		return
	}
	c.JSON(http.StatusOK, registry.Copy())
}

// DeleteOrgRegistry
//
//	@Summary	Delete an organization registry by name
//	@Router		/orgs/{org_id}/registries/{registry} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Organization registries
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		registry		path	string	true	"the registry's name"
func DeleteOrgRegistry(c *gin.Context) {
	org := session.Org(c)
	addr := c.Param("registry")

	registryService := server.Config.Services.Manager.RegistryService()
	if err := registryService.OrgRegistryDelete(org.ID, addr); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
