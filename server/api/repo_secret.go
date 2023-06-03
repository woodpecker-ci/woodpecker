// Copyright 2018 Drone.IO Inc.
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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
)

// GetSecret
//
//	@Summary	Get a named secret
//	@Router		/repos/{owner}/{name}/secrets/{secretName} [get]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Repository secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the repository owner's name"
//	@Param		name			path	string	true	"the repository name"
//	@Param		secretName		path	string	true	"the secret name"
func GetSecret(c *gin.Context) {
	var (
		repo = session.Repo(c)
		name = c.Param("secret")
	)
	secret, err := server.Config.Services.Secrets.SecretFind(repo, name)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PostSecret
//
//	@Summary	Persist/create a secret
//	@Router		/repos/{owner}/{name}/secrets [post]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Repository secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the repository owner's name"
//	@Param		name			path	string	true	"the repository name"
//	@Param		secret			body	Secret	true	"the new secret"
func PostSecret(c *gin.Context) {
	repo := session.Repo(c)

	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing secret. %s", err)
		return
	}
	secret := &model.Secret{
		RepoID:      repo.ID,
		Name:        strings.ToLower(in.Name),
		Value:       in.Value,
		Events:      in.Events,
		Images:      in.Images,
		PluginsOnly: in.PluginsOnly,
	}
	if err := secret.Validate(); err != nil {
		c.String(http.StatusUnprocessableEntity, "Error inserting secret. %s", err)
		return
	}
	if err := server.Config.Services.Secrets.SecretCreate(repo, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error inserting secret %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// PatchSecret
//
//	@Summary	Update a named secret
//	@Router		/repos/{owner}/{name}/secrets/{secretName} [patch]
//	@Produce	json
//	@Success	200	{object}	Secret
//	@Tags		Repository secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the repository owner's name"
//	@Param		name			path	string	true	"the repository name"
//	@Param		secretName		path	string	true	"the secret name"
//	@Param		secret			body	Secret	true	"the secret itself"
func PatchSecret(c *gin.Context) {
	var (
		repo = session.Repo(c)
		name = c.Param("secret")
	)

	in := new(model.Secret)
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing secret. %s", err)
		return
	}

	secret, err := server.Config.Services.Secrets.SecretFind(repo, name)
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
		c.String(http.StatusUnprocessableEntity, "Error updating secret. %s", err)
		return
	}
	if err := server.Config.Services.Secrets.SecretUpdate(repo, secret); err != nil {
		c.String(http.StatusInternalServerError, "Error updating secret %q. %s", in.Name, err)
		return
	}
	c.JSON(http.StatusOK, secret.Copy())
}

// GetSecretList
//
//	@Summary	Get the secret list
//	@Router		/repos/{owner}/{name}/secrets [get]
//	@Produce	json
//	@Success	200	{array}	Secret
//	@Tags		Repository secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the repository owner's name"
//	@Param		name			path	string	true	"the repository name"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetSecretList(c *gin.Context) {
	repo := session.Repo(c)
	list, err := server.Config.Services.Secrets.SecretList(repo, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting secret list. %s", err)
		return
	}
	// copy the secret detail to remove the sensitive
	// password and token fields.
	for i, secret := range list {
		list[i] = secret.Copy()
	}
	c.JSON(http.StatusOK, list)
}

// DeleteSecret
//
//	@Summary	Delete a named secret
//	@Router		/repos/{owner}/{name}/secrets/{secretName} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Repository secrets
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		owner			path	string	true	"the repository owner's name"
//	@Param		name			path	string	true	"the repository name"
//	@Param		secretName		path	string	true	"the secret name"
func DeleteSecret(c *gin.Context) {
	var (
		repo = session.Repo(c)
		name = c.Param("secret")
	)
	if err := server.Config.Services.Secrets.SecretDelete(repo, name); err != nil {
		handleDbGetError(c, err)
		return
	}
	c.String(http.StatusNoContent, "")
}
