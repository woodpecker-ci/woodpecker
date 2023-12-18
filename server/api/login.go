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

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/httputil"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

func HandleLogin(c *gin.Context) {
	if err := c.Request.FormValue("error"); err != "" {
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login/error?code="+err)
	} else {
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/authorize")
	}
}

func HandleAuth(c *gin.Context) {
	_store := store.FromContext(c)
	_forge := server.Config.Services.Forge

	// when dealing with redirects we may need to adjust the content type. I
	// cannot, however, remember why, so need to revisit this line.
	c.Writer.Header().Del("Content-Type")

	tmpuser, err := _forge.Login(c, c.Writer, c.Request)
	if err != nil {
		log.Error().Msgf("cannot authenticate user. %s", err)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=oauth_error")
		return
	}
	// this will happen when the user is redirected by the forge as
	// part of the authorization workflow.
	if tmpuser == nil {
		return
	}

	// get the user from the database
	u, err := _store.GetUserRemoteID(tmpuser.ForgeRemoteID, tmpuser.Login)
	if err != nil {
		if !errors.Is(err, types.RecordNotExist) {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// if self-registration is disabled we should return a not authorized error
		if !server.Config.Permissions.Open && !server.Config.Permissions.Admins.IsAdmin(tmpuser) {
			log.Error().Msgf("cannot register %s. registration closed", tmpuser.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
			return
		}

		// if self-registration is enabled for allowed organizations we need to
		// check the user's organization membership.
		if server.Config.Permissions.Orgs.IsConfigured {
			teams, terr := _forge.Teams(c, tmpuser)
			if terr != nil || !server.Config.Permissions.Orgs.IsMember(teams) {
				log.Error().Err(terr).Msgf("cannot verify team membership for %s.", u.Login)
				c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
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
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}

		// if another user already have activated repos on behave of that user,
		// the user was stored as org. now we adopt it to the user.
		if org, err := _store.OrgFindByName(u.Login); err == nil && org != nil {
			org.IsUser = true
			if err := _store.OrgUpdate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not mark org as user")
			}
		}
	}

	// update the user meta data and authorization data.
	u.Token = tmpuser.Token
	u.Secret = tmpuser.Secret
	u.Email = tmpuser.Email
	u.Avatar = tmpuser.Avatar
	u.ForgeRemoteID = tmpuser.ForgeRemoteID
	u.Login = tmpuser.Login
	u.Admin = u.Admin || server.Config.Permissions.Admins.IsAdmin(tmpuser)

	// if self-registration is enabled for allowed organizations we need to
	// check the user's organization membership.
	if server.Config.Permissions.Orgs.IsConfigured {
		teams, terr := _forge.Teams(c, u)
		if terr != nil || !server.Config.Permissions.Orgs.IsMember(teams) {
			log.Error().Err(terr).Msgf("cannot verify team membership for %s.", u.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
			return
		}
	}

	if err := _store.UpdateUser(u); err != nil {
		log.Error().Msgf("cannot update %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	tokenString, err := token.New(token.SessToken, u.Login).SignExpires(u.Hash, exp)
	if err != nil {
		log.Error().Msgf("cannot create token for %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	repos, _ := _forge.Repos(c, u)
	for _, forgeRepo := range repos {
		dbRepo, err := _store.GetRepoForgeID(forgeRepo.ForgeRemoteID)
		if err != nil && errors.Is(err, types.RecordNotExist) {
			continue
		}
		if err != nil {
			log.Error().Msgf("cannot list repos for %s. %s", u.Login, err)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}

		if !dbRepo.IsActive {
			continue
		}

		log.Debug().Msgf("Synced user permission for %s %s", u.Login, dbRepo.FullName)
		perm := forgeRepo.Perm
		perm.Repo = dbRepo
		perm.RepoID = dbRepo.ID
		perm.UserID = u.ID
		perm.Synced = time.Now().Unix()
		if err := _store.PermUpsert(perm); err != nil {
			log.Error().Msgf("cannot update permissions for %s. %s", u.Login, err)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}
	}

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenString)

	c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/")
}

func GetLogout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	httputil.DelCookie(c.Writer, c.Request, "user_last")
	c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/")
}

func GetLoginToken(c *gin.Context) {
	_store := store.FromContext(c)

	in := &tokenPayload{}
	err := c.Bind(in)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	login, err := server.Config.Services.Forge.Auth(c, in.Access, in.Refresh)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	user, err := _store.GetUserLogin(login)
	if err != nil {
		handleDBError(c, err)
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
