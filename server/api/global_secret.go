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

// GetGlobalSecretList
//
//	@Summary	Get the global secret list
//	@Router		/secrets [get]
//	@Produce	json
//	@Success	200	{array}	Secret
//	@Tags		Secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetGlobalSecretList(c *gin.Context) {
	list, err := server.Config.Services.Secrets.GlobalSecretList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting global secret list. %s", err)
		return
	}
	// copy the secret detail to remove the sensitive
	// password and token fields.
	for i, secret := range list {
		list[i] = secret.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// GetGlobalSecret
//
//	@Summary	Get a global secret by name
//	@Router		/secrets/{secret} [get]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		secret			path	string	true	"the secret's name"
func GetGlobalSecret(c *gin.Context) {
	name := c.Param("secret")
	secret, err := server.Config.Services.Secrets.GlobalSecretFind(name)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PostGlobalSecret
//
//	@Summary	Persist/create a global secret
//	@Router		/secrets [post]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Secrets
//	@Param		Authorization	header	string			true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		secret			body	Secret	true	"the secret object data"
func PostGlobalSecret(c *gin.Context) {
	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing global secret. %s", err)
		return
	}
	secret := &model.Secret{
		Name:        in.Name,
		Value:       in.Value,
		Events:      in.Events,
		Images:      in.Images,
		PluginsOnly: in.PluginsOnly,
	}
	if err := secret.Validate(); err != nil {
		c.String(http.StatusBadRequest, "Error inserting global secret. %s", err)
		return
	}
	if err := server.Config.Services.Secrets.GlobalSecretCreate(secret); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting global secret %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PatchGlobalSecret
//
//	@Summary	Update a global secret by name
//	@Router		/secrets/{secret} [patch]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Secrets
//	@Param		Authorization	header	string			true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		secret			path	string			true	"the secret's name"
//	@Param		secretData		body	Secret	true	"the secret's data"
func PatchGlobalSecret(c *gin.Context) {
	name := c.Param("secret")

	in := new(model.Secret)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing secret. %s", err)
		return
	}

	secret, err := server.Config.Services.Secrets.GlobalSecretFind(name)
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
		c.String(http.StatusBadRequest, "Error updating global secret. %s", err)
		return
	}
	if err := server.Config.Services.Secrets.GlobalSecretUpdate(secret); err != nil {
		c.String(http.StatusInternalServerError, "Error updating global secret %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// DeleteGlobalSecret
//
//	@Summary	Delete a global secret by name
//	@Router		/secrets/{secret} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		secret			path	string	true	"the secret's name"
func DeleteGlobalSecret(c *gin.Context) {
	name := c.Param("secret")
	if err := server.Config.Services.Secrets.GlobalSecretDelete(name); err != nil {
		handleDbGetError(c, err)
		return
	}
	c.String(http.StatusNoContent, "")
}
