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
	"encoding/base32"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
	"github.com/woodpecker-ci/woodpecker/shared/httputil"
	"github.com/woodpecker-ci/woodpecker/shared/token"
)

func HandleLogin(c *gin.Context) {
	var (
		w = c.Writer
		r = c.Request
	)
	if err := r.FormValue("error"); err != "" {
		http.Redirect(w, r, "/login/error?code="+err, 303)
	} else {
		http.Redirect(w, r, "/authorize", 303)
	}
}

func HandleAuth(c *gin.Context) {
	_store := store.FromContext(c)
	forge := session.Forge(c)

	// when dealing with redirects we may need to adjust the content type. I
	// cannot, however, remember why, so need to revisit this line.
	c.Writer.Header().Del("Content-Type")

	tmpuser, err := forge.Login(c, c.Writer, c.Request)
	if err != nil {
		log.Error().Msgf("cannot authenticate user. %s", err)
		c.Redirect(http.StatusSeeOther, "/login?error=oauth_error")
		return
	}
	// this will happen when the user is redirected by the forge as
	// part of the authorization workflow.
	if tmpuser == nil {
		return
	}
	config := ToConfig(c)

	// get the user from the database
	u, err := _store.GetUserRemoteID(tmpuser.ForgeRemoteID, tmpuser.Login)
	if err != nil {
		if !errors.Is(err, types.RecordNotExist) {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// if self-registration is disabled we should return a not authorized error
		if !config.Open && !config.IsAdmin(tmpuser) {
			log.Error().Msgf("cannot register %s. registration closed", tmpuser.Login)
			c.Redirect(http.StatusSeeOther, "/login?error=access_denied")
			return
		}

		// if self-registration is enabled for whitelisted organizations we need to
		// check the user's organization membership.
		if len(config.Orgs) != 0 {
			teams, terr := forge.Teams(c, tmpuser)
			if terr != nil || !config.IsMember(teams) {
				log.Error().Msgf("cannot verify team membership for %s.", u.Login)
				c.Redirect(303, "/login?error=access_denied")
				return
			}
		}

		// create the user account
		u = &model.User{
			Login:         tmpuser.Login,
			ForgeRemoteID: tmpuser.ForgeRemoteID,
			Token:         tmpuser.Token,
			Secret:        tmpuser.Secret,
			Email:         tmpuser.Email,
			Avatar:        tmpuser.Avatar,
			Hash: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
		}

		// insert the user into the database
		if err := _store.CreateUser(u); err != nil {
			log.Error().Msgf("cannot insert %s. %s", u.Login, err)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}
	}

	// update the user meta data and authorization data.
	u.Token = tmpuser.Token
	u.Secret = tmpuser.Secret
	u.Email = tmpuser.Email
	u.Avatar = tmpuser.Avatar
	u.ForgeRemoteID = tmpuser.ForgeRemoteID
	u.Login = tmpuser.Login
	u.Admin = u.Admin || config.IsAdmin(tmpuser)

	// if self-registration is enabled for whitelisted organizations we need to
	// check the user's organization membership.
	if len(config.Orgs) != 0 {
		teams, terr := forge.Teams(c, u)
		if terr != nil || !config.IsMember(teams) {
			log.Error().Msgf("cannot verify team membership for %s.", u.Login)
			c.Redirect(http.StatusSeeOther, "/login?error=access_denied")
			return
		}
	}

	if err := _store.UpdateUser(u); err != nil {
		log.Error().Msgf("cannot update %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	tokenString, err := token.New(token.SessToken, u.Login).SignExpires(u.Hash, exp)
	if err != nil {
		log.Error().Msgf("cannot create token for %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
		return
	}

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenString)

	c.Redirect(http.StatusSeeOther, "/")
}

func GetLogout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	httputil.DelCookie(c.Writer, c.Request, "user_last")
	c.Redirect(http.StatusSeeOther, "/")
}

func GetLoginToken(c *gin.Context) {
	_store := store.FromContext(c)
	forge := session.Forge(c)

	in := &tokenPayload{}
	err := c.Bind(in)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	login, err := forge.Auth(c, in.Access, in.Refresh)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	user, err := _store.GetUserLogin(login)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	newToken := token.New(token.SessToken, user.Login)
	tokenStr, err := newToken.SignExpires(user.Hash, exp)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &tokenPayload{
		Access:  tokenStr,
		Expires: exp - time.Now().Unix(),
	})
}

type tokenPayload struct {
	Access  string `json:"access_token,omitempty"`
	Refresh string `json:"refresh_token,omitempty"`
	Expires int64  `json:"expires_in,omitempty"`
}

// ToConfig returns the config from the Context
func ToConfig(c *gin.Context) *model.Settings {
	v := c.MustGet("config")
	return v.(*model.Settings)
}
