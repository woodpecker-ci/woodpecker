// Copyright 2024 Woodpecker Authors
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
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/setup"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// validateForge ensures the forge configuration can actually be constructed,
// so invalid configurations (e.g. an unparsable GitHub App private key or a
// private key without an app id) are rejected at save time instead of
// breaking every operation of the forge afterwards. Addon forges are not
// validated as constructing them starts an external process.
func validateForge(forge *model.Forge) error {
	if forge.Type == model.ForgeTypeAddon {
		return nil
	}

	// additional options are a schemaless JSON map, ensure the string options
	// have the right type before the setup helpers silently coerce them
	if forge.Type == model.ForgeTypeGithub {
		for _, key := range []string{model.ForgeGithubOptionAppID, model.ForgeGithubOptionAppCloneTokenScope} {
			if value, ok := forge.AdditionalOptions[key]; ok {
				if _, isString := value.(string); !isString {
					return fmt.Errorf("additional option %q must be a string", key)
				}
			}
		}
	}

	_, err := setup.Forge(forge)
	return err
}

// GetForges
//
//	@Summary	List forges
//	@Router		/forges [get]
//	@Produce	json
//	@Success	200	{array}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string	false	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetForges(c *gin.Context) {
	forges, err := store.FromContext(c).ForgeList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting forge list. %s", err)
		return
	}

	user := session.User(c)
	if user != nil && user.Admin {
		for _, forge := range forges {
			forge.RedactSecrets()
		}
		c.JSON(http.StatusOK, forges)
		return
	}

	// copy forges data without sensitive information
	for i, forge := range forges {
		forges[i] = forge.PublicCopy()
	}

	c.JSON(http.StatusOK, forges)
}

