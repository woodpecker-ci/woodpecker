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

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/server/store"
)

// GetUsers
//
//	@Summary		Get all users
//	@Description	Returns all registered, active users in the system. Requires admin rights.
//	@Router			/users [get]
//	@Produce		json
//	@Success		200	{array}	User
//	@Tags			Users
//	@Param			Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param			page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param			perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetUsers(c *gin.Context) {
	users, err := store.FromContext(c).GetUserList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting user list. %s", err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser
//
//	@Summary		Get a user
//	@Description	Returns a user with the specified login name. Requires admin rights.
//	@Router			/users/{login} [get]
//	@Produce		json
//	@Success		200	{object}	User
//	@Tags			Users
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			login			path	string	true	"the user's login name"
func GetUser(c *gin.Context) {
	user, err := store.FromContext(c).GetUserLogin(c.Param("login"))
	if err != nil {
		handleDbError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// PatchUser
//
//	@Summary		Change a user
//	@Description	Changes the data of an existing user. Requires admin rights.
//	@Router			/users/{login} [patch]
//	@Produce		json
//	@Accept			json
//	@Success		200	{object}	User
//	@Tags			Users
//	@Param			Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			login			path	string		true	"the user's login name"
//	@Param			user			body	User	true	"the user's data"
func PatchUser(c *gin.Context) {
	_store := store.FromContext(c)

	in := &model.User{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, err := _store.GetUserLogin(c.Param("login"))
	if err != nil {
		handleDbError(c, err)
		return
	}

	// TODO: allow to change login (currently used as primary key)
	// TODO: disallow to change login, email, avatar if the user is using oauth
	user.Email = in.Email
	user.Avatar = in.Avatar
	user.Admin = in.Admin

	err = _store.UpdateUser(user)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, user)
}

// PostUser
//
//	@Summary		Create a user
//	@Description	Creates a new user account with the specified external login. Requires admin rights.
//	@Router			/users [post]
//	@Produce		json
//	@Success		200	{object}	User
//	@Tags			Users
//	@Param			Authorization	header	string		true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			user			body	User	true	"the user's data"
func PostUser(c *gin.Context) {
	in := &model.User{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	user := &model.User{
		Login:  in.Login,
		Email:  in.Email,
		Avatar: in.Avatar,
		Hash: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
	}
	if err = user.Validate(); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err = store.FromContext(c).CreateUser(user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUser
//
//	@Summary		Delete a user
//	@Description	Deletes the given user. Requires admin rights.
//	@Router			/users/{login} [delete]
//	@Produce		plain
//	@Success		204
//	@Tags			Users
//	@Param			Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param			login			path	string	true	"the user's login name"
func DeleteUser(c *gin.Context) {
	_store := store.FromContext(c)

	user, err := _store.GetUserLogin(c.Param("login"))
	if err != nil {
		handleDbError(c, err)
		return
	}
	if err = _store.DeleteUser(user); err != nil {
		handleDbError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
