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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

//
// Global Agents.
//

// GetAgents
//
//	@Summary	List agents
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
//	@Summary	Get an agent
//	@Router		/agents/{agent_id} [get]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
func GetAgent(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := store.FromContext(c).AgentFind(agentID)
	if err != nil {
		handleDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, agent)
}

// GetAgentTasks
//
//	@Summary	List agent tasks
//	@Router		/agents/{agent_id}/tasks [get]
//	@Produce	json
//	@Success	200	{array}	Task
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
func GetAgentTasks(c *gin.Context) {
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := store.FromContext(c).AgentFind(agentID)
	if err != nil {
		handleDBError(c, err)
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
//	@Summary	Update an agent
//	@Router		/agents/{agent_id} [patch]
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

	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// Update allowed fields
	agent.Name = in.Name
	agent.NoSchedule = in.NoSchedule
	if agent.NoSchedule {
		server.Config.Services.Queue.KickAgentWorkers(agent.ID)
	}

	err = _store.AgentUpdate(agent)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, agent)
}

// PostAgent
//
//	@Summary	Create a new agent
//	@Description Creates a new agent with a random token
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
		OwnerID:    user.ID,
		OrgID:      model.IDNotSet,
		NoSchedule: in.NoSchedule,
		Token:      model.GenerateNewAgentToken(),
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
//	@Router		/agents/{agent_id} [delete]
//	@Produce	plain
//	@Success	200
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		agent			path	int		true	"the agent's id"
func DeleteAgent(c *gin.Context) {
	_store := store.FromContext(c)

	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		handleDBError(c, err)
		return
	}

	// prevent deletion of agents with running tasks
	info := server.Config.Services.Queue.Info(c)
	for _, task := range info.Running {
		if task.AgentID == agent.ID {
			c.String(http.StatusConflict, "Agent has running tasks")
			return
		}
	}

	// kick workers to remove the agent from the queue
	server.Config.Services.Queue.KickAgentWorkers(agent.ID)

	if err = _store.AgentDelete(agent); err != nil {
		c.String(http.StatusInternalServerError, "Error deleting user. %s", err)
		return
	}
	c.Status(http.StatusNoContent)
}

//
// Org/User Agents.
//

// PostOrgAgent
//
//	@Summary	Create a new organization-scoped agent
//	@Description Creates a new agent with a random token, scoped to the specified organization
//	@Router		/orgs/{org_id}/agents [post]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	int		true	"the organization's id"
//	@Param		agent			body	Agent	true	"the agent's data (only 'name' and 'no_schedule' are read)"
func PostOrgAgent(c *gin.Context) {
	_store := store.FromContext(c)
	user := session.User(c)

	orgID, err := strconv.ParseInt(c.Param("org_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid organization ID")
		return
	}

	in := new(model.Agent)
	err = c.Bind(in)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	agent := &model.Agent{
		Name:       in.Name,
		OwnerID:    user.ID,
		OrgID:      orgID,
		NoSchedule: in.NoSchedule,
		Token:      model.GenerateNewAgentToken(),
	}

	if err = _store.AgentCreate(agent); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, agent)
}

// GetOrgAgents
//
//	@Summary	List agents for an organization
//	@Router		/orgs/{org_id}/agents [get]
//	@Produce	json
//	@Success	200	{array}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"				default(Bearer <personal access token>)
//	@Param		org_id			path	int		true	"the organization's id"
//	@Param		page			query	int		false	"for response pagination, page offset number"	default(1)
//	@Param		perPage			query	int		false	"for response pagination, max items per page"	default(50)
func GetOrgAgents(c *gin.Context) {
	_store := store.FromContext(c)
	org := session.Org(c)

	agents, err := _store.AgentListForOrg(org.ID, session.Pagination(c))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting agent list. %s", err)
		return
	}

	c.JSON(http.StatusOK, agents)
}

// PatchOrgAgent
//
//	@Summary	Update an organization-scoped agent
//	@Router		/orgs/{org_id}/agents/{agent_id} [patch]
//	@Produce	json
//	@Success	200	{object}	Agent
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	int		true	"the organization's id"
//	@Param		agent_id		path	int		true	"the agent's id"
//	@Param		agent			body	Agent	true	"the agent's updated data"
func PatchOrgAgent(c *gin.Context) {
	_store := store.FromContext(c)
	org := session.Org(c)

	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid agent ID")
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		c.String(http.StatusNotFound, "Agent not found")
		return
	}

	if agent.OrgID != org.ID {
		c.String(http.StatusBadRequest, "Agent does not belong to this organization")
		return
	}

	in := new(model.Agent)
	if err := c.Bind(in); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Update allowed fields
	agent.Name = in.Name
	agent.NoSchedule = in.NoSchedule
	if agent.NoSchedule {
		server.Config.Services.Queue.KickAgentWorkers(agent.ID)
	}

	if err := _store.AgentUpdate(agent); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, agent)
}

// DeleteOrgAgent
//
//	@Summary	Delete an organization-scoped agent
//	@Router		/orgs/{org_id}/agents/{agent_id} [delete]
//	@Produce	plain
//	@Success	204
//	@Tags		Agents
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
//	@Param		org_id			path	int		true	"the organization's id"
//	@Param		agent_id		path	int		true	"the agent's id"
func DeleteOrgAgent(c *gin.Context) {
	_store := store.FromContext(c)
	org := session.Org(c)

	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid agent ID")
		return
	}

	agent, err := _store.AgentFind(agentID)
	if err != nil {
		c.String(http.StatusNotFound, "Agent not found")
		return
	}

	if agent.OrgID != org.ID {
		c.String(http.StatusBadRequest, "Agent does not belong to this organization")
		return
	}

	// Check if the agent has any running tasks
	info := server.Config.Services.Queue.Info(c)
	for _, task := range info.Running {
		if task.AgentID == agent.ID {
			c.String(http.StatusConflict, "Agent has running tasks")
			return
		}
	}

	// Kick workers to remove the agent from the queue
	server.Config.Services.Queue.KickAgentWorkers(agent.ID)

	if err := _store.AgentDelete(agent); err != nil {
		c.String(http.StatusInternalServerError, "Error deleting agent. %s", err)
		return
	}

	c.Status(http.StatusNoContent)
}