// GetForge
//
//	@Summary	Get a forge
//	@Router		/forges/{forge_id} [get]
//	@Produce	json
//	@Success	200	{object}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string	false	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forge_id		path	int		true	"the forge's id"
func GetForge(c *gin.Context) {
	forgeID, err := strconv.ParseInt(c.Param("forge_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := store.FromContext(c).ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	user := session.User(c)
	if user != nil && user.Admin {
		forge.RedactSecrets()
		c.JSON(http.StatusOK, forge)
	} else {
		c.JSON(http.StatusOK, forge.PublicCopy())
	}
}

// PatchForge
//
//	@Summary	Update a forge
//	@Router		/forges/{forge_id} [patch]
//	@Produce	json
//	@Success	200	{object}	Forge
//	@Tags		Forges
//	@Param		Authorization	header	string						true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forge_id		path	int							true	"the forge's id"
//	@Param		forgeData		body	ForgeWithOAuthClientSecret	true	"the forge's data"
func PatchForge(c *gin.Context) {
	_store := store.FromContext(c)

	in := &model.ForgeWithOAuthClientSecret{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	forgeID, err := strconv.ParseInt(c.Param("forge_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := _store.ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	oldAdditionalOptions := forge.AdditionalOptions
	forge.URL = in.URL
	forge.Type = in.Type
	forge.OAuthClientID = in.OAuthClientID
	forge.OAuthHost = in.OAuthHost
	forge.SkipVerify = in.SkipVerify
	forge.AdditionalOptions = in.AdditionalOptions
	if in.OAuthClientSecret != "" {
		forge.OAuthClientSecret = in.OAuthClientSecret
	}
	restoreSecretOptions(forge, oldAdditionalOptions)

	if err := validateForge(forge); err != nil {
		c.String(http.StatusBadRequest, "invalid forge configuration: %s", err)
		return
	}

	err = _store.ForgeUpdate(forge)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	forge.RedactSecrets()
	c.JSON(http.StatusOK, forge)
}

// restoreSecretOptions copies write-only secret options from the stored
// forge configuration when an update request omits them or sends them empty,
// mirroring the OAuthClientSecret semantics.
func restoreSecretOptions(forge *model.Forge, oldOptions map[string]any) {
	submittedAppKey, _ := forge.AdditionalOptions[model.ForgeGithubOptionAppPrivateKey].(string)

	for _, key := range model.SecretForgeOptions(forge.Type) {
		// never persist the redaction marker clients echo back
		delete(forge.AdditionalOptions, model.SecretForgeOptionSetMarker(key))
		oldValue, _ := oldOptions[key].(string)
		newValue, _ := forge.AdditionalOptions[key].(string)
		if oldValue != "" && newValue == "" {
			if forge.AdditionalOptions == nil {
				forge.AdditionalOptions = make(map[string]any)
			}
			forge.AdditionalOptions[key] = oldValue
		}
	}

	// a restored github app private key is only kept while an app id is
	// configured, so clearing the app id disables app authentication again -
	// a key explicitly submitted with this request is left alone, the
	// validation rejects such inconsistent configurations instead
	if forge.Type == model.ForgeTypeGithub && submittedAppKey == "" {
		if appID, _ := forge.AdditionalOptions[model.ForgeGithubOptionAppID].(string); appID == "" {
			delete(forge.AdditionalOptions, model.ForgeGithubOptionAppPrivateKey)
		}
	}
}

// dropSecretOptionMarkers removes the redaction markers clients echo back,
// so they are never persisted.
func dropSecretOptionMarkers(forge *model.Forge) {
	for _, key := range model.SecretForgeOptions(forge.Type) {
		delete(forge.AdditionalOptions, model.SecretForgeOptionSetMarker(key))
	}
}

// PostForge
//
//	@Summary		Create a new forge
//	@Description	Creates a new forge with a random token
//	@Router			/forges [post]
//	@Produce		json
//	@Success		200	{object}	Forge
//	@Tags			Forges
//	@Param			Authorization	header	string						true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			forge			body	ForgeWithOAuthClientSecret	true	"the forge's data (only 'name' and 'no_schedule' are read)"
func PostForge(c *gin.Context) {
	in := &model.ForgeWithOAuthClientSecret{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	forge := &model.Forge{
		URL:               in.URL,
		Type:              in.Type,
		OAuthClientID:     in.OAuthClientID,
		OAuthClientSecret: in.OAuthClientSecret,
		OAuthHost:         in.OAuthHost,
		SkipVerify:        in.SkipVerify,
		AdditionalOptions: in.AdditionalOptions,
	}
	dropSecretOptionMarkers(forge)

	if err := validateForge(forge); err != nil {
		c.String(http.StatusBadRequest, "invalid forge configuration: %s", err)
		return
	}

	if err = store.FromContext(c).ForgeCreate(forge); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	forge.RedactSecrets()
	c.JSON(http.StatusOK, forge)
}

// ForgeAppHealth is the result of checking a forge's GitHub App configuration.
type ForgeAppHealth struct {
	Healthy       bool   `json:"healthy"`
	AppName       string `json:"app_name,omitempty"`
	Installations int    `json:"installations"`
	Error         string `json:"error,omitempty"`
} //	@name	ForgeAppHealth

// githubAppChecker is implemented by forges that can verify their GitHub App
// configuration.
type githubAppChecker interface {
	AppHealth(ctx context.Context) (name string, installations int, err error)
}

// appCheckTimeout bounds the outbound forge API calls of the app check.
const appCheckTimeout = 10 * time.Second

// GetForgeAppHealth
//
//	@Summary		Check the GitHub App configuration of a forge
//	@Description	Authenticates as the configured GitHub App and reports whether the credentials work.
//	@Router			/forges/{forge_id}/app-health [get]
//	@Produce		json
//	@Success		200	{object}	ForgeAppHealth
//	@Tags			Forges
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			forge_id		path	int		true	"the forge's id"
func GetForgeAppHealth(c *gin.Context) {
	forgeID, err := strconv.ParseInt(c.Param("forge_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := store.FromContext(c).ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// only github forges support app checks, and constructing other forge
	// types may have side effects (addon forges start an external process)
	if forge.Type != model.ForgeTypeGithub {
		c.String(http.StatusBadRequest, "forge does not support app checks")
		return
	}

	// construct a fresh instance from the stored configuration, bypassing
	// the service manager cache so recently saved credentials are the ones
	// being tested
	forgeInstance, err := setup.Forge(forge)
	if err != nil {
		c.JSON(http.StatusOK, ForgeAppHealth{Healthy: false, Error: err.Error()})
		return
	}

	checker, ok := forgeInstance.(githubAppChecker)
	if !ok {
		c.String(http.StatusBadRequest, "forge does not support app checks")
		return
	}

	ctx, cancel := context.WithTimeout(c, appCheckTimeout)
	defer cancel()

	name, installations, err := checker.AppHealth(ctx)
	if err != nil {
		c.JSON(http.StatusOK, ForgeAppHealth{Healthy: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ForgeAppHealth{Healthy: true, AppName: name, Installations: installations})
}

// DeleteForge
//
//	@Summary	Delete a forge
//	@Router		/forges/{forge_id} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Forges
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		forge_id		path	int		true	"the forge's id"
func DeleteForge(c *gin.Context) {
	_store := store.FromContext(c)

	forgeID, err := strconv.ParseInt(c.Param("forge_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	forge, err := _store.ForgeGet(forgeID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	if err = _store.ForgeDelete(forge); err != nil {
		c.String(http.StatusInternalServerError, "Error deleting user. %s", err)
		return
	}
	c.Status(http.StatusNoContent)
}
