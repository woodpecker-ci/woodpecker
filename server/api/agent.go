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
	"encoding/base32"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func GetAgents(c *gin.Context) {
	agents, err := store.FromContext(c).AgentList()
	if err != nil {
		c.String(500, "Error getting agent list. %s", err)
		return
	}
	c.JSON(http.StatusOK, agents)
}

func GetAgent(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := store.FromContext(c).AgentFind(agentID)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	c.JSON(http.StatusOK, agent)
}

func PatchAgent(c *gin.Context) {
	_store := store.FromContext(c)

	in := &model.Agent{}
	err := c.Bind(in)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	agentID, err := strconv.ParseInt(c.Param("agent"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	agent.Name = in.Name
	agent.NoSchedule = in.NoSchedule

	err = _store.AgentUpdate(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, agent)
}

// PostAgent create a new agent with a random token so a new agent can connect to the server
func PostAgent(c *gin.Context) {
	in := &model.Agent{}
	err := c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	user := session.User(c)

	agent := &model.Agent{
		Name:       in.Name,
		NoSchedule: in.NoSchedule,
		OwnerID:    user.ID,
		Token: base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		),
	}
	if err = store.FromContext(c).AgentCreate(agent); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, agent)
}

func DeleteAgent(c *gin.Context) {
	_store := store.FromContext(c)

	agentID, err := strconv.ParseInt(c.Param("agent"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		handleDbGetError(c, err)
		return
	}
	if err = _store.AgentDelete(agent); err != nil {
		c.String(http.StatusInternalServerError, "Error deleting user. %s", err)
		return
	}
	c.String(http.StatusNoContent, "")
}
