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
	"fmt"
	"strings"
)

// Task defines scheduled pipeline Task.
type Task struct {
	ID           string                 `json:"id"           xorm:"PK UNIQUE 'id'"`
	Data         []byte                 `json:"-"            xorm:"LONGBLOB 'data'"`
	Labels       map[string]string      `json:"labels"       xorm:"json 'labels'"`
	Dependencies []string               `json:"dependencies" xorm:"json 'dependencies'"`
	RunOn        []string               `json:"run_on"       xorm:"json 'run_on'"`
	DepStatus    map[string]StatusValue `json:"dep_status"   xorm:"json 'dependencies_status'"`
	AgentID      int64                  `json:"agent_id"     xorm:"'agent_id'"`
} //	@name Task

// TableName return database table name for xorm.
func (Task) TableName() string {
	return "tasks"
}

func (t *Task) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%s) - %s", t.ID, t.Dependencies, t.DepStatus))
	return sb.String()
}

func (t *Task) ApplyLabelsFromRepo(r *Repo) error {
	if r == nil {
		return fmt.Errorf("repo is nil but needed to get task labels")
	}
	if t.Labels == nil {
		t.Labels = make(map[string]string)
	}
	t.Labels["repo"] = r.FullName
	t.Labels[agentFilterOrgID] = fmt.Sprintf("%d", r.OrgID)
	return nil
}

// ShouldRun tells if a task should be run or skipped, based on dependencies.
func (t *Task) ShouldRun() bool {
	if t.runsOnFailure() && t.runsOnSuccess() {
		return true
	}

	if !t.runsOnFailure() && t.runsOnSuccess() {
		for _, status := range t.DepStatus {
			if status != StatusSuccess {
				return false
			}
		}
		return true
	}

	if t.runsOnFailure() && !t.runsOnSuccess() {
		for _, status := range t.DepStatus {
			if status == StatusSuccess {
				return false
			}
		}
		return true
	}

	return false
}

func (t *Task) runsOnFailure() bool {
	for _, status := range t.RunOn {
		if status == string(StatusFailure) {
			return true
		}
	}
	return false
}

func (t *Task) runsOnSuccess() bool {
	if len(t.RunOn) == 0 {
		return true
	}

	for _, status := range t.RunOn {
		if status == string(StatusSuccess) {
			return true
		}
	}
	return false
}
