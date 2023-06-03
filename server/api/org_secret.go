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
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

// GetOrgSecret
//
//	@Summary	Get the named organization secret
//	@Router		/orgs/{owner}/secrets/{secret} [get]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
//	@Param		secret			path	string	true	"the secret's name"
func GetOrgSecret(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("secret")
	)
	secret, err := server.Config.Services.Secrets.OrgSecretFind(owner, name)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// GetOrgSecretList
//
//	@Summary	Get the organization secret list
//	@Router		/orgs/{owner}/secrets [get]
//	@Produce	json
//	@Success	200	{array}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgSecretList(c *gin.Context) {
	owner := c.Param("owner")
	list, err := server.Config.Services.Secrets.OrgSecretList(owner, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting secret list for %q. %s", owner, err)
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
//	@Summary	Persist/create an organization secret
//	@Router		/orgs/{owner}/secrets [post]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
//	@Param		secretData		body	Secret	true	"the new secret"
func PostOrgSecret(c *gin.Context) {
	owner := c.Param("owner")

	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing org %q secret. %s", owner, err)
		return
	}
	secret := &model.Secret{
		Owner:       owner,
		Name:        in.Name,
		Value:       in.Value,
		Events:      in.Events,
		Images:      in.Images,
		PluginsOnly: in.PluginsOnly,
	}
	if err := secret.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting org %q secret. %s", owner, err)
		return
	}
	if err := server.Config.Services.Secrets.OrgSecretCreate(owner, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting org %q secret %q. %s", owner, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PatchOrgSecret
//
//	@Summary	Update an organization secret
//	@Router		/orgs/{owner}/secrets/{secret} [patch]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
//	@Param		secret			path	string	true	"the secret's name"
//	@Param		secretData		body	Secret	true	"the update secret data"
func PatchOrgSecret(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("secret")
	)

	in := new(model.Secret)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing secret. %s", err)
		return
	}

	secret, err := server.Config.Services.Secrets.OrgSecretFind(owner, name)
	if err != nil {
		handleDbGetError(c, err)
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
	secret.PluginsOnly = in.PluginsOnly

	if err := secret.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error updating org %q secret. %s", owner, err)
		return
	}
	if err := server.Config.Services.Secrets.OrgSecretUpdate(owner, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error updating org %q secret %q. %s", owner, in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// DeleteOrgSecret
//
//	@Summary	Delete the named secret from an organization
//	@Router		/orgs/{owner}/secrets/{secret} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Organization secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the owner's name"
//	@Param		secret			path	string	true	"the secret's name"
func DeleteOrgSecret(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("secret")
	)
	if err := server.Config.Services.Secrets.OrgSecretDelete(owner, name); err != nil {
		handleDbGetError(c, err)
		return
	}
	c.String(http.StatusNoContent, "")
}
