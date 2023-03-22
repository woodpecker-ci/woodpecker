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
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func GetUsers(c *gin.Context) {
	users, err := store.FromContext(c).GetUserList(session.Pagination(c))
	if err != nil {
		c.String(500, "Error getting user list. %s", err)
		return
	}
	c.JSON(200, users)
}

func GetUser(c *gin.Context) {
	user, err := store.FromContext(c).GetUserLogin(c.Param("login"))
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(200, user)
}

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
		handleDbGetError(c, err)
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

func DeleteUser(c *gin.Context) {
	_store := store.FromContext(c)

	user, err := _store.GetUserLogin(c.Param("login"))
	if err != nil {
		c.String(404, "Cannot find user. %s", err)
		return
	}
	if err = _store.DeleteUser(user); err != nil {
		c.String(500, "Error deleting user. %s", err)
		return
	}
	c.String(200, "")
}
