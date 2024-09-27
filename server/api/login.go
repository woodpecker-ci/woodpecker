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
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/httputil"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

const stateTokenDuration = time.Minute * 5

func HandleAuth(c *gin.Context) {
	// TODO: check if this is really needed
	c.Writer.Header().Del("Content-Type")

	// redirect when getting oauth error from forge to login page
	if err := c.Request.FormValue("error"); err != "" {
		query := url.Values{}
		query.Set("error", err)
		if errorDescription := c.Request.FormValue("error_description"); errorDescription != "" {
			query.Set("error_description", errorDescription)
		}
		if errorURI := c.Request.FormValue("error_uri"); errorURI != "" {
			query.Set("error_uri", errorURI)
		}
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("%s/login?%s", server.Config.Server.RootPath, query.Encode()))
		return
	}

	_store := store.FromContext(c)

	code := c.Request.FormValue("code")
	state := c.Request.FormValue("state")
	isCallback := code != "" && state != ""
	var forgeID int64

	if isCallback { // validate the state token
		stateToken, err := token.Parse([]token.Type{token.OAuthStateToken}, state, func(_ *token.Token) (string, error) {
			return server.Config.Server.JWTSecret, nil
		})
		if err != nil {
			log.Error().Err(err).Msg("cannot verify state token")
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=invalid_state")
			return
		}

		_forgeID := stateToken.Get("forge-id")
		forgeID, err = strconv.ParseInt(_forgeID, 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("forge-id of state token invalid")
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=invalid_state")
			return
		}
	} else { // only generate a state token if not a callback
		var err error

		_forgeID := c.Request.FormValue("forge_id")
		if _forgeID == "" {
			forgeID = 1 // fallback to main forge
		} else {
			forgeID, err = strconv.ParseInt(_forgeID, 10, 64)
			if err != nil {
				log.Error().Err(err).Msg("forge-id of state token invalid")
				c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=invalid_state")
				return
			}
		}

		jwtSecret := server.Config.Server.JWTSecret
		exp := time.Now().Add(stateTokenDuration).Unix()
		stateToken := token.New(token.OAuthStateToken)
		stateToken.Set("forge-id", strconv.FormatInt(forgeID, 10))
		state, err = stateToken.SignExpires(jwtSecret, exp)
		if err != nil {
			log.Error().Err(err).Msg("cannot create state token")
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}
	}

	_forge, err := server.Config.Services.Manager.ForgeByID(forgeID)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot get forge by id %d", forgeID)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	userFromForge, redirectURL, err := _forge.Login(c, &forge_types.OAuthRequest{
		Code:  c.Request.FormValue("code"),
		State: state,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot authenticate user")
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=oauth_error")
		return
	}
	// The user is not authorized yet -> redirect
	if userFromForge == nil {
		http.Redirect(c.Writer, c.Request, redirectURL, http.StatusSeeOther)
		return
	}

	// if organization filter is enabled, we need to check if the user is a member of one
	// of the configured organizations
	if server.Config.Permissions.Orgs.IsConfigured {
		teams, terr := _forge.Teams(c, userFromForge)
		if terr != nil || !server.Config.Permissions.Orgs.IsMember(teams) {
			log.Error().Err(terr).Msgf("cannot verify team membership for %s", userFromForge.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=org_access_denied")
			return
		}
	}

	// get the user from the database
	user, err := _store.GetUserRemoteID(userFromForge.ForgeRemoteID, userFromForge.Login)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		log.Error().Err(err).Msgf("cannot get user %s", userFromForge.Login)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	if user == nil || errors.Is(err, types.RecordNotExist) {
		// if self-registration is disabled we should return a not authorized error
		if !server.Config.Permissions.Open && !server.Config.Permissions.Admins.IsAdmin(userFromForge) {
			log.Error().Msgf("cannot register %s. registration closed", userFromForge.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=registration_closed")
			return
		}

		// create the user account
		user = &model.User{
			Login:         userFromForge.Login,
			ForgeRemoteID: userFromForge.ForgeRemoteID,
			Token:         userFromForge.Token,
			Secret:        userFromForge.Secret,
			Email:         userFromForge.Email,
			Avatar:        userFromForge.Avatar,
			ForgeID:       forgeID,
			Hash: base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			),
		}

		// insert the user into the database
		if err := _store.CreateUser(user); err != nil {
			log.Error().Err(err).Msgf("cannot insert %s", user.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}
	}

	// create or set the user's organization if it isn't linked yet
	if user.OrgID == 0 {
		// check if an org with the same name exists already and assign it to the user if it does
		if org, err := _store.OrgFindByName(user.Login); err == nil && org != nil {
			org.IsUser = true
			user.OrgID = org.ID

			if err := _store.OrgUpdate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not mark org as user")
			}
		}
		if err != nil && !errors.Is(err, types.RecordNotExist) {
			log.Error().Err(err).Msgf("cannot get org %s", user.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}

		if user.OrgID == 0 {
			org := &model.Org{
				Name:    user.Login,
				IsUser:  true,
				Private: false,
				ForgeID: user.ForgeID,
			}
			if err := _store.OrgCreate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not create org for user")
			}
			user.OrgID = org.ID
		}
	} else {
		// update org name if necessary
		org, err := _store.OrgGet(user.OrgID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot get org %s", user.Login)
			c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
			return
		}
		if org != nil && org.Name != user.Login {
			org.Name = user.Login
			if err := _store.OrgUpdate(org); err != nil {
				log.Error().Err(err).Msgf("on user creation, could not mark org as user")
			}
		}
	}

	// update the user meta data and authorization data.
	user.Token = userFromForge.Token
	user.Secret = userFromForge.Secret
	user.Email = userFromForge.Email
	user.Avatar = userFromForge.Avatar
	user.ForgeID = forgeID
	user.ForgeRemoteID = userFromForge.ForgeRemoteID
	user.Login = userFromForge.Login
	user.Admin = user.Admin || server.Config.Permissions.Admins.IsAdmin(userFromForge)

	if err := _store.UpdateUser(user); err != nil {
		log.Error().Err(err).Msgf("cannot update %s", user.Login)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	exp := time.Now().Add(server.Config.Server.SessionExpires).Unix()
	_token := token.New(token.SessToken)
	_token.Set("user-id", strconv.FormatInt(user.ID, 10))
	tokenString, err := _token.SignExpires(user.Hash, exp)
	if err != nil {
		log.Error().Msgf("cannot create token for %s", user.Login)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	err = updateRepoPermissions(c, user, _store, _forge)
	if err != nil {
		log.Error().Err(err).Msgf("cannot update repo permissions for %s", user.Login)
		c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/login?error=internal_error")
		return
	}

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenString)

	c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/")
}

func updateRepoPermissions(c *gin.Context, user *model.User, _store store.Store, _forge forge.Forge) error {
	repos, _ := _forge.Repos(c, user)

	for _, forgeRepo := range repos {
		dbRepo, err := _store.GetRepoForgeID(forgeRepo.ForgeRemoteID)
		if err != nil && errors.Is(err, types.RecordNotExist) {
			continue
		}
		if err != nil {
			return err
		}

		if !dbRepo.IsActive {
			continue
		}

		log.Debug().Msgf("synced user permission for %s %s", user.Login, dbRepo.FullName)
		perm := forgeRepo.Perm
		perm.Repo = dbRepo
		perm.RepoID = dbRepo.ID
		perm.UserID = user.ID
		perm.Synced = time.Now().Unix()
		if err := _store.PermUpsert(perm); err != nil {
			return err
		}
	}

	return nil
}

func GetLogout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	httputil.DelCookie(c.Writer, c.Request, "user_last")
	c.Redirect(http.StatusSeeOther, server.Config.Server.RootPath+"/")
}
