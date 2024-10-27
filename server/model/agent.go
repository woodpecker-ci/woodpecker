// Copyright 2021 Woodpecker Authors
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

package model

import (
	"encoding/base32"
	"fmt"

	"github.com/gorilla/securecookie"
)

type Agent struct {
	ID           int64             `json:"id"            xorm:"pk autoincr 'id'"`
	Created      int64             `json:"created"       xorm:"created"`
	Updated      int64             `json:"updated"       xorm:"updated"`
	Name         string            `json:"name"          xorm:"name"`
	OwnerID      int64             `json:"owner_id"      xorm:"'owner_id'"`
	Token        string            `json:"token"         xorm:"token"`
	LastContact  int64             `json:"last_contact"  xorm:"last_contact"`
	LastWork     int64             `json:"last_work"     xorm:"last_work"` // last time the agent did something, this value is used to determine if the agent is still doing work used by the autoscaler
	Platform     string            `json:"platform"      xorm:"VARCHAR(100) 'platform'"`
	Backend      string            `json:"backend"       xorm:"VARCHAR(100) 'backend'"`
	Capacity     int32             `json:"capacity"      xorm:"capacity"`
	Version      string            `json:"version"       xorm:"'version'"`
	NoSchedule   bool              `json:"no_schedule"   xorm:"no_schedule"`
	CustomLabels map[string]string `json:"custom_labels" xorm:"JSON 'custom_labels'"`
	// OrgID is counted as unset if set to -1, this is done to ensure a new(Agent) still enforce the OrgID check by default
	OrgID int64 `json:"org_id"        xorm:"INDEX 'org_id'"`
} //	@name Agent

const (
	IDNotSet         = -1
	agentFilterOrgID = "org-id"
)

// TableName return database table name for xorm.
func (Agent) TableName() string {
	return "agents"
}

func (a *Agent) IsSystemAgent() bool {
	return a.OwnerID == IDNotSet
}

func GenerateNewAgentToken() string {
	return base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
}

func (a *Agent) GetServerLabels() (map[string]string, error) {
	filters := make(map[string]string)

	// enforce filters for user and organization agents
	if a.OrgID != IDNotSet {
		filters[agentFilterOrgID] = fmt.Sprintf("%d", a.OrgID)
	} else {
		filters[agentFilterOrgID] = "*"
	}

	return filters, nil
}

func (a *Agent) CanAccessRepo(repo *Repo) bool {
	// global agent
	if a.OrgID == IDNotSet {
		return true
	}

	// agent has access to the organization
	if a.OrgID == repo.OrgID {
		return true
	}

	return false
}
