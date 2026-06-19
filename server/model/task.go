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
	"slices"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
)

// Task defines scheduled pipeline Task.
type Task struct {
	ID           string                 `json:"id"           xorm:"PK UNIQUE 'id'"`
	PID          int                    `json:"pid"          xorm:"'pid'"`
	Name         string                 `json:"name"         xorm:"'name'"`
	Data         []byte                 `json:"-"            xorm:"LONGBLOB 'data'"`
	Labels       map[string]string      `json:"labels"       xorm:"json 'labels'"`
	Dependencies []string               `json:"dependencies" xorm:"json 'dependencies'"`
	RunOn        []string               `json:"run_on"       xorm:"json 'run_on'"`
	DepStatus    map[string]StatusValue `json:"dep_status"   xorm:"json 'dependencies_status'"`
	AgentID      int64                  `json:"agent_id"     xorm:"'agent_id'"`
	PipelineID   int64                  `json:"pipeline_id"  xorm:"'pipeline_id'"`
	RepoID       int64                  `json:"repo_id"      xorm:"'repo_id'"`
	// ConcurrencyLimit is the maximum number of tasks sharing the same
	// ConcurrencyGroup that may run at once. A value <= 0 means unlimited.
	ConcurrencyLimit int `json:"concurrency_limit" xorm:"NOT NULL DEFAULT 0 'concurrency_limit'"`
	// ConcurrencyGroup identifies tasks that are limited against each other.
	// It is empty when no concurrency limit applies.
	ConcurrencyGroup string `json:"concurrency_group" xorm:"'concurrency_group'"`
	// Created is the unix timestamp the task's pipeline was created at. It
	// defines the queue ordering across pipelines.
	Created int64 `json:"created" xorm:"NOT NULL DEFAULT 0 'created'"`
} //	@name	Task

// TableName return database table name for xorm.
func (Task) TableName() string {
	return "tasks"
}

func (t *Task) String() string {
	return fmt.Sprintf("%s (%s) - %s", t.ID, t.Dependencies, t.DepStatus)
}

func (t *Task) ApplyLabelsFromRepo(r *Repo) error {
	if r == nil {
		return fmt.Errorf("repo is nil but needed to get task labels")
	}
	if t.Labels == nil {
		t.Labels = make(map[string]string)
	}
	t.Labels[pipeline.LabelFilterRepo] = r.FullName
	t.Labels[pipeline.LabelFilterOrg] = fmt.Sprintf("%d", r.OrgID)
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
	return slices.Contains(t.RunOn, string(StatusFailure))
}

func (t *Task) runsOnSuccess() bool {
	if len(t.RunOn) == 0 {
		return true
	}

	return slices.Contains(t.RunOn, string(StatusSuccess))
}
