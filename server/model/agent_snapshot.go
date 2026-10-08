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

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
)

// AgentSnapshot is the state of an agent at the time it got a workflow assigned, so it is
// still known after the agent got changed or deleted, e.g. by an autoscaler. Identical
// snapshots are stored once (identified by ContentHash) and referenced by all their
// workflows, also across agents with the same metadata. OrgID decides who may read it,
// like for the agent itself.
type AgentSnapshot struct {
	ID           int64             `json:"-"             xorm:"pk autoincr 'id'"`
	Hash         string            `json:"-"             xorm:"UNIQUE 'hash'"`
	OrgID        int64             `json:"org_id"        xorm:"org_id"`
	Name         string            `json:"name"          xorm:"name"`
	Platform     string            `json:"platform"      xorm:"platform"`
	Backend      string            `json:"backend"       xorm:"backend"`
	CustomLabels map[string]string `json:"custom_labels" xorm:"json 'custom_labels'"`

	// AgentID is the id of the agent that got the workflow assigned. It is not stored with
	// the snapshot but taken from the workflow, so agents with the same metadata share one.
	AgentID int64 `json:"id" xorm:"-"`
} //	@name	AgentSnapshot

// TableName return database table name for xorm.
func (AgentSnapshot) TableName() string {
	return "agent_snapshots"
}

// ContentHash identifies the snapshot by its stored content. ID and Hash (the row's
// identity) and AgentID (not stored) are never part of it, and missing and empty labels
// are the same. As encoding/json sorts map keys, equal content gets the same hash.
func (s *AgentSnapshot) ContentHash() (string, error) {
	content := *s
	content.ID, content.Hash, content.AgentID = 0, "", 0
	if len(content.CustomLabels) == 0 {
		content.CustomLabels = nil
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw)), nil
}

// NewAgentSnapshot takes a snapshot of the given agent without any secrets and without
// its id, which the workflow keeps.
func NewAgentSnapshot(agent *Agent) *AgentSnapshot {
	return &AgentSnapshot{
		OrgID:        agent.OrgID,
		Name:         agent.Name,
		Platform:     agent.Platform,
		Backend:      agent.Backend,
		CustomLabels: maps.Clone(agent.CustomLabels),
	}
}
