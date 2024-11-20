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

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
)

// GetOrgVariable
//
//	@Summary	Get a organization variable by name
//	@Router		/orgs/{org_id}/variables/{variable} [get]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Organization variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		variable			path	string	true	"the variable's name"
func GetOrgVariable(c *gin.Context) {
	name := c.Param("variable")

	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	variable, err := variableService.OrgVariableFind(orgID, name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// GetOrgVariableList
//
//	@Summary	List organization variables
//	@Router		/orgs/{org_id}/variables [get]
//	@Produce	json
//	@Success	200	{array}	Variable
//	@Tags		Organization variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgVariableList(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	list, err := variableService.OrgVariableList(orgID, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting variable list for %q. %s", orgID, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// PostOrgVariable
//
//	@Summary	Create an organization variable
//	@Router		/orgs/{org_id}/variables [post]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Organization variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		variableData		body	Variable	true	"the new variable"
func PostOrgVariable(c *gin.Context) {
	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	in := new(model.Variable)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing org %q variable. %s", orgID, err)
		return
	}
	variable := &model.Variable{
		OrgID: orgID,
		Name:  in.Name,
		Value: in.Value,
	}
	if err := variable.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting org %q variable. %s", orgID, err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	if err := variableService.OrgVariableCreate(orgID, variable); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting org %q variable %q. %s", orgID, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// PatchOrgVariable
//
//	@Summary	Update an organization variable by name
//	@Router		/orgs/{org_id}/variables/{variable} [patch]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Organization variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		variable			path	string	true	"the variable's name"
//	@Param		variableData		body	Variable	true	"the update variable data"
func PatchOrgVariable(c *gin.Context) {
	name := c.Param("variable")
	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	in := new(model.Variable)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing variable. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	variable, err := variableService.OrgVariableFind(orgID, name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if in.Value != "" {
		variable.Value = in.Value
	}

	if err := variable.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error updating org %q variable. %s", orgID, err)
		return
	}

	if err := variableService.OrgVariableUpdate(orgID, variable); err != nil {
		c.String(http.StatusInternalServerError, "Error updating org %q variable %q. %s", orgID, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// DeleteOrgVariable
//
//	@Summary	Delete an organization variable by name
//	@Router		/orgs/{org_id}/variables/{variable} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Organization variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		variable			path	string	true	"the variable's name"
func DeleteOrgVariable(c *gin.Context) {
	name := c.Param("variable")
	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	if err := variableService.OrgVariableDelete(orgID, name); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
