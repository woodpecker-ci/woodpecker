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
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/httputil"
	"github.com/woodpecker-ci/woodpecker/shared/token"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandleLogin(c *gin.Context) {
	var (
		w = c.Writer
		r = c.Request
	)
	if err := r.FormValue("error"); err != "" {
		http.Redirect(w, r, "/login/error?code="+err, 303)
	} else {
		intendedURL := r.URL.Query()["url"]
		if len(intendedURL) > 0 {
			http.Redirect(w, r, "/authorize?url="+intendedURL[0], 303)
		} else {
			http.Redirect(w, r, "/authorize", 303)
		}
	}
}

func HandleAuth(c *gin.Context) {

	// when dealing with redirects we may need to adjust the content type. I
	// cannot, however, remember why, so need to revisit this line.
	c.Writer.Header().Del("Content-Type")

	tmpuser, err := remote.Login(c, c.Writer, c.Request)
	if err != nil {
		logrus.Errorf("cannot authenticate user. %s", err)
		c.Redirect(303, "/login?error=oauth_error")
		return
	}
	// this will happen when the user is redirected by the remote provider as
	// part of the authorization workflow.
	if tmpuser == nil {
		return
	}
	config := ToConfig(c)

	// get the user from the database
	u, err := store.GetUserLogin(c, tmpuser.Login)
	if err != nil {

		// if self-registration is disabled we should return a not authorized error
		if !config.Open && !config.IsAdmin(tmpuser) {
			logrus.Errorf("cannot register %s. registration closed", tmpuser.Login)
			c.Redirect(303, "/login?error=access_denied")
			return
		}

		// if self-registration is enabled for whitelisted organizations we need to
		// check the user's organization membership.
		if len(config.Orgs) != 0 {
			teams, terr := remote.Teams(c, tmpuser)
			if terr != nil || config.IsMember(teams) == false {
				logrus.Errorf("cannot verify team membership for %s.", u.Login)
				c.Redirect(303, "/login?error=access_denied")
				return
			}
		}

		// create the user account
		u = &model.User{
			Login:  tmpuser.Login,
			Token:  tmpuser.Token,
			Secret: tmpuser.Secret,
			Email:  tmpuser.Email,
			Avatar: tmpuser.Avatar,
			Hash: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
		}

		// insert the user into the database
		if err := store.CreateUser(c, u); err != nil {
			logrus.Errorf("cannot insert %s. %s", u.Login, err)
			c.Redirect(303, "/login?error=internal_error")
			return
		}
	}

	// update the user meta data and authorization data.
	u.Token = tmpuser.Token
	u.Secret = tmpuser.Secret
	u.Email = tmpuser.Email
	u.Avatar = tmpuser.Avatar

	// if self-registration is enabled for whitelisted organizations we need to
	// check the user's organization membership.
	if len(config.Orgs) != 0 {
		teams, terr := remote.Teams(c, u)
		if terr != nil || config.IsMember(teams) == false {
			logrus.Errorf("cannot verify team membership for %s.", u.Login)
			c.Redirect(303, "/login?error=access_denied")
			return
		}
	}

	if err := store.UpdateUser(c, u); err != nil {
		logrus.Errorf("cannot update %s. %s", u.Login, err)
		c.Redirect(303, "/login?error=internal_error")
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	tokenString, err := token.New(token.SessToken, u.Login).SignExpires(u.Hash, exp)
	if err != nil {
		logrus.Errorf("cannot create token for %s. %s", u.Login, err)
		c.Redirect(303, "/login?error=internal_error")
		return
	}

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenString)

	intendedURL := c.Request.URL.Query()["url"]
	if len(intendedURL) > 0 {
		c.Redirect(303, intendedURL[0])
	} else {
		c.Redirect(303, "/")
	}
}

func GetLogout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	httputil.DelCookie(c.Writer, c.Request, "user_last")
	c.Redirect(303, "/")
}

func GetLoginToken(c *gin.Context) {
	in := &tokenPayload{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	login, err := remote.Auth(c, in.Access, in.Refresh)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	user, err := store.GetUserLogin(c, login)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	token := token.New(token.SessToken, user.Login)
	tokenstr, err := token.SignExpires(user.Hash, exp)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &tokenPayload{
		Access:  tokenstr,
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
