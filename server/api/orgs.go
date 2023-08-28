// Copyright 2023 Woodpecker Authors
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

	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// GetOrgs
//
//	@Summary		Get all orgs
//	@Description	Returns all registered orgs in the system. Requires admin rights.
//	@Router			/orgs [get]
//	@Produce		json
//	@Success		200	{array}	Org
//	@Tags			Orgs
//	@Param			Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param			page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param			perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgs(c *gin.Context) {
	orgs, err := store.FromContext(c).GetOrgList(session.Pagination(c))
	if err != nil {
		c.String(500, "Error getting user list. %s", err)
		return
	}
	c.JSON(200, orgs)
}

// DeleteOrg
//
//	@Summary		Delete an org
//	@Description	Deletes the given org. Requires admin rights.
//	@Router			/orgs/{id} [delete]
//	@Produce		json
//	@Success		204	{object}	Org
//	@Tags			Orgs
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			id			path	string	true	"the org's id"
func DeleteOrg(c *gin.Context) {
	_store := store.FromContext(c)

	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	err = _store.OrgDelete(orgID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting org %d. %s", orgID, err)
	}

	c.String(http.StatusNoContent, "")
}
