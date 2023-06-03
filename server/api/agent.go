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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware/session"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// GetAgents
//
//	@Summary	Get agent list
//	@Router		/agents [get]
//	@Produce	json
//	@Success	200	{array}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetAgents(c *gin.Context) {
	agents, err := store.FromContext(c).AgentList(session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting agent list. %s", err)
		return
	}
	c.JSON(http.StatusOK, agents)
}

// GetAgent
//
//	@Summary	Get agent information
//	@Router		/agents/{agent} [get]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
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

// GetAgentTasks
//
//	@Summary	Get agent tasks
//	@Router		/agents/{agent}/tasks [get]
//	@Produce	json
//	@Success	200	{array}	Task
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
func GetAgentTasks(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := store.FromContext(c).AgentFind(agentID)
	if err != nil {
		c.String(http.StatusNotFound, "Cannot find agent. %s", err)
		return
	}

	var tasks []*model.Task
	info := server.Config.Services.Queue.Info(c)
	for _, task := range info.Running {
		if task.AgentID == agent.ID {
			tasks = append(tasks, task)
		}
	}

	c.JSON(http.StatusOK, tasks)
}

// PatchAgent
//
//	@Summary	Update agent information
//	@Router		/agents/{agent} [patch]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
//	@Param		agentData		body	Agent	true	"the agent's data"
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

// PostAgent
//
//	@Summary	Create a new agent with a random token so a new agent can connect to the server
//	@Router		/agents [post]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			body	Agent	true	"the agent's data (only 'name' and 'no_schedule' are read)"
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

// DeleteAgent
//
//	@Summary	Delete an agent
//	@Router		/agents/{agent} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
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
