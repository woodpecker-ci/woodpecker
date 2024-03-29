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

import "errors"

type Agent struct {
	ID          int64  `json:"id"            xorm:"pk autoincr 'id'"`
	Created     int64  `json:"created"       xorm:"created"`
	Updated     int64  `json:"updated"       xorm:"updated"`
	Name        string `json:"name"          xorm:"name"`
	OwnerID     int64  `json:"owner_id"      xorm:"'owner_id'"`
	Token       string `json:"token"         xorm:"token"`
	LastContact int64  `json:"last_contact"  xorm:"last_contact"`
	Platform    string `json:"platform"      xorm:"VARCHAR(100) 'platform'"`
	Backend     string `json:"backend"       xorm:"VARCHAR(100) 'backend'"`
	Capacity    int32  `json:"capacity"      xorm:"capacity"`
	Version     string `json:"version"       xorm:"'version'"`
	NoSchedule  bool   `json:"no_schedule"   xorm:"no_schedule"`
	// Server side enforced agent filters
	Filters map[string]string `json:"filters" xorm:"'filters' json"`
} //	@name Agent

// TableName return database table name for xorm
func (Agent) TableName() string {
	return "agents"
}

func (a *Agent) IsSystemAgent() bool {
	return a.OwnerID == -1
}

var ErrFiltersBroken = errors.New("while creating filters map error ocured")

func (a *Agent) GetFilters() (map[string]string, error) {
	filters := a.Filters
	if filters == nil {
		filters = make(map[string]string)
	}

	// enforce filters for user and organization agents
	if a.IsSystemAgent() {
		filters["repo"] = "*"  // allow all repos by default
		filters["owner"] = "*" // allow all owners by default
	} else {
		filters["owner"] = "*" // we don't have org agents implemented yet
		// we expect this filter to be set else we fail
		if _, ok := filters["repo"]; !ok {
			return nil, ErrFiltersBroken
		}
	}

	return filters, nil
}
