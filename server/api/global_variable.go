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

// GetGlobalVariableList
//
//	@Summary	List global variables
//	@Router		/variables [get]
//	@Produce	json
//	@Success	200	{array}	Variable
//	@Tags		Variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetGlobalVariableList(c *gin.Context) {
	variableService := server.Config.Services.Manager.VariableService()
	list, err := variableService.GlobalVariableList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting global variable list. %s", err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetGlobalVariable
//
//	@Summary	Get a global variable by name
//	@Router		/variables/{variable} [get]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		variable			path	string	true	"the variable's name"
func GetGlobalVariable(c *gin.Context) {
	name := c.Param("variable")
	variableService := server.Config.Services.Manager.VariableService()
	variable, err := variableService.GlobalVariableFind(name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// PostGlobalVariable
//
//	@Summary	Create a global variable
//	@Router		/variables [post]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		variable			body	Variable	true	"the variable object data"
func PostGlobalVariable(c *gin.Context) {
	in := new(model.Variable)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing global variable. %s", err)
		return
	}
	variable := &model.Variable{
		Name:  in.Name,
		Value: in.Value,
	}
	if err := variable.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error inserting global variable. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	if err := variableService.GlobalVariableCreate(variable); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting global variable %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// PatchGlobalVariable
//
//	@Summary	Update a global variable by name
//	@Router		/variables/{variable} [patch]
//	@Produce	json
//	@Success	200	{object}	Variable
//	@Tags		Variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		variable			path	string	true	"the variable's name"
//	@Param		variableData		body	Variable	true	"the variable's data"
func PatchGlobalVariable(c *gin.Context) {
	name := c.Param("variable")

	in := new(model.Variable)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing variable. %s", err)
		return
	}

	variableService := server.Config.Services.Manager.VariableService()
	variable, err := variableService.GlobalVariableFind(name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if in.Value != "" {
		variable.Value = in.Value
	}

	if err := variable.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error updating global variable. %s", err)
		return
	}

	if err := variableService.GlobalVariableUpdate(variable); err != nil {
		c.String(http.StatusInternalServerError, "Error updating global variable %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, variable)
}

// DeleteGlobalVariable
//
//	@Summary	Delete a global variable by name
//	@Router		/variables/{variable} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Variables
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		variable			path	string	true	"the variable's name"
func DeleteGlobalVariable(c *gin.Context) {
	name := c.Param("variable")
	variableService := server.Config.Services.Manager.VariableService()
	if err := variableService.GlobalVariableDelete(name); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
