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

	"github.com/gin-gonic/gin"
)

// GetOrgSecret gets the named organization secret from the database
// and writes to the response in json format.
func GetOrgSecret(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("secret")
	)
	secret, err := server.Config.Services.Secrets.OrgSecretFind(owner, name)
	if err != nil {
		c.String(404, "Error getting org %q secret %q. %s", owner, name, err)
		return
	}
	c.JSON(200, secret.Copy())
}

// GetOrgSecretList gest the organization secret list from
// the database and writes to the response in json format.
func GetOrgSecretList(c *gin.Context) {
	owner := c.Param("owner")
	list, err := server.Config.Services.Secrets.OrgSecretList(owner)
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

// PostOrgSecret persists an organization secret to the database.
func PostOrgSecret(c *gin.Context) {
	owner := c.Param("owner")

	in := new(model.Secret)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, "Error parsing org %q secret. %s", owner, err)
		return
	}
	secret := &model.Secret{
		Owner:  owner,
		Name:   in.Name,
		Value:  in.Value,
		Events: in.Events,
		Images: in.Images,
	}
	if err := secret.Validate(); err != nil {
		c.String(400, "Error inserting org %q secret. %s", owner, err)
		return
	}
	if err := server.Config.Services.Secrets.OrgSecretCreate(owner, secret); err != nil {
		c.String(500, "Error inserting org %q secret %q. %s", owner, in.Name, err)
		return
	}
	c.JSON(200, secret.Copy())
}

// PatchOrgSecret updates an organization secret in the database.
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
		c.String(404, "Error getting org %q secret %q. %s", owner, name, err)
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
		c.String(400, "Error updating org %q secret. %s", owner, err)
		return
	}
	if err := server.Config.Services.Secrets.OrgSecretUpdate(owner, secret); err != nil {
		c.String(500, "Error updating org %q secret %q. %s", owner, in.Name, err)
		return
	}
	c.JSON(200, secret.Copy())
}

// DeleteOrgSecret deletes the named organization secret from the database.
func DeleteOrgSecret(c *gin.Context) {
	var (
		owner = c.Param("owner")
		name  = c.Param("secret")
	)
	if err := server.Config.Services.Secrets.OrgSecretDelete(owner, name); err != nil {
		c.String(500, "Error deleting org %q secret %q. %s", owner, name, err)
		return
	}
	c.String(204, "")
}
