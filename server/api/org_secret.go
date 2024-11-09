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

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
)

// GetOrgSecret
//
//	@Summary	Get a organization secret by name
//	@Router		/orgs/{org_id}/secrets/{secret} [get]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		secret			path	string	true	"the secret's name"
func GetOrgSecret(c *gin.Context) {
	org := session.Org(c)
	name := c.Param("secret")

	secretService := server.Config.Services.Manager.SecretService()
	secret, err := secretService.OrgSecretFind(org.ID, name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// GetOrgSecretList
//
//	@Summary	List organization secrets
//	@Router		/orgs/{org_id}/secrets [get]
//	@Produce	json
//	@Success	200	{array}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgSecretList(c *gin.Context) {
	org := session.Org(c)

	secretService := server.Config.Services.Manager.SecretService()
	list, err := secretService.OrgSecretList(org.ID, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting secret list for %q. %s", org.ID, err)
		return
	}
	// copy the secret detail to remove the sensitive
	// password and token fields.
	for i, secret := range list {
		list[i] = secret.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// PostOrgSecret
//
//	@Summary	Create an organization secret
//	@Router		/orgs/{org_id}/secrets [post]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		secretData		body	Secret	true	"the new secret"
func PostOrgSecret(c *gin.Context) {
	org := session.Org(c)

	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing org %q secret. %s", org.ID, err)
		return
	}
	secret := &model.Secret{
		OrgID:  org.ID,
		Name:   in.Name,
		Value:  in.Value,
		Events: in.Events,
		Images: in.Images,
	}
	if err := secret.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting org %q secret. %s", org.ID, err)
		return
	}

	secretService := server.Config.Services.Manager.SecretService()
	if err := secretService.OrgSecretCreate(org.ID, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting org %q secret %q. %s", org.ID, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PatchOrgSecret
//
//	@Summary	Update an organization secret by name
//	@Router		/orgs/{org_id}/secrets/{secret} [patch]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		secret			path	string	true	"the secret's name"
//	@Param		secretData		body	Secret	true	"the update secret data"
func PatchOrgSecret(c *gin.Context) {
	org := session.Org(c)
	name := c.Param("secret")

	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing secret. %s", err)
		return
	}

	secretService := server.Config.Services.Manager.SecretService()
	secret, err := secretService.OrgSecretFind(org.ID, name)
	if err != nil {
		handleDBError(c, err)
		return
	}
	if in.Value != "" {
		secret.Value = in.Value
	}
	if in.Events != nil {
		secret.Events = in.Events
	}
	if in.Images != nil {
		secret.Images = in.Images
	}

	if err := secret.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error updating org %q secret. %s", org.ID, err)
		return
	}

	if err := secretService.OrgSecretUpdate(org.ID, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error updating org %q secret %q. %s", org.ID, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// DeleteOrgSecret
//
//	@Summary	Delete an organization secret by name
//	@Router		/orgs/{org_id}/secrets/{secret} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	string	true	"the org's id"
//	@Param		secret			path	string	true	"the secret's name"
func DeleteOrgSecret(c *gin.Context) {
	org := session.Org(c)
	name := c.Param("secret")

	secretService := server.Config.Services.Manager.SecretService()
	if err := secretService.OrgSecretDelete(org.ID, name); err != nil {
		handleDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
