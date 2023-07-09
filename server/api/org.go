// Copyright 2022 Woodpecker Authors
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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"

	"github.com/gin-gonic/gin"
)

// GetOrgPermissions
//
//	@Summary	Get the permissions of the current user in the given organization
//	@Router		/orgs/{org_id}/permissions [get]
//	@Produce	json
//	@Success	200	{array}	OrgPerm
//	@Tags		Organization permissions
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the organziation's id"
func GetOrgPermissions(c *gin.Context) {
	user := session.User(c)
	_store := store.FromContext(c)

	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	if user == nil {
		c.JSON(http.StatusOK, &model.OrgPerm{})
		return
	}

	org, err := _store.OrgFind(orgID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting org %d. %s", orgID, err)
		return
	}

	perm, err := server.Config.Services.Membership.Get(c, user, org.Name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting membership for %d. %s", orgID, err)
		return
	}

	c.JSON(http.StatusOK, perm)
}
