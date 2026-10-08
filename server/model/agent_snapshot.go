// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import "maps"

// AgentSnapshot is the state of an agent at the time it got a workflow assigned, so it is
// still known after the agent got changed or deleted, e.g. by an autoscaler. As agents
// rarely change, identical snapshots are stored once (identified by a hash of their
// content) and referenced by all their workflows. OrgID decides who may read it, like
// for the agent itself.
type AgentSnapshot struct {
	ID           int64             `json:"-"             xorm:"pk autoincr 'id'"`
	Hash         string            `json:"-"             xorm:"UNIQUE 'hash'"`
	AgentID      int64             `json:"id"            xorm:"agent_id"`
	OrgID        int64             `json:"org_id"        xorm:"org_id"`
	Name         string            `json:"name"          xorm:"name"`
	Platform     string            `json:"platform"      xorm:"platform"`
	Backend      string            `json:"backend"       xorm:"backend"`
	CustomLabels map[string]string `json:"custom_labels" xorm:"json 'custom_labels'"`
} //	@name	AgentSnapshot

// TableName return database table name for xorm.
func (AgentSnapshot) TableName() string {
	return "agent_snapshots"
}

// NewAgentSnapshot takes a snapshot of the given agent without any secrets.
func NewAgentSnapshot(agent *Agent) *AgentSnapshot {
	return &AgentSnapshot{
		AgentID:      agent.ID,
		OrgID:        agent.OrgID,
		Name:         agent.Name,
		Platform:     agent.Platform,
		Backend:      agent.Backend,
		CustomLabels: maps.Clone(agent.CustomLabels),
	}
}
