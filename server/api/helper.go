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
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func handlePipelineErr(c *gin.Context, err error) {
	if pipeline.IsErrNotFound(err) {
		c.String(http.StatusNotFound, "%v", err)
	} else if pipeline.IsErrBadRequest(err) {
		c.String(http.StatusBadRequest, "%v", err)
	} else if pipeline.IsErrFiltered(err) {
		c.String(http.StatusNoContent, "%v", err)
	} else {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// if the forge has a refresh token, the current access token may be stale.
// Therefore, we should refresh prior to dispatching the job.
func refreshUserToken(c *gin.Context, user *model.User) {
	_forge := session.Forge(c)
	_store := store.FromContext(c)
	if refresher, ok := _forge.(forge.Refresher); ok {
		ok, err := refresher.Refresh(c, user)
		if err != nil {
			log.Error().Err(err).Msgf("refresh oauth token of user '%s' failed", user.Login)
		} else if ok {
			if err := _store.UpdateUser(user); err != nil {
				log.Error().Err(err).Msg("fail to save user to store after refresh oauth token")
			}
		}
	}
}
