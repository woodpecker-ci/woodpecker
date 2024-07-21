package woodpecker

import "fmt"

const (
	pathAgents     = "%s/api/agents"
	pathAgent      = "%s/api/agents/%d"
	pathAgentTasks = "%s/api/agents/%d/tasks"
)

// AgentCreate creates a new agent.
func (c *client) AgentCreate(in *Agent) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgents, c.addr)
	return out, c.post(uri, in, out)
}

// AgentList returns a list of all registered agents.
func (c *client) AgentList() ([]*Agent, error) {
	out := make([]*Agent, 0, 5)
	uri := fmt.Sprintf(pathAgents, c.addr)
	return out, c.get(uri, &out)
}

// Agent returns an agent by id.
func (c *client) Agent(agentID int64) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgent, c.addr, agentID)
	return out, c.get(uri, out)
}

// AgentUpdate updates the agent with the provided Agent struct.
func (c *client) AgentUpdate(in *Agent) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgent, c.addr, in.ID)
	return out, c.patch(uri, in, out)
}

// AgentDelete deletes the agent with the given id.
func (c *client) AgentDelete(agentID int64) error {
	uri := fmt.Sprintf(pathAgent, c.addr, agentID)
	return c.delete(uri)
}

// AgentTasksList returns a list of all tasks for the agent with the given id.
func (c *client) AgentTasksList(agentID int64) ([]*Task, error) {
	out := make([]*Task, 0, 5)
	uri := fmt.Sprintf(pathAgentTasks, c.addr, agentID)
	return out, c.get(uri, &out)
}
