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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

// GetOrgs
//
//	@Summary		List organizations
//	@Description	Returns all registered orgs in the system. Requires admin rights.
//	@Router			/orgs [get]
//	@Produce		json
//	@Success		200	{array}	Org
//	@Tags			Orgs
//	@Param			Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param			page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param			perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgs(c *gin.Context) {
	orgs, err := store.FromContext(c).OrgList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting user list. %s", err)
		return
	}
	c.JSON(http.StatusOK, orgs)
}

// GetOrg
//
//	@Summary	Get an organization
//	@Router		/orgs/{org_id} [get]
//	@Produce	json
//	@Success	200	{array}	Org
//	@Tags		Organization
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the organization's id"
func GetOrg(c *gin.Context) {
	org := session.Org(c)
	c.JSON(http.StatusOK, org)
}

// GetOrgPermissions
//
//	@Summary	Get the permissions of the currently authenticated user for the given organization
//	@Router		/orgs/{org_id}/permissions [get]
//	@Produce	json
//	@Success	200	{array}	OrgPerm
//	@Tags		Organization permissions
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the organization's id"
func GetOrgPermissions(c *gin.Context) {
	user := session.User(c)
	org := session.Org(c)

	_forge, err := server.Config.Services.Manager.ForgeFromUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from user")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if user == nil {
		c.JSON(http.StatusOK, &model.OrgPerm{})
		return
	}

	if (org.IsUser && org.Name == user.Login) || (user.Admin && !org.IsUser) {
		c.JSON(http.StatusOK, &model.OrgPerm{
			Member: true,
			Admin:  true,
		})
		return
	} else if org.IsUser {
		c.JSON(http.StatusOK, &model.OrgPerm{})
		return
	}

	perm, err := server.Config.Services.Membership.Get(c, _forge, user, org.Name)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting membership for %d. %s", org.ID, err)
		return
	}

	c.JSON(http.StatusOK, perm)
}

// LookupOrg
//
//	@Summary	Lookup an organization by full name
//	@Router		/org/lookup/{org_full_name} [get]
//	@Produce	json
//	@Success	200	{object}	Org
//	@Tags		Organizations
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_full_name	path	string	true	"the organizations full name / slug"
func LookupOrg(c *gin.Context) {
	_store := store.FromContext(c)
	user := session.User(c)
	_forge, err := server.Config.Services.Manager.ForgeFromUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Cannot get forge from user")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	orgFullName := strings.TrimLeft(c.Param("org_full_name"), "/")

	org, err := _store.OrgFindByName(orgFullName)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// don't leak private org infos
	if org.Private {
		user := session.User(c)
		if user == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if !user.Admin && org.Name != user.Login {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else if !user.Admin {
			perm, err := server.Config.Services.Membership.Get(c, _forge, user, org.Name)
			if err != nil {
				log.Error().Err(err).Msg("failed to check membership")
				c.Status(http.StatusInternalServerError)
				return
			}

			if perm == nil || !perm.Member {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrg
//
//	@Summary		Delete an organization
//	@Description	Deletes the given org. Requires admin rights.
//	@Router			/orgs/{id} [delete]
//	@Produce		plain
//	@Success		204
//	@Tags			Orgs
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			id				path	string	true	"the org's id"
func DeleteOrg(c *gin.Context) {
	_store := store.FromContext(c)
	org := session.Org(c)

	if err := _store.OrgDelete(org.ID); err != nil {
		handleDBError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
