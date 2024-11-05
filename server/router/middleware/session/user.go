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

package session

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/shared/token"
)

func User(c *gin.Context) *model.User {
	v, ok := c.Get("user")
	if !ok {
		return nil
	}
	u, ok := v.(*model.User)
	if !ok {
		return nil
	}
	return u
}

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *model.User

		t, err := token.ParseRequest([]token.Type{token.UserToken, token.SessToken}, c.Request, func(t *token.Token) (string, error) {
			var err error
			userID, err := strconv.ParseInt(t.Get("user-id"), 10, 64)
			if err != nil {
				return "", err
			}
			user, err = store.FromContext(c).GetUser(userID)
			return user.Hash, err
		})
		if err == nil {
			c.Set("user", user)

			// if this is a session token (ie not the API token)
			// this means the user is accessing with a web browser,
			// so we should implement CSRF protection measures.
			if t.Type == token.SessToken {
				err = token.CheckCsrf(c.Request, func(_ *token.Token) (string, error) {
					return user.Hash, nil
				})
				// if csrf token validation fails, exit immediately
				// with a not authorized error.
				if err != nil {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}
		}
		c.Next()
	}
}

func MustAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		switch {
		case user == nil:
			c.String(http.StatusUnauthorized, "User not authorized")
			c.Abort()
		case !user.Admin:
			c.String(http.StatusForbidden, "User not authorized")
			c.Abort()
		default:
			c.Next()
		}
	}
}

func MustRepoAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		perm := Perm(c)
		switch {
		case user == nil:
			c.String(http.StatusUnauthorized, "User not authorized")
			c.Abort()
		case !perm.Admin:
			c.String(http.StatusForbidden, "User not authorized")
			c.Abort()
		default:
			c.Next()
		}
	}
}

func MustUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		switch {
		case user == nil:
			c.String(http.StatusUnauthorized, "User not authorized")
			c.Abort()
		default:
			c.Next()
		}
	}
}

func MustOrgMember(admin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		if user == nil {
			c.String(http.StatusUnauthorized, "User not authorized")
			c.Abort()
			return
		}

		org := Org(c)
		if org == nil {
			c.String(http.StatusBadRequest, "Organization not loaded")
			c.Abort()
			return
		}

		// User can access his own, admin can access all
		if (org.Name == user.Login) || user.Admin {
			c.Next()
			return
		}

		_forge, err := server.Config.Services.Manager.ForgeFromUser(user)
		if err != nil {
			log.Error().Err(err).Msg("Cannot get forge from user")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		perm, err := server.Config.Services.Membership.Get(c, _forge, user, org.Name)
		if err != nil {
			log.Error().Err(err).Msg("failed to check membership")
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			c.Abort()
			return
		}

		if perm == nil || (!admin && !perm.Member) || (admin && !perm.Admin) {
			c.String(http.StatusForbidden, "user not authorized")
			c.Abort()
			return
		}

		c.Next()
	}
}
