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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"

	"github.com/gin-gonic/gin"
)

// GetOrgPermissions
//
//	@Summary	Get the permissions of the current user in the given organization
//	@Router		/orgs/{owner}/permissions [get]
//	@Produce	json
//	@Success	200	{array}	OrgPerm
//	@Tags		Organization permissions
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
func GetOrgPermissions(c *gin.Context) {
	var (
		err   error
		user  = session.User(c)
		owner = c.Param("owner")
	)

	if user == nil {
		c.JSON(http.StatusOK, &model.OrgPerm{})
		return
	}

	perm, err := server.Config.Services.Membership.Get(c, user, owner)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting membership for %q. %s", owner, err)
		return
	}

	c.JSON(http.StatusOK, perm)
}
