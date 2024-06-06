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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
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
	_forge, err := server.Config.Services.Manager.ForgeMain()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	forgeID := int64(1) // TODO: replace with forge id when multiple forges are supported

	// when dealing with redirects, we may need to adjust the content type. I
	// cannot, however, remember why, so need to revisit this line.
	c.Writer.Header().Del("Content-Type")

	tmpUser, redirectURL, err := _forge.Login(c, &forge_types.OAuthRequest{
		Error:            c.Request.FormValue("error"),
		ErrorURI:         c.Request.FormValue("error_uri"),
		ErrorDescription: c.Request.FormValue("error_description"),
		Code:             c.Request.FormValue("code"),
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot authenticate user")
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=oauth_error")
		return
	}
	// The user is not authorized yet -> redirect
	if tmpUser == nil {
		http.Redirect(c.Writer, c.Request, redirectURL, http.StatusSeeOther)
		return
	}

	// get the user from the database
	u, err := _store.GetUserRemoteID(tmpUser.ForgeRemoteID, tmpUser.Login)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if errors.Is(err, types.RecordNotExist) {
		// if self-registration is disabled we should return a not authorized error
		if !server.Config.Permissions.Open && !server.Config.Permissions.Admins.IsAdmin(tmpUser) {
			log.Error().Msgf("cannot register %s. registration closed", tmpUser.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
			return
		}

		// if self-registration is enabled for allowed organizations we need to
		// check the user's organization membership.
		if server.Config.Permissions.Orgs.IsConfigured {
			teams, terr := _forge.Teams(c, tmpUser)
			if terr != nil || !server.Config.Permissions.Orgs.IsMember(teams) {
				log.Error().Err(terr).Msgf("cannot verify team membership for %s.", u.Login)
				c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
				return
			}
		}

		// create the user account
		u = &model.User{
			Login:         tmpUser.Login,
			ForgeRemoteID: tmpUser.ForgeRemoteID,
			Token:         tmpUser.Token,
			Secret:        tmpUser.Secret,
			Email:         tmpUser.Email,
			Avatar:        tmpUser.Avatar,
			ForgeID:       forgeID,
			Hash: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
		}

		// insert the user into the database
		if err := _store.CreateUser(u); err != nil {
			log.Error().Err(err).Msgf("cannot insert %s", u.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}

		// if another user already have activated repos on behave of that user,
		// the user was stored as org. now we adopt it to the user.
		if org, err := _store.OrgFindByName(u.Login); err == nil && org != nil {
			org.IsUser = true
			u.OrgID = org.ID
			if err := _store.OrgUpdate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not mark org as user")
			}
		} else {
			if err != nil && !errors.Is(err, types.RecordNotExist) {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			org = &model.Org{
				Name:    u.Login,
				IsUser:  true,
				Private: false,
				ForgeID: u.ForgeID,
			}
			if err := _store.OrgCreate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not create org for user")
			}
			u.OrgID = org.ID
		}
	}

	// update org name
	if u.Login != tmpUser.Login {
		org, err := _store.OrgGet(u.OrgID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot get org %s", u.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}
		org.Name = u.Login
		if err := _store.OrgUpdate(org); err != nil {
			log.Error().Err(err).Msgf("on user creation, could not mark org as user")
		}
	}

	// update the user meta data and authorization data.
	u.Token = tmpUser.Token
	u.Secret = tmpUser.Secret
	u.Email = tmpUser.Email
	u.Avatar = tmpUser.Avatar
	u.ForgeRemoteID = tmpUser.ForgeRemoteID
	u.Login = tmpUser.Login
	u.Admin = u.Admin || server.Config.Permissions.Admins.IsAdmin(tmpUser)

	// if self-registration is enabled for allowed organizations we need to
	// check the user's organization membership.
	if server.Config.Permissions.Orgs.IsConfigured {
		teams, terr := _forge.Teams(c, u)
		if terr != nil || !server.Config.Permissions.Orgs.IsMember(teams) {
			log.Error().Err(terr).Msgf("cannot verify team membership for %s", u.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=access_denied")
			return
		}
	}

	if err := _store.UpdateUser(u); err != nil {
		log.Error().Err(err).Msgf("cannot update %s", u.Login)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	_token := token.New(token.SessToken)
	_token.Set("user-id", strconv.FormatInt(u.ID, 10))
	tokenString, err := _token.SignExpires(u.Hash, exp)
	if err != nil {
		log.Error().Msgf("cannot create token for %s", u.Login)
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
			log.Error().Err(err).Msgf("cannot list repos for %s", u.Login)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}

		if !dbRepo.IsActive {
			continue
		}

		log.Debug().Msgf("synced user permission for %s %s", u.Login, dbRepo.FullName)
		perm := forgeRepo.Perm
		perm.Repo = dbRepo
		perm.RepoID = dbRepo.ID
		perm.UserID = u.ID
		perm.Synced = time.Now().Unix()
		if err := _store.PermUpsert(perm); err != nil {
			log.Error().Err(err).Msgf("cannot update permissions for %s", u.Login)
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

	_forge, err := server.Config.Services.Manager.ForgeMain() // TODO: get selected forge from auth request
	if err != nil {
		log.Error().Err(err).Msg("Cannot get main forge")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	in := &tokenPayload{}
	err = c.Bind(in)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	login, err := _forge.Auth(c, in.Access, in.Refresh)
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
	newToken := token.New(token.SessToken)
	newToken.Set("user-id", strconv.FormatInt(user.ID, 10))
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
