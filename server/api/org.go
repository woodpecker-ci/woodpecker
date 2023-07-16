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
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"

	"github.com/gin-gonic/gin"
)

// GetOrg
//
//	@Summary	Get organization by id
//	@Router		/orgs/{org_id} [get]
//	@Produce	json
//	@Success	200	{array}	Org
//	@Tags		Organization
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the organziation's id"
func GetOrg(c *gin.Context) {
	_store := store.FromContext(c)

	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing org id. %s", err)
		return
	}

	org, err := _store.OrgGet(orgID)
	if err != nil {
		if errors.Is(err, types.RecordNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, org)
}

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

	org, err := _store.OrgGet(orgID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting org %d. %s", orgID, err)
		return
	}

	if (org.IsUser && org.Name == user.Login) || user.Admin {
		c.JSON(http.StatusOK, &model.OrgPerm{
			Member: true,
			Admin:  true,
		})
		return
	}

	perm, err := server.Config.Services.Membership.Get(c, user, org.Name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting membership for %d. %s", orgID, err)
		return
	}

	c.JSON(http.StatusOK, perm)
}

// LookupOrg
//
//	@Summary	Lookup organization by full-name
//	@Router		/org/lookup/{org_full_name} [get]
//	@Produce	json
//	@Success	200	{object}	Org
//	@Tags		Organizations
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_full_name	path	string	true	"the organizations full-name / slug"
func LookupOrg(c *gin.Context) {
	_store := store.FromContext(c)

	orgFullName := strings.TrimLeft(c.Param("org_full_name"), "/")

	org, err := _store.OrgFindByName(orgFullName)
	if err != nil {
		if errors.Is(err, types.RecordNotExist) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// don't leak private org infos
	if org.Private {
		user := session.User(c)
		if user == nil {
			c.String(http.StatusUnauthorized, "User not authorized")
			return
		}

		if !user.Admin && org.IsUser {
			if org.Name != user.Login {
				c.String(http.StatusForbidden, "User not authorized")
				return
			}
		} else if !user.Admin {
			perm, err := server.Config.Services.Membership.Get(c, user, org.Name)
			if err != nil {
				log.Error().Msgf("Failed to check membership: %v", err)
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}

			if perm == nil || !perm.Member {
				c.String(http.StatusForbidden, "User not authorized")
				return
			}
		}
	}

	c.JSON(http.StatusOK, org)
}
